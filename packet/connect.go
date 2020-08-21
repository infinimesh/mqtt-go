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

package packet

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	RecieveMaximumID      = 33
	MaximumPacketSizeID   = 39
	TopicAliasMaximumID   = 34
	RequestResponseInfoID = 25
	RequestProblemInfoID  = 23
)

type ConnectProperties struct {
	PropertyLength         int //variable header properties length
	RecieveMaximumValue    int //limits the number of QoS 1 and QoS 2 Pub at Client - default 65,535
	MaximumPacketSize      int //represents max packet size client accepts
	TopicAliasMaximumValue int //max num of topic alias accepted by client
	RequestResponseInfo    int //0 = no response info in CONNACK
	RequestProblemInfo     int //0 = no reason string in CONNACK
}

type ConnectFlags struct {
	UserName   bool
	Password   bool
	WillRetain bool
	WillQoS    int // 2 bytes actually
	WillFlag   bool
	CleanStart bool
}

type ConnectControlPacket struct {
	FixedHeader    FixedHeader
	VariableHeader ConnectVariableHeader
	ConnectPayload ConnectPayload
}

type ConnectVariableHeader struct {
	ProtocolName      string
	ProtocolLevel     byte
	ConnectFlags      ConnectFlags
	KeepAlive         int
	ConnectProperties ConnectProperties
}

type ConnectPayload struct {
	ClientID string
}

func getConnectVariableHeader(r io.Reader, propertiesLength int) (hdr ConnectVariableHeader, len int, err error) {
	// Protocol name
	protocolName, n, err := getProtocolName(r)
	len += n
	if err != nil {
		return hdr, 0, err
	}
	hdr.ProtocolName = protocolName

	if hdr.ProtocolName != "MQTT" && hdr.ProtocolName != "MQIsdp" {
		return hdr, 0, fmt.Errorf("Invalid protocol: %v", hdr.ProtocolName)
	}

	// Get Proto level
	protocolLevelBytes := make([]byte, 1)
	n, err = r.Read(protocolLevelBytes)
	len += n
	if err != nil {
		return
	}
	hdr.ProtocolLevel = protocolLevelBytes[0]

	// Get Flags
	connectFlagsByte := make([]byte, 1)
	n, err = r.Read(connectFlagsByte)
	if n != 1 {
		return hdr, len, errors.New("Failed to read flags byte")
	}
	len += n
	if err != nil {
		return
	}

	hdr.ConnectFlags.UserName = connectFlagsByte[0]&128 == 1
	hdr.ConnectFlags.Password = connectFlagsByte[0]&64 == 1
	hdr.ConnectFlags.WillRetain = connectFlagsByte[0]&32 == 1
	hdr.ConnectFlags.WillFlag = connectFlagsByte[0]&4 == 1
	hdr.ConnectFlags.CleanStart = connectFlagsByte[0]&2 == 1

	keepAliveByte := make([]byte, 2)
	n, err = r.Read(keepAliveByte)
	len += n
	if err != nil {
		return hdr, len, errors.New("Could not read keepalive byte")
	}
	if n != 2 {
		return hdr, len, errors.New("Could not read enough keepalive bytes")
	}

	hdr.KeepAlive = int(binary.BigEndian.Uint16(keepAliveByte))

	// TODO Will QoS
	if connectFlagsByte[0]&16 == 1 && connectFlagsByte[0]&8 == 1 {
		hdr.ConnectFlags.WillQoS = 3
	} else if connectFlagsByte[0]&8 == 1 {
		hdr.ConnectFlags.WillQoS = 2
	} else {
		hdr.ConnectFlags.WillQoS = 1
	}

	bufProperties := make([]byte, propertiesLength)
	n, err = io.ReadFull(r, bufProperties)
	len += n
	if n != propertiesLength {
		return hdr, len, errors.New("short properties read")
	}
	if err != nil {
		return hdr, len, err
	}
	hdr.ConnectProperties.PropertyLength = propertiesLength

	return
}

func readConnectPayload(r io.Reader, len int) (ConnectPayload, error) {
	payloadBytes := make([]byte, len)
	n, err := io.ReadFull(r, payloadBytes)
	// TODO set upper limit for payload
	// TODO only stream it
	if err != nil {
		return ConnectPayload{}, err
	}
	if n != len {
		return ConnectPayload{}, errors.New("Payload length incorrect")
	}

	// CONNECT MUST have the client id
	// REGEX 0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ
	// MAY allow more than that, but this must be possible

	// Client Identifier, Will Topic, Will Message, User Name, Password

	// TODO am besten so viel einlesen wie moeglich, und dann reslicen / reader zusammenstecken

	clientIDLengthBytes := payloadBytes[:2]
	clientIDLength := binary.BigEndian.Uint16(clientIDLengthBytes)

	clientID := string(payloadBytes[2 : 2+clientIDLength])
	return ConnectPayload{
		ClientID: clientID,
	}, nil

}
