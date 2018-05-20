package main

import (
	"fmt"
	"net"

	"github.com/infinimesh/mqtt-go/packet"
)

//openssl req  -nodes -new -x509  -keyout server.key -out server.cert
func main() {
	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}

	for {
		conn, _ := listener.Accept()
		go handleConn(conn)
	}
}

func handleConn(c net.Conn) {
	p, err := packet.ReadPacket(c)
	if err != nil {
		fmt.Printf("Error while reading packet in client loop: %v\n", err)
		return
	}

	connectPacket, ok := p.(*packet.ConnectControlPacket)
	if !ok {
		fmt.Println("Got wrong packet as first packjet..need connect!")
		return
	}

	id := connectPacket.ConnectPayload.ClientID
	fmt.Printf("Client with ID %v connected!\n", id)

	resp := packet.ConnAckControlPacket{
		FixedHeader: packet.FixedHeader{
			ControlPacketType: packet.CONNACK,
		},
		VariableHeader: packet.ConnAckVariableHeader{},
	}

	packet.SerializeConnAckControlPacket(&resp, c)

	for {
		p, err := packet.ReadPacket(c)
		if err != nil {
			fmt.Printf("Error while reading packet in client loop: %v. Disconnecting client.\n", err)
			c.Close()
			break
		}

		switch p.(type) {
		case packet.ConnectControlPacket:
		}
	}
	fmt.Println("Exited loop of connection")
}
