package socket

import (
	"math/rand"
	"time"
)

type Form map[string]string

type RandomString string

const (
	Alphabet RandomString = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	Numeric  RandomString = "0123456789"
)

func (c RandomString) Generate(n int) string {
	var r = rand.New(rand.NewSource(time.Now().UnixNano()))
	var b = make([]byte, n)
	var length = len(c)
	for i := 0; i < n; i++ {
		var idx = r.Intn(length)
		b[i] = c[idx]
	}
	return string(b)
}
