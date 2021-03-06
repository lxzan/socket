package socket

import (
	"encoding/binary"
	"errors"
)

type MessageType uint8

const (
	TextMessage MessageType = iota
	BinaryMessage
	PingMessage
	PongMessage
	HandshakeMessage
)

type CryptoAlgo uint8

const (
	CryptoAlgo_NoCrypto CryptoAlgo = iota
	CryptoAlgo_RsaAes
)

type CompressAlgo uint8

const (
	CompressAlgo_NoCompress CompressAlgo = iota
	CompressAlgo_Gzip
	CompressAlgo_Flate
)

var (
	protocolMapping = map[byte]string{
		0: "1.0",
	}
)

const (
	currentProtocol = 0 // 1.0
)

type Message struct {
	Header *Header
	Body   []byte
}

type Header struct {
	ProtocolVersion   string
	MessageType       MessageType
	CompressAlgorithm CompressAlgo
	CryptoAlgorithm   CryptoAlgo
	HeaderLength      uint16
	Form              Form
}

func (this *Header) Get(k string) string {
	return this.Form[k]
}

func (this *Header) decodeProtocolHeader(d []byte) error {
	this.Form = Form{}
	var p1 = d[0]
	if protocol, ok := protocolMapping[p1]; ok {
		this.ProtocolVersion = protocol
	} else {
		return errors.New("unsupported protocol version")
	}

	this.MessageType = MessageType(d[1])
	this.CompressAlgorithm = CompressAlgo(d[2])
	this.CryptoAlgorithm = CryptoAlgo(d[3])
	this.HeaderLength = binary.LittleEndian.Uint16(d[4:6])

	return nil
}
