package main

import (
	"crypto/tls"
	"io"
	"log"
	"net/http"

	"github.com/quic-go/quic-go/http3"
)

type RespBody struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// func main(){
// 	body := RespBody{
// 		Name: "Dharun",
// 		Age: 22,
// 	}
// 	transport := &http3.RoundTripper{
// 		TLSClientConfig: &tls.Config{
// 			InsecureSkipVerify: true,
// 			NextProtos: []string{"h3"},
// 		},
// 		QUICConfig: &quic.Config{Allow0RTT: true},
// 	}

// 	client := &http.Client{
// 		Transport: transport,
// 	}

// 	data, err := json.Marshal(body)
// 	if err != nil {
// 		log.Fatalf("Error while Marshalling struct %v",err)
// 	}

// 	resp, err := client.Post("https://localhost:5000/other","application/json",bytes.NewBuffer(data))
// 	if err !=nil {
// 		log.Fatalf("Error while Get Method %v",err)
// 	}

// 	defer resp.Body.Close()
// 	respBody, err := io.ReadAll(resp.Body)

// 	if err !=nil {
// 		log.Fatalf("Error while Reading Get response %v",err)
// 	}

// 	log.Printf("The response %s",string(respBody))
// }

func main() {
	transport := &http3.RoundTripper{
		TLSClientConfig: &tls.Config{
			ClientSessionCache: tls.NewLRUClientSessionCache(100),
			InsecureSkipVerify: true,
		},
	}
	req, err := http.NewRequest(http3.MethodGet0RTT, "https://127.0.0.1:5000/", nil)
	if err != nil {
		log.Fatalf("Error in get -> %v", err)
	}
	resp, err := transport.RoundTrip(req)
	if err != nil {
		log.Fatalf("Error in transport -> %v", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error in parsing response -> %v", err)
	}
	log.Printf("The response is %s", string(respBody))
}
