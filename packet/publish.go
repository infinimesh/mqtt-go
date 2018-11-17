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
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

type PublishControlPacket struct {
	FixedHeader      FixedHeader
	FixedHeaderFlags PublishHeaderFlags
	VariableHeader   PublishVariableHeader
	Payload          []byte
}

type PublishHeaderFlags struct {
	QoS    qosLevel
	Dup    bool
	Retain bool
}

type PublishVariableHeader struct {
	Topic    string
	PacketID int
}

func interpretPublishHeaderFlags(header byte) (flags PublishHeaderFlags, err error) {
	flags.Retain = header&1 > 0
	flags.Dup = header&8 > 0

	if header&2 > 0 && header&4 > 0 {
		err = errors.New("Both bits for QoS are set, this is invalid")
	}

	if header&2 > 0 {
		flags.QoS = QoSLevelAtLeastOnce
	} else if header&4 > 0 {
		flags.QoS = QoSLevelExactyleOnce
	} else {
		flags.QoS = QoSLevelNone
	}
	return
}

func readPublishVariableHeader(r io.Reader, flags PublishHeaderFlags) (vh PublishVariableHeader, len int, err error) {
	topicLength, err := readUint16(r)
	len += 2
	if err != nil {
		return
	}
	bufTopic := make([]byte, topicLength)
	n, err := io.ReadFull(r, bufTopic)
	len += n
	if err != nil {
		return
	}

	vh.Topic = string(bufTopic)

	if flags.QoS == QoSLevelAtLeastOnce || flags.QoS == QoSLevelExactyleOnce {
		vh.PacketID, err = readUint16(r)
		if err != nil {
			return
		}
		len += 2
	}

	return
}

func readPublishPayload(r io.Reader, len int) (buf []byte, err error) {
	buf = make([]byte, len)
	_, err = io.ReadFull(r, buf)
	return
}

func (p *PublishControlPacket) WriteTo(w io.Writer) (n int64, err error) {
	var nWritten int64

	// Calc Variable Header + Payload
	p.FixedHeader.RemainingLength = 2 + len(p.VariableHeader.Topic) + len(p.Payload)

	if p.FixedHeaderFlags.QoS == QoSLevelAtLeastOnce || p.FixedHeaderFlags.QoS == QoSLevelExactyleOnce {
		p.FixedHeader.RemainingLength += 2
	}

	nWritten, err = p.FixedHeader.WriteTo(w)
	n += nWritten
	if err != nil {
		return n, err
	}

	nWritten, err = p.VariableHeader.WriteTo(w)
	n += nWritten
	if err != nil {
		return n, err
	}

	nWritten, err = io.Copy(w, bytes.NewReader(p.Payload))
	n += nWritten
	if err != nil {
		return n, err
	}

	return n, err

}

func (c *PublishVariableHeader) WriteTo(w io.Writer) (n int64, err error) {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(len(c.Topic)))

	// TODO maybe write packetID if QoS... we dont do it for now

	return io.Copy(w, bytes.NewReader(b))
}
