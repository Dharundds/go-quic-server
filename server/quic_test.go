package main

import (
	"context"
	"crypto/tls"
	"log"
	"os"
	"os/exec"

	"github.com/quic-go/quic-go"
)

const addr = "localhost:5000"
const message = "Hello from server"

func quic_test() {
	tlsConf := generateTLSConfig()
	if tlsConf == nil {
		log.Fatal("Error in listening at address, failed to read TLS config")
		return
	}

	listener, err := quic.ListenAddr(addr, tlsConf, nil)
	if err != nil {
		log.Fatalf("Error in listening at address: %v", err)
		return
	}

	log.Printf("QUIC Server listening on %s", addr)

	for {
		conn, err := listener.Accept(context.Background())
		if err != nil {
			log.Fatalf("Error in accepting connections: %v", err)
		}
		log.Println("Inside Accept")

		go handleConnSession(conn)
	}
}

func handleConnSession(session quic.Connection) {
	stream, err := session.AcceptStream(context.Background())
	if err != nil {
		log.Fatalf("Failed to accept stream: %v", err)
	}
	defer stream.Close()

	buf := make([]byte, 1024)
	n, err := stream.Read(buf)
	if err != nil {
		log.Fatalf("Failed to read from stream: %v", err)
	}

	log.Printf("Received message: %s", string(buf[:n]))
	copy(buf[:],[]byte(message))
	_, err = stream.Write(buf[:n])
	if err != nil {
		log.Fatalf("Failed to write from stream: %v", err)
	}
}

func generateTLSConfig() *tls.Config {
	certFile, keyFile := "certificate.crt", "key.pem"
	_, err := os.Stat(certFile)
	if os.IsNotExist(err) {
		genCert(certFile, keyFile)
	}
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatalf("Failed to load key pair: %v", err)
	}
	return &tls.Config{Certificates: []tls.Certificate{cert}}
}

func genCert(certFile, keyFile string) {
	cmd := exec.Command("openssl", "req", "-x509", "-newkey", "rsa:2048", "-keyout", keyFile, "-out", certFile, "-days", "365", "-nodes", "-subj", "/CN=localhost")
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Failed to generate certificate: %v", err)
	}
}
