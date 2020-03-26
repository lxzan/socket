package socket

import (
	"bytes"
	"encoding/binary"
	"errors"
	json "github.com/json-iterator/go"
	"net"
)

type Client struct {
	net.Conn
	readBufSize int64
	aes         *AesCrypto
	OnMessage   func(msg *Message)
	OnError     func(err error)
}

func (this *Client) handleMessage() {
	var buf = bytes.NewBufferString("")
	for {
		pack := make([]byte, this.readBufSize)
		_, err := this.Read(pack)
		if err != nil {
			this.OnError(ERR_ReadMessage.Wrap(err.Error()))
			return
		}

		pl := packLength(pack)
		buf.Write(pack[:pl])
		var rl uint32 = 4
		var rlb = true
		for uint32(buf.Len()) >= rl {
			var p = make([]byte, rl)
			_, err = buf.Read(p)
			if err != nil {
				this.OnError(ERR_ReadMessage.Wrap(err.Error()))
				return
			}

			if rlb {
				rl = binary.LittleEndian.Uint32(p)
				rlb = false
			} else {
				msg, err := this.decodeMessage(p)
				if err != nil {
					this.OnError(ERR_DecodeMessage.Wrap(err.Error()))
					return
				}
				this.OnMessage(msg)

				rl = 4
				rlb = true
			}
		}
	}
}

func (this *Client) decodeMessage(data []byte) (msg *Message, err error) {
	msg = &Message{}
	//var compressionAlgo = CompressionAlgo_Gzip
	var totalLength = len(data)
	//if totalLength < 5*1024 {
	//	compressionAlgo = CompressionAlgo_Gzip
	//}

	if totalLength < 4 {
		return nil, errors.New("receive invalid data")
	}

	var cryptoAlgo = CryptoAlgo(data[0])
	var compressionAlgo = CompressionAlgo(data[1])

	var b1 = data[2:4]
	var headerLength = binary.LittleEndian.Uint16(b1)
	//var rawHeader = data[4 : 4+headerLength]
	//var rawBody = data[4+headerLength:]

	var b2 []byte
	if compressionAlgo == CompressionAlgo_Gzip {
		tmp, err := GzipEncoder.Decode(data[4:])
		if err != nil {
			return nil, err
		}
		b2 = tmp
	}

	if err := json.Unmarshal(b2[4:4+headerLength], &msg.Header); err != nil {
		return nil, err
	}

	msg.Body = b2[4+headerLength:]

	return msg, nil
}
