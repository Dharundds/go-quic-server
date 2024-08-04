package main

import (
	"crypto/tls"
	"io"
	"log"
	"net/http"

	// "github.com/Dharundds/go-quic-server/helpers"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

func main(){
	mux := http.NewServeMux()

	mux.HandleFunc("GET /",homeHandler)
	mux.HandleFunc("POST /other",otherHandler)
	mux.HandleFunc("GET /404", notFoundHandler)
	// mux.HandleFunc("/", notFoundHandler)

	log.Printf("Http/3 Server listening on %s", "localhost:5000")
	server := http3.Server{
		Addr: "localhost:5000",
		Handler: mux,
		TLSConfig: &tls.Config{NextProtos: []string{"h3"}},
		QUICConfig: &quic.Config{Allow0RTT: true},

	}
	// err := http3.ListenAndServeQUIC(
	// 	":5000",
	// 	"certificate.crt",
	// 	"key.pem",
	// 	mux,
	// )
	log.Fatal(server.ListenAndServeTLS("certificate.crt","key.pem"))

	// if err != nil{
	// 	log.Fatalf("Error in serving http3 server %v",err)
	// }
}

func homeHandler(w http.ResponseWriter, r *http.Request){
	
	log.Printf("Got GET request %v %v",r.Proto, r.TLS.HandshakeComplete)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("GET request received"))
	
}

func otherHandler(w http.ResponseWriter, r *http.Request){
	
	log.Printf("Got POST request %v",r.TLS.HandshakeComplete)
	reqBody, err := io.ReadAll(r.Body)

	if err !=nil {
		log.Fatalf("Error while Reading Get response %v",err)
	}

	log.Printf("The response %s",string(reqBody))

	
	w.Write([]byte("POST received your data"))
	

}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusNotFound) // Set status code to 404 Not Found
    w.Write([]byte("404 - Page not found"))
}

