package main

import (
	"crypto/tls"
	"io"
	"log"
	"net/http"

	"github.com/quic-go/quic-go/http3"
)


func main(){
	transport := &http3.RoundTripper{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	client := &http.Client{
		Transport: transport,
	}

	resp, err := client.Get("https://localhost:5000/")
	if err !=nil {
		log.Fatalf("Error while Get Method %v",err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if err !=nil {
		log.Fatalf("Error while Reading Get response %v",err)
	}

	log.Printf("The response %s",string(body))
}