package socket

import "testing"

func TestCompress(t *testing.T) {
	var text = []byte("id，版本号，平台，下载地址，版本说明，是否强制更新，添加时间")
	d1, err := Compress(text)
	d2, err := UnCompress(d1)
	println(&d2, &err)
}

func BenchmarkCompress(b *testing.B) {
	var text = `
我想使用go从文件中读出一个块,将其视为一个字符串并gzip这个块.我知道如何从文件中读取并将其视为字符串,但是当涉及到“compress / gzip”时,我迷失了.我应该创建一个写入buf(字节片)的io.writer,使用gzip.Writer(io.writer)获取指向io.writer的编写器指针,然后使用gzip.Write(chunk_of_file)将chunk_of_file写入buf？然后我需要将字符串视为字节切片..
`
	var c = []byte(text)
	for i := 0; i < b.N; i++ {
		Compress(c)
	}
}
