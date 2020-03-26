package socket

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
)

type Encoder interface {
	Encode([]byte) ([]byte, error)
	Decode([]byte) ([]byte, error)
}

var (
	GzipEncoder = new(gzipEncoder)
)

type gzipEncoder struct{}

func (this *gzipEncoder) Encode(d []byte) ([]byte, error) {
	var buf = bytes.NewBufferString("")
	gzipWriter := gzip.NewWriter(buf)
	defer gzipWriter.Close()

	if _, err := gzipWriter.Write(d); err != nil {
		return nil, err
	}

	if err := gzipWriter.Flush(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (this *gzipEncoder) Decode(d []byte) ([]byte, error) {
	var r = bytes.NewReader(d)
	gzipReader, err := gzip.NewReader(r)
	defer gzipReader.Close()
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(gzipReader)
}
