package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"google.golang.org/grpc"
)

func createGrpcServer() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 8080))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	grpc.WithInsecure()
	srv := &hs.HubServiceService{}

	hs.RegisterHubServiceService(grpcServer, srv)
	grpcServer.Serve(lis)
	defer lis.Close()
}

func runWebserver() {
	cert, err := tls.LoadX509KeyPair("/home/pihpah/mycerts/sandbox.io.crt","/home/pihpah/mycerts/sandbox.io.key")
	if err != nil { panic(err) }
	//caCertPool := x509.NewCertPool()
	//var rootCert x509.Certificate
	rootPEM, _ := ioutil.ReadFile("/home/pihpah/mycerts/sandbox.io-rootCA.pem")

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(rootPEM))
	if !ok {
		panic("failed to parse root certificate")
	}

	if err != nil {
		panic(err)
	}
	//caCertPool.AddCert(rootCert)
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{ cert },
		RootCAs: roots,
	}
	server := &http.Server{
		Addr:              "sandbox.io:8080",
		Handler:           nil,
		TLSConfig:         tlsConfig,
		ReadTimeout:       0,
		ReadHeaderTimeout: 0,
		WriteTimeout:      0,
		IdleTimeout:       0,
		MaxHeaderBytes:    0,
		TLSNextProto:      nil,
		ConnState:         nil,
		ErrorLog:          nil,
		BaseContext:       nil,
		ConnContext:       nil,
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		//buf := make([]byte, 1024)
		//br.Read(buf)

		fmt.Printf("Body len: %d\n", len(body))

		w.WriteHeader()

		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})

	log.Fatal(server.ListenAndServeTLS("", ""))
}

func main() {
	runWebserver()
}
