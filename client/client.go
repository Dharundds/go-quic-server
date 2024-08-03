package main

import (
	"context"
	"crypto/tls"
	"io"
	"log"

	"github.com/quic-go/quic-go"
)


const addr = "localhost:5000"
func client() {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
	}

	session, err := quic.DialAddr(context.TODO(),addr, tlsConf, nil)
	if err != nil {
		log.Fatalf("Failed to dial: %v", err)
	}
	defer session.CloseWithError(0, "bye")

	stream, err := session.OpenStreamSync(context.Background())
	if err != nil {
		log.Fatalf("Failed to open stream: %v", err)
	}
	defer stream.Close()

	message := "Hello from client!"
	_, err = stream.Write([]byte(message))
	if err != nil {
		log.Fatalf("Failed to write to stream: %v", err)
	}
	log.Printf("Sent: %s", message)

	reply := make([]byte, 1024)
	n, err := stream.Read(reply)
	if err != nil && err != io.EOF {
		log.Fatalf("Failed to read from stream: %v", err)
	}

	log.Printf("Received: %s", string(reply[:n]))
}
