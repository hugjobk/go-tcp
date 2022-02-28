package tcp

import (
	"encoding/binary"
	"fmt"
	"io"
)

const MagicNumber uint32 = 0x123456

type PacketWrapFunc func(data []byte) (packet []byte)

type PacketUnwrapFunc func(packet []byte) (data []byte)

type PacketSplitFunc func(data []byte, atEOF bool) (advance int, token []byte, err error)

type Packet struct {
	Data       []byte
	LocalAddr  string
	RemoteAddr string
}

func LengthValuePacketWrapFunc(magicNumber uint32) PacketWrapFunc {
	return func(data []byte) []byte {
		data = append(data, 0, 0, 0, 0, 0, 0)
		copy(data[6:], data)
		binary.BigEndian.PutUint32(data, magicNumber)
		binary.BigEndian.PutUint16(data[4:], uint16(len(data)-6))
		return data
	}
}

func LengthValuePacketUnwrapFunc(packet []byte) []byte {
	return packet[6:]
}

func LengthValuePacketSplitFunc(magicNumber uint32) PacketSplitFunc {
	return func(data []byte, atEOF bool) (int, []byte, error) {
		if len(data) > 6 {
			if n := binary.BigEndian.Uint32(data); n != magicNumber {
				return 0, nil, fmt.Errorf("invalid magic number: 0x%08x", n)
			}
			l := binary.BigEndian.Uint16(data[4:])
			pl := int(l) + 6
			if pl <= len(data) {
				return pl, data[:pl], nil
			}
		}
		if atEOF && len(data) > 0 {
			return 0, nil, io.ErrUnexpectedEOF
		}
		return 0, nil, nil
	}
}
