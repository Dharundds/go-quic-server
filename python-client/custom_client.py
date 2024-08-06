import asyncio
from loguru import logger as log
import os
import pickle
import ssl
import time
import json
from collections import deque
from typing import AsyncIterator, Deque, Dict, Optional, Tuple, cast
from urllib.parse import urlparse

import httpx
from aioquic.asyncio.client import connect
from aioquic.asyncio.protocol import QuicConnectionProtocol
from aioquic.h3.connection import H3_ALPN, H3Connection
from aioquic.h3.events import DataReceived, H3Event, Headers, HeadersReceived
from aioquic.quic.configuration import QuicConfiguration
from aioquic.quic.events import QuicEvent
from aioquic.quic.packet import QuicProtocolVersion
from aioquic.quic.logger import QuicFileLogger
import uvloop


class H3ResponseStream(httpx.AsyncByteStream):
    def __init__(self, aiterator: AsyncIterator[bytes]):
        self._aiterator = aiterator

    async def __aiter__(self) -> AsyncIterator[bytes]:
        async for part in self._aiterator:
            yield part


class H3Transport(QuicConnectionProtocol, httpx.AsyncBaseTransport):
    def __init__(self, *args, **kwargs) -> None:
        super().__init__(*args, **kwargs)

        self._http = H3Connection(self._quic)
        self._read_queue: Dict[int, Deque[H3Event]] = {}
        self._read_ready: Dict[int, asyncio.Event] = {}

    async def handle_async_request(self, request: httpx.Request) -> httpx.Response:
        assert isinstance(request.stream, httpx.AsyncByteStream)

        stream_id = self._quic.get_next_available_stream_id()
        self._read_queue[stream_id] = deque()
        self._read_ready[stream_id] = asyncio.Event()

        # prepare request
        self._http.send_headers(
            stream_id=stream_id,
            headers=[
                (b":method", request.method.encode()),
                (b":scheme", request.url.raw_scheme),
                (b":authority", request.url.netloc),
                (b":path", request.url.raw_path),
            ]
            + [
                (k.lower(), v)
                for (k, v) in request.headers.raw
                if k.lower() not in (b"connection", b"host")
            ],
        )
        async for data in request.stream:
            self._http.send_data(stream_id=stream_id, data=data, end_stream=False)
        self._http.send_data(stream_id=stream_id, data=b"", end_stream=True)

        # transmit request
        self.transmit()

        # process response
        status_code, headers, stream_ended = await self._receive_response(stream_id)

        return httpx.Response(
            status_code=status_code,
            headers=headers,
            stream=H3ResponseStream(
                self._receive_response_data(stream_id, stream_ended)
            ),
            extensions={
                "http_version": b"HTTP/3",
            },
        )

    def http_event_received(self, event: H3Event):
        if isinstance(event, (HeadersReceived, DataReceived)):
            stream_id = event.stream_id
            if stream_id in self._read_queue:
                self._read_queue[event.stream_id].append(event)
                self._read_ready[event.stream_id].set()

    def quic_event_received(self, event: QuicEvent):
        #  pass event to the HTTP layer
        if self._http is not None:
            for http_event in self._http.handle_event(event):
                self.http_event_received(http_event)

    async def _receive_response(self, stream_id: int) -> Tuple[int, Headers, bool]:
        """
        Read the response status and headers.
        """
        stream_ended = False
        while True:
            event = await self._wait_for_http_event(stream_id)
            if isinstance(event, HeadersReceived):
                stream_ended = event.stream_ended
                break

        headers = []
        status_code = 0
        for header, value in event.headers:
            if header == b":status":
                status_code = int(value.decode())
            else:
                headers.append((header, value))
        return status_code, headers, stream_ended

    async def _receive_response_data(
        self, stream_id: int, stream_ended: bool
    ) -> AsyncIterator[bytes]:
        """
        Read the response data.
        """
        while not stream_ended:
            event = await self._wait_for_http_event(stream_id)
            if isinstance(event, DataReceived):
                stream_ended = event.stream_ended
                yield event.data
            elif isinstance(event, HeadersReceived):
                stream_ended = event.stream_ended

    async def _wait_for_http_event(self, stream_id: int) -> H3Event:
        """
        Returns the next HTTP/3 event for the given stream.
        """
        if not self._read_queue[stream_id]:
            await self._read_ready[stream_id].wait()
        event = self._read_queue[stream_id].popleft()
        if not self._read_queue[stream_id]:
            self._read_ready[stream_id].clear()
        return event


def save_session_ticket(ticket):
    """
    Callback which is invoked by the TLS engine when a new session ticket
    is received.
    """
    log.info("New session ticket received")
    # if args.session_ticket:
    with open("./session_ticket", "wb") as fp:
        pickle.dump(ticket, fp)


async def send_request(
    host: str,
    port: str,
    configuration: QuicConfiguration,
    data: Optional[str],
    zero_rtt: bool = True,
):
    async with connect(
        host,
        port,
        configuration=configuration,
        create_protocol=H3Transport,
        session_ticket_handler=save_session_ticket,
        wait_connected=not zero_rtt,
    ) as transport:

        log.info("Starting to establish QUIC connection")

        # log.info(f"Established QUIC connection with 0RTT {'Enabled' if not transport._connected else 'Disabled'}")
        async with httpx.AsyncClient(
            transport=cast(httpx.AsyncBaseTransport, transport)
        ) as client:
            # perform request
            start = time.time()
            if data is not None:
                response = await client.post(
                    url,
                    content=data.encode(),
                    headers={"content-type": "application/x-www-form-urlencoded"},
                )
            else:
                response = await client.get(url)

            elapsed = time.time() - start
            log.success(response)
        # print speed
        octets = len(response.content)
        log.info(
            "Received %d bytes in %.5f s (%.3f Mbps)"
            % (octets, elapsed, octets * 8 / elapsed / 1000000)
        )
        response_data = None
        try:
            response_data = json.loads(response.content)
            if response_data:
                log.info(f"Received Response data --> \n{json.dumps(response_data,indent=4)}")
        except Exception as e:
            pass


async def main( 
    configuration: QuicConfiguration,
    url: str,
    data: Optional[str],
    zero_rtt: bool = True,
    task_count: int = 1,
) -> None:
    # parse URL
    parsed = urlparse(url)
    assert parsed.scheme == "https", "Only https:// URLs are supported."
    host = parsed.hostname
    if parsed.port is not None:
        port = parsed.port
    else:
        port = 443

    await asyncio.wait(
        [
            asyncio.create_task(
                send_request(
                    host=host,
                    port=port,
                    configuration=configuration,
                    data=data,
                    zero_rtt=zero_rtt,
                )
            )
            for _ in range(task_count)
        ]
    )


if __name__ == "__main__":
    try:
        log.add("./client.log")
        defaults: Optional[QuicConfiguration] = QuicConfiguration(is_client=True)
        configuration: Optional[QuicConfiguration] = QuicConfiguration(
            is_client=True,
            alpn_protocols=H3_ALPN,
        )
        configuration.verify_mode = ssl.CERT_NONE
        configuration.load_verify_locations("/tests/pycacert.pem")
        configuration.original_version = QuicProtocolVersion.VERSION_2
        configuration.supported_versions = [
            QuicProtocolVersion.VERSION_2,
            QuicProtocolVersion.VERSION_1,
        ]
        try:
            with open("./session_ticket", "rb") as fp:
                configuration.session_ticket = pickle.load(fp)
                log.debug(f"Using existing session ticket for --> server_name : {configuration.session_ticket.server_name} , expiry_date : {configuration.session_ticket.not_valid_after}")
        except FileNotFoundError:
            pass


        url: str = "https://localhost:5000/tasks"
        zero_rtt: bool = True
        data = None
        # data = "Implement postgres"

        
        with asyncio.Runner(loop_factory=uvloop.new_event_loop) as runner:
            runner.run(
                main(
                    configuration=configuration,
                    url=url,
                    data=data,
                    zero_rtt=zero_rtt,
                )
            )
    except KeyboardInterrupt as e:
        runner.close()
        log.info("Shutting down")
