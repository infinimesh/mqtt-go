package packet

import (
	"bytes"
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
	Flags             byte
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

// Return specific error, so server can answer with correct packet & error code (i.e. CONNACK with error 0x01)
func ReadPacket(r io.Reader) (ControlPacket, error) {
	fh, err := getFixedHeader(r)
	if err != nil {
		return nil, err
	}

	// Ensure that we always read the remaining bytes
	bufRemaining := make([]byte, fh.RemainingLength)
	n, err := io.ReadFull(r, bufRemaining)
	if n != fh.RemainingLength {
		return nil, errors.New("Short read!")
	}
	if err != nil {
		return nil, err
	}

	remainingReader := bytes.NewBuffer(bufRemaining)

	switch fh.ControlPacketType {
	case CONNECT:
		vh, variableHeaderSize, err := getConnectVariableHeader(remainingReader)
		if err != nil {
			return nil, err
		}
		payloadLength := fh.RemainingLength - variableHeaderSize

		cp, err := readConnectPayload(remainingReader, payloadLength)
		if err != nil {
			return nil, err
		}

		packet := &ConnectControlPacket{
			FixedHeader:    fh,
			VariableHeader: vh,
			ConnectPayload: cp,
		}

		return packet, nil
	case PUBLISH:
		fmt.Println("Received publish packet")
	case DISCONNECT:
		fmt.Println("Client disconnected")
	default:
		fmt.Println("IDK can't handle this", fh.ControlPacketType)
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

func serializeRemainingLength(w io.Writer, len int) (n int, err error) {
	stuffToWrite := make([]byte, 0)
	for {
		encodedByte := byte(len % 128)
		len = len / 128

		if len > 0 {
			encodedByte |= 128 //set topmost bit to true because we
			//still have stuff to write
			stuffToWrite = append(stuffToWrite, encodedByte)
		} else {
			stuffToWrite = append(stuffToWrite, encodedByte)
			break
		}
	}
	return w.Write(stuffToWrite)
}
