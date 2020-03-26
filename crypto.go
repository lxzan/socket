package socket

import (
	"bytes"
	"crypto/aes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
)

type AesCrypto struct {
	key []byte
}

func NewAES(key []byte) *AesCrypto {
	return &AesCrypto{key: key,}
}

// ecb
func (this *AesCrypto) Encode(plainText []byte) (cryptText []byte, err error) {
	if len(this.key) != 16 && len(this.key) != 24 && len(this.key) != 32 {
		return nil, errors.New("ErrKeyLengthSixteen")
	}

	block, _ := aes.NewCipher(this.key)
	plainText = this.PKCS5Padding(plainText, block.BlockSize())
	decrypted := make([]byte, len(plainText))
	size := block.BlockSize()

	for bs, be := 0, size; bs < len(plainText); bs, be = bs+size, be+size {
		block.Encrypt(decrypted[bs:be], plainText[bs:be])
	}

	return decrypted, nil
}

func (this *AesCrypto) Decode(cryptText []byte) (plainText []byte, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()

	if len(this.key) != 16 && len(this.key) != 24 && len(this.key) != 32 {
		return nil, errors.New("ErrKeyLengthSixteen")
	}
	var length = len(cryptText)
	block, _ := aes.NewCipher(this.key)
	decrypted := make([]byte, len(cryptText))
	size := block.BlockSize()
	if size > length {
		return nil, errors.New("ErrData")
	}

	for bs, be := 0, size; bs < len(cryptText); bs, be = bs+size, be+size {
		block.Decrypt(decrypted[bs:be], cryptText[bs:be])
	}
	return this.PKCS5UnPadding(decrypted)
}

func (this *AesCrypto) PKCS5Padding(plainText []byte, blockSize int) []byte {
	padding := blockSize - (len(plainText) % blockSize)
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	newText := append(plainText, padText...)
	return newText
}

// 数组可能会越界
func (this *AesCrypto) PKCS5UnPadding(plainText []byte) ([]byte, error) {
	length := len(plainText)
	number := int(plainText[length-1])
	if number >= length {
		return nil, errors.New("ErrPaddingSize")
	}
	return plainText[:length-number], nil
}

type RsaCrypto struct {
	pub *pem.Block
	prv *pem.Block
}

// allow empty string
func NewRSA(pubPath string, prvPath string) (*RsaCrypto, error) {
	var o = &RsaCrypto{}
	if pubPath != "" {
		d, err := ioutil.ReadFile(pubPath)
		if err != nil {
			return nil, err
		}

		block, _ := pem.Decode(d)
		o.pub = block
	}

	if prvPath != "" {
		d, err := ioutil.ReadFile(prvPath)
		if err != nil {
			return nil, err
		}

		block, _ := pem.Decode(d)
		o.prv = block
	}

	return o, nil
}

func (this *RsaCrypto) Encode(plainText []byte) (cryptText []byte, err error) {
	pub, err := x509.ParsePKIXPublicKey(this.pub.Bytes)
	if err != nil {
		return nil, err
	}

	var publicKey = pub.(*rsa.PublicKey)
	cipherText, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, plainText)
	if err != nil {
		return nil, err
	}
	return cipherText, nil
}

func (this *RsaCrypto) Decode(cryptText []byte) (plainText []byte, err error) {
	privateKey, err := x509.ParsePKCS1PrivateKey(this.prv.Bytes)
	if err != nil {
		return []byte{}, err
	}
	plainText, err = rsa.DecryptPKCS1v15(rand.Reader, privateKey, cryptText)
	if err != nil {
		return []byte{}, err
	}
	return plainText, nil
}
