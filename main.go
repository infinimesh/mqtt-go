package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"

	"github.com/infinimesh/mqtt-go/packet"
)

//openssl req  -nodes -new -x509  -keyout server.key -out server.cert
func main() {
	pair, err := tls.LoadX509KeyPair("./server.cert", "./server.key")
	if err != nil {
		panic(err)
	}

	clientCAPool := x509.NewCertPool()

	crt, _ := ioutil.ReadFile("./server.cert")

	ok := clientCAPool.AppendCertsFromPEM(crt)
	if !ok {
		fmt.Println("failed to append cert from pem")
	}

	cfg := &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{pair},
		ClientCAs:    clientCAPool,
	}
	_, _ = tls.Listen("tcp", "localhost:8081", cfg)
	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}

	fmt.Println("waiting for conn")

	for {
		conn, _ := listener.Accept()
		fmt.Println("Accept")
		go handleConn(conn)
	}
}

func handleConn(c net.Conn) {
	for {
		// For the moment read only one
		p, err := packet.ReadPacket(c)
		if err != nil {
			fmt.Printf("Error while reading packet in client loop: %v", err)
			break
		}

		id := p.(*packet.ConnectControlPacket).ConnectPayload.ClientID
		fmt.Printf("Client with ID %v connected!", id)
	}
	fmt.Println("Exited loop of connection")
}
