package socket

import (
	"bytes"
	"encoding/binary"
	"errors"
	json "github.com/json-iterator/go"
	"net"
	"strconv"
)

type Client struct {
	conn        net.Conn
	readBufSize int64
	aes         Encoder
	compression Encoder
	OnMessage   func(msg *Message)
	OnError     func(err error)
}

func (this *Client) HandleMessage() {
	var buf = bytes.NewBufferString("")
	for {
		pack := make([]byte, this.readBufSize)
		_, err := this.conn.Read(pack)
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
	msg = &Message{
		Header: Form{},
	}
	var totalLength = len(data)

	if totalLength < 4 {
		return nil, errors.New("receive invalid data")
	}

	var cryptoAlgo = CryptoAlgo(data[0])
	var compressionAlgo = CompressionAlgo(data[1])

	var b1 = data[2:4]
	var headerLength = binary.LittleEndian.Uint16(b1)

	var b2 []byte
	if compressionAlgo == CompressionAlgo_Gzip {
		tmp, err := GzipEncoder.Decode(data[4:])
		if err != nil {
			return nil, err
		}
		b2 = tmp
	} else {
		b2 = data[4:]
	}

	var b3 []byte
	if cryptoAlgo == CryptoAlgo_RsaAes {
		tmp, err := this.aes.Decode(b2)
		if err != nil {
			return nil, err
		}
		b3 = tmp
	} else {
		b3 = data[4:]
	}

	if err := json.Unmarshal(b3[:headerLength], &msg.Header); err != nil {
		return nil, err
	}

	msg.Body = b3[headerLength:]

	return msg, nil
}

func (this *Client) WriteMessage(typ MessageType, header Form, data []byte) (n int, err error) {
	if header == nil {
		header = Form{}
	}
	header["MessageType"] = strconv.Itoa(int(typ))

	var b0 = make([]byte, 4)
	var b1 = byte(0)
	var b2 = byte(0)
	var b3 = make([]byte, 2)
	var b4, _ = json.Marshal(header)
	var headerLength = len(b4)
	binary.LittleEndian.PutUint16(b3, uint16(headerLength))
	binary.LittleEndian.PutUint32(b0, uint32(4+len(b4)+len(data)))

	var buf = bytes.NewBuffer(b0)
	buf.Write([]byte{b1, b2})
	buf.Write(b3)
	buf.Write(b4)
	buf.Write(data)
	return this.conn.Write(buf.Bytes())
}
