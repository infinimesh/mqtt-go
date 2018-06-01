//--------------------------------------------------------------------------
// Copyright 2018 infinimesh, INC
// www.infinimesh.io
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.
//--------------------------------------------------------------------------

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
	defer fmt.Println("Exited loop of connection")
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

	_, err = resp.WriteTo(c)
	if err != nil {
		fmt.Println("Failed to write ConnAck. Closing connection.")
		return
	}

	for {
		p, err := packet.ReadPacket(c)
		if err != nil {
			fmt.Printf("Error while reading packet in client loop: %v. Disconnecting client.\n", err)
			_ = c.Close()
			break
		}

		switch p := p.(type) {
		case packet.ConnectControlPacket:
		case *packet.PublishControlPacket:
			println("Received Publish with payload:", string(p.Payload))
		}
	}
}
