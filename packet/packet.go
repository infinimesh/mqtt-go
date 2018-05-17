package packet

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type ControlPacketType byte

const (
	CONNECT     = 1
	CONNACK     = 2
	PUBLISH     = 3
	PUBACK      = 4
	PUBREC      = 5
	PUBREL      = 6
	PUBCOMP     = 7
	SUBSCRIBE   = 8
	SUBACK      = 9
	UNSUBSCRIBE = 10
	UNSUBACK    = 11
	PINGREQ     = 12
	PINGRESP    = 13
	DISCONNECT  = 14
)

// FixedHeader is contained in every packet (thus, fixed). It consists of the
// Packet Type, Packet-specific Flags and the length of the rest of the message.
type FixedHeader struct {
	ControlPacketType ControlPacketType
	RemainingLength   int
}

type ControlPacket interface {
}

func getProtocolName(r io.Reader) (protocolName string, len int, err error) {
	protocolNameLengthBytes := make([]byte, 2)
	n, err := r.Read(protocolNameLengthBytes)
	len += n
	if err != nil {
		return "", len, errors.New("Failed to read length of protocolNameLengthBytes")
	}
	if n != 2 {

		return "", len, errors.New("Failed to read length of protocolNameLengthBytes, not enough bytes")
	}

	protocolNameLength := binary.BigEndian.Uint16(protocolNameLengthBytes)

	protocolNameBuffer := make([]byte, protocolNameLength)
	n, err = r.Read(protocolNameBuffer) // use ReadFull, its not guaranteed that we get enough out of a single read
	len += n
	if err != nil {
		return "", len, err
	}
	if n != int(protocolNameLength) {
		return "", len, err
	}

	return string(protocolNameBuffer), len, nil
}

func getFixedHeader(r io.Reader) (fh FixedHeader, err error) {
	buf := make([]byte, 1)
	n, err := io.ReadFull(r, buf)
	if err != nil {
		return FixedHeader{}, err
	}
	if n != 1 {
		return FixedHeader{}, errors.New("Failed to read MQTT Packet Control Type from Client Stream")
	}
	fh.ControlPacketType = ControlPacketType(buf[0] >> 4)
	remainingLength, err := getRemainingLength(r) // Length VariableHeader + Payload
	if err != nil {
		return FixedHeader{}, err
	}
	fh.RemainingLength = remainingLength
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

// Return specific error, so server can answer with correct packet & error code (i.e. CONNACK with error 0x01)
func ReadPacket(r io.Reader) (ControlPacket, error) {
	fh, err := getFixedHeader(r)
	if err != nil {
		return nil, err
	}

	switch fh.ControlPacketType {
	case CONNECT:
		fmt.Println("Got CONNECT Message!")
		// TODO wie variable is variable header, kommts auf message type an?

		// The payload of the CONNECT Packet contains one or more length-prefixed fields, whose presence is determined by the flags in the variable header. These fields, if present, MUST appear in the order Client Identifier, Will Topic, Will Message, User Name, Password [MQTT-3.1.3-1].
		vh, variableHeaderSize, err := getConnectVariableHeader(r)
		if err != nil {
			return nil, err
		}
		payloadLength := fh.RemainingLength - variableHeaderSize

		cp, err := readConnectPayload(r, payloadLength)
		if err != nil {
			return nil, err
		}

		packet := &ConnectControlPacket{
			FixedHeader:    fh,
			VariableHeader: vh,
			ConnectPayload: cp,
		}

		return packet, nil

	default:
		fmt.Println("IDK can't handle this")
	}

	return nil, nil
}

// starts with variable header

// http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/os/mqtt-v3.1.1-os.html#_Toc398718023
func getRemainingLength(r io.Reader) (remaining int, err error) {
	// max 4 times / 4 rem. len.
	multiplier := 1
	for i := 0; i < 4; i++ {
		b := make([]byte, 1)
		n, err := r.Read(b)
		if err != nil {
			return 0, errors.New("Couldnt get remaning length")
		}
		if n != 1 {
			return 0, errors.New("Failed to get rem len")
		}

		valueThisTime := int(b[0] & 127)

		remaining += valueThisTime * multiplier
		multiplier *= 128

		moreBytes := b[0] & 128 // get only most significant bit
		if moreBytes == 0 {
			break
		}
	}
	return
}
