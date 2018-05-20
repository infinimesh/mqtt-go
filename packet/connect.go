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

type ConnectControlPacket struct {
	FixedHeader    FixedHeader
	VariableHeader ConnectVariableHeader
	ConnectPayload ConnectPayload
}

type ConnAckControlPacket struct {
	FixedHeader    FixedHeader
	VariableHeader ConnAckVariableHeader
}

type ConnectVariableHeader struct {
	ProtocolName  string
	ProtocolLevel byte
	ConnectFlags  ConnectFlags
	KeepAlive     int
}

type ConnectPayload struct {
	ClientID string
}

type ConnAckVariableHeader struct {
	SessionPresent bool
	ReturnCode     byte
}

func SerializeFixedHeader(fh *FixedHeader, w io.Writer, remainingLength int) error {
	b := byte(fh.ControlPacketType) << 4

	// Flags must be < 16
	b |= fh.Flags

	_, err := w.Write([]byte{b})
	if err != nil {
		return err
	}

	_, err = serializeRemainingLength(w, remainingLength)
	return err

}

func SerializeConnAckControlPacket(connAck *ConnAckControlPacket, w io.Writer) error {
	if err := SerializeFixedHeader(&connAck.FixedHeader, w, 2 /* always 2 for ConnAck */); err != nil {
		return err
	}
	if err := SerializeConnAckVariableHeader(&connAck.VariableHeader, w); err != nil {
		return err
	}
	return nil
}

func SerializeConnAckVariableHeader(c *ConnAckVariableHeader, w io.Writer) error {
	buf := make([]byte, 2)
	connAckFlags := buf[0]
	connAckFlags |= 1

	buf[1] = c.ReturnCode

	n, err := w.Write(buf)
	if n != 2 {
		return errors.New("Failed to serialize variable header")
	}

	if err != nil {
		return errors.New("Failed to serialize variable header")
	}

	return nil
}

func getConnectVariableHeader(r io.Reader) (hdr ConnectVariableHeader, len int, err error) {
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
