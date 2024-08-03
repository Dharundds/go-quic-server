package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

func main(){
	tlsConfig := loadTlsCertificate()

	server := &http3.Server{
		Addr: "localhost:5000",
		TLSConfig: tlsConfig,
		Handler: http.HandlerFunc(httpHandler),
	}

	log.Printf("Http/3 Server listening on %s", "localhost:5000")

	err := server.ListenAndServeTLS("certificate.crt","key.pem")
	if err != nil{
		log.Fatalf("Error in serving http3 server %v",err)
	}
}
func loadTlsCertificate() *tls.Config {
	certFile,pemFile := "certificate.crt","key.pem"
	_, err:= os.Stat(certFile)
	if os.IsNotExist(err){
		log.Fatalf("Certificate file not found %v",err)
	}
	cert, err := tls.LoadX509KeyPair(certFile,pemFile)

	if err != nil{
		log.Fatalf("Error while loading certificates %v",err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
}


func httpHandler(w http.ResponseWriter, r *http.Request){

	if conn, ok:= r.Context().Value(http3.ServerContextKey).(quic.Connection); ok{
		log.Printf("Connection ID %s",conn.ConnectionState().Version)
	}else{
		log.Println()
	}


	switch r.Method{
	case http.MethodGet:
		handleGet(w,r)
	case http.MethodPost:
		handlePost(w,r)
	default:
		log.Fatal("Method not Allowed")
	}
}



func handleGet(w http.ResponseWriter, r *http.Request){
	log.Printf("Got GET request %v",r.Proto)
	w.Write([]byte("GET request received"))
}

func handlePost(w http.ResponseWriter, r *http.Request){
	log.Printf("Got POST request %v",r)
	buf := make([]byte,1024)
	n, err:= r.Body.Read(buf)

	if err != nil {
		log.Fatalf("Error in Reading post body %v",err)
	}
	log.Default().Printf("Recived message %s",string(buf[:n]))
	w.Write([]byte("POST received your data"))
}


