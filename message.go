package socket

import (
	jsoniter "github.com/json-iterator/go"
	"strconv"
)

type MessageType uint8

const (
	TextMessage MessageType = iota
	BinaryMessage
	PingMessage
	PongMessage
	CloseMessage
)

type CryptoAlgo uint8

const (
	CryptoAlgo_NoCrypto CryptoAlgo = iota
	CryptoAlgo_RsaAes
)

type CompressionAlgo uint8

const (
	CompressionAlgo_NoCompression CompressionAlgo = iota
	CompressionAlgo_Gzip
)

type Message struct {
	Header map[string]string
	Body   []byte
}

type Header struct {
	CompressionAlgo
	CryptoAlgo
	form Form
}

func (this *Header) Get(k string) string {
	return this.form[k]
}

func decodeHeader(d []byte) (*Header, error) {
	var header = &Header{form: Form{}}
	if err := jsoniter.Unmarshal(d, &header.form); err != nil {
		return nil, err
	}

	if num, err := strconv.Atoi(header.form["CompressionAlgo"]); err != nil {
		return nil, err
	} else {
		header.CompressionAlgo = CompressionAlgo(num)
	}

	if num, err := strconv.Atoi(header.form["CryptoAlgo"]); err != nil {
		return nil, err
	} else {
		header.CryptoAlgo = CryptoAlgo(num)
	}
	return header, nil
}
