package packet

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type ConnectFlags struct {
	UserName     bool
	Passwore     bool
	WillRetain   bool
	WillQoS      byte // 2 bytes actually
	WillFlag     bool
	CleanSession bool
}

type VariableHeaderConnect struct {
	ProtocolName  string
	ProtocolLevel byte
	ConnectFlags  ConnectFlags
	KeepAlive     int
}

type ConnectControlPacket struct {
	FixedHeader    FixedHeader
	VariableHeader VariableHeaderConnect
	ConnectPayload ConnectPayload
}

func getConnectVariableHeader(r io.Reader) (hdr VariableHeaderConnect, len int, err error) {
	// Protocol name
	protocolName, n, err := getProtocolName(r)
	len += n
	if err != nil {
		return hdr, 0, err
	}
	hdr.ProtocolName = protocolName

	if hdr.ProtocolName != "MQTT" {
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
	hdr.ConnectFlags.Passwore = connectFlagsByte[0]&64 == 1
	hdr.ConnectFlags.WillRetain = connectFlagsByte[0]&32 == 1
	hdr.ConnectFlags.WillFlag = connectFlagsByte[0]&4 == 1
	hdr.ConnectFlags.CleanSession = connectFlagsByte[0]&2 == 1

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

	return
}
