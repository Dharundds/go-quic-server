package main

import (
	"io"
	"log"
	"net/http"
	"os"

	// "github.com/Dharundds/go-quic-server/helpers"
	"github.com/Dharundds/go-quic-server/helpers"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", homeHandler)
	mux.HandleFunc("POST /other", otherHandler)
	mux.HandleFunc("GET /404", notFoundHandler)
	// mux.HandleFunc("/", notFoundHandler)

	log.Printf("Http/3 Server listening on %s", "localhost:5000")
	certFile, keyFile := "certificate.crt", "key.pem"
	_, err := os.Stat(certFile)
	if os.IsNotExist(err) {
		helpers.GenCert(certFile, keyFile)
	}

	server := http3.Server{
		Addr:       ":5000",
		Handler:    mux,
		TLSConfig:  http3.ConfigureTLSConfig(helpers.GenerateTLSConfig(certFile, keyFile)),
		QUICConfig: &quic.Config{Allow0RTT: true},
	}
	// err := http3.ListenAndServeQUIC(
	// 	":5000",
	// 	"certificate.crt",
	// 	"key.pem",
	// 	mux,
	// )

	log.Fatal(server.ListenAndServe())

	// if err != nil{
	// 	log.Fatalf("Error in serving http3 server %v",err)
	// }
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Got GET request ->  HTTP Version : %v , 0RTT : %v , IPaddr: %v", r.Proto, !r.TLS.HandshakeComplete, r.RemoteAddr)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("GET request received"))

}

func otherHandler(w http.ResponseWriter, r *http.Request) {

	log.Printf("Got POST request %v", r.TLS.HandshakeComplete)
	reqBody, err := io.ReadAll(r.Body)

	if err != nil {
		log.Fatalf("Error while Reading Get response %v", err)
	}

	log.Printf("The response %s", string(reqBody))

	w.Write([]byte("POST received your data"))

}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound) // Set status code to 404 Not Found
	w.Write([]byte("404 - Page not found"))
}
