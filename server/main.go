package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"

	// "github.com/Dharundds/go-quic-server/helpers"
	"github.com/Dharundds/go-quic-server/helpers"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"

	"encoding/json"

	"github.com/jackc/pgx/v5"
)

var conn *pgx.Conn

func main() {
	var err error
	mux := http.NewServeMux()
	conn, err = pgx.Connect(context.Background(), "postgres://postgres:postgres@localhost:5432/todo")
	if err != nil {
		log.Fatalf("Unable to connection to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())
	mux.HandleFunc("GET /", homeHandler)
	mux.HandleFunc("GET /tasks", tasksHandler)
	mux.HandleFunc("POST /addTask", addTaskHandler)

	log.Printf("Http/3 Server listening on %s", "localhost:5000")
	certFile, keyFile := "certificate.crt", "key.pem"
	_, err = os.Stat(certFile)
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

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Got POST request ->  HTTP Version : %v , 0RTT : %v , IPaddr: %v", r.Proto, !r.TLS.HandshakeComplete, r.RemoteAddr)
	reqBody, err := io.ReadAll(r.Body)

	if err != nil {
		log.Fatalf("Error while Reading post body %v", err)
	}

	_, err = conn.Exec(context.Background(), "insert into tasks(description) values($1)", string(reqBody))
	if err != nil {
		log.Fatalf("Error while performing insert %v", err)
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("POST request received"))
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Got GET request ->  HTTP Version : %v , 0RTT : %v , IPaddr: %v", r.Proto, !r.TLS.HandshakeComplete, r.RemoteAddr)
	selectData, err := conn.Query(context.Background(), "select * from tasks")
	if err != nil {
		log.Fatalf("Error while performing insert %v", err)
	}
	retData := make(map[int]string)
	for selectData.Next() {
		var id int
		var description string
		err := selectData.Scan(&id, &description)
		if err != nil {
			log.Fatalf("Error while iterating rows %v\n", err)
		}
		retData[id] = description
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(retData); err != nil {
		log.Fatalf("Error while encoding json %v\n", err)
	}

}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Got GET request ->  HTTP Version : %v , 0RTT : %v , IPaddr: %v", r.Proto, !r.TLS.HandshakeComplete, r.RemoteAddr)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("GET request received"))

}

// func otherHandler(w http.ResponseWriter, r *http.Request) {

// 	log.Printf("Got POST request %v", r.TLS.HandshakeComplete)
// 	reqBody, err := io.ReadAll(r.Body)

// 	if err != nil {
// 		log.Fatalf("Error while Reading Get response %v", err)
// 	}

// 	log.Printf("The response %s", string(reqBody))

// 	w.Write([]byte("POST received your data"))

// }

// func notFoundHandler(w http.ResponseWriter, r *http.Request) {
// 	w.WriteHeader(http.StatusNotFound) // Set status code to 404 Not Found
// 	w.Write([]byte("404 - Page not found"))
// }
