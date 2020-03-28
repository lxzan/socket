package socket

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"errors"
	"io/ioutil"
)

type Encoder interface {
	Encode([]byte) ([]byte, error)
	Decode([]byte) ([]byte, error)
}

var (
	encoderMapping = map[CompressAlgo]Encoder{
		CompressAlgo_Gzip:  new(GzipEncode),
		CompressAlgo_Flate: new(FlateEncode),
	}
)

type GzipEncode struct{}

func (this *GzipEncode) Encode(d []byte) ([]byte, error) {
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

func (this *GzipEncode) Decode(d []byte) ([]byte, error) {
	var r = bytes.NewReader(d)
	gzipReader, err := gzip.NewReader(r)
	defer func() {
		if gzipReader != nil {
			gzipReader.Close()
		}
	}()

	if err != nil {
		return nil, err
	}

	// TODO fix unexpected EOF
	result, err := ioutil.ReadAll(gzipReader)
	if err != nil && err.Error() == "unexpected EOF" {
		return result, nil
	}
	return result, err
}

type FlateEncode struct{}

func (this *FlateEncode) Encode(d []byte) ([]byte, error) {
	var buf = bytes.NewBufferString("")
	flateWriter, err := flate.NewWriter(buf, flate.DefaultCompression)
	if err != nil {
		return nil, err
	}

	defer flateWriter.Close()
	if _, err := flateWriter.Write(d); err != nil {
		return nil, err
	}

	if err := flateWriter.Flush(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (this *FlateEncode) Decode(d []byte) ([]byte, error) {
	var r = bytes.NewReader(d)
	flateReader := flate.NewReader(r)
	defer flateReader.Close()

	result, err := ioutil.ReadAll(flateReader)
	if err != nil && err.Error() == "unexpected EOF" {
		return result, nil
	}
	return result, err
}

func compress(algo CompressAlgo, d []byte) ([]byte, error) {
	if algo == CompressAlgo_NoCompress {
		return d, nil
	}

	obj, ok := encoderMapping[algo]
	if !ok {
		return nil, errors.New("unsupported compress algo")
	}
	return obj.Encode(d)
}

func uncompress(algo CompressAlgo, d []byte) ([]byte, error) {
	if algo == CompressAlgo_NoCompress {
		return d, nil
	}

	obj, ok := encoderMapping[algo]
	if !ok {
		return nil, errors.New("unsupported compress algo")
	}
	return obj.Decode(d)
}
