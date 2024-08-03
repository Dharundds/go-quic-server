package main

import (
	"log"
	"net/http"

	"github.com/quic-go/quic-go/http3"
)

func main(){
	mux := http.NewServeMux()

	mux.HandleFunc("/",homeHandler)
	mux.HandleFunc("/other",otherHandler)
	mux.HandleFunc("/404", notFoundHandler)
	// mux.HandleFunc("/", notFoundHandler)

	log.Printf("Http/3 Server listening on %s", "localhost:5000")
	err := http3.ListenAndServeQUIC(
		":5000",
		"certificate.crt",
		"key.pem",
		mux,
	)


	if err != nil{
		log.Fatalf("Error in serving http3 server %v",err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request){
	if r.Method == http.MethodGet{
		log.Printf("Got GET request %v",r.Proto)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("GET request received"))
	} else{
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("/ only accepts GET"))
	}
}

func otherHandler(w http.ResponseWriter, r *http.Request){
	if r.Method == http.MethodPost{
		log.Printf("Got POST request %v",r)
		buf := make([]byte,1024)
		n, err:= r.Body.Read(buf)

		if err != nil {
			log.Fatalf("Error in Reading post body %v",err)
		}
		log.Default().Printf("Recived message %s",string(buf[:n]))
		w.Write([]byte("POST received your data"))
	} else{
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("/others only accepts POST"))
	}

}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusNotFound) // Set status code to 404 Not Found
    w.Write([]byte("404 - Page not found"))
}
