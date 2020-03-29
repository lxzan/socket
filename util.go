package socket

import (
	"encoding/binary"
	"math/rand"
	"net"
	"time"
)

type Form map[string]string

type RandomString string

const (
	Alphabet RandomString = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	Numeric  RandomString = "0123456789"
)

func (c RandomString) Generate(n int, salt int64) string {
	var r = rand.New(rand.NewSource(time.Now().UnixNano() + salt))
	var b = make([]byte, n)
	var length = len(c)
	for i := 0; i < n; i++ {
		var idx = r.Intn(length)
		b[i] = c[idx]
	}
	return string(b)
}

func MTS() int64 {
	return time.Now().UnixNano() / 1000000
}

func MacAddrNumeric() (num uint32) {
	arr, err := net.InterfaceAddrs()
	if err != nil {
		return
	}

	for _, item := range arr {
		s := item.String()
		if s == "127.0.0.1" || s == "::1" {
			continue
		}

		if obj, ok := item.(*net.IPNet); ok {
			num = binary.LittleEndian.Uint32(obj.IP[12:])
			break
		}
	}
	return
}
