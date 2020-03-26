package socket

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
