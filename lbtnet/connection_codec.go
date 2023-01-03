package lbtnet

import (
	"crypto/rc4"
	"io"
	"syscall"
	"time"

	"github.com/agility323/liberty/compress/zlib"
)

type EncryptReader struct {
	r io.Reader
	cipher *rc4.Cipher
}

func NewEncryptReader(r io.Reader, key []byte) (*EncryptReader, error) {
	cipher, err := rc4.NewCipher(key)
	if err != nil { return nil, err }
	return &EncryptReader{
		r: r,
		cipher: cipher,
	}, nil
}

func (er *EncryptReader) Read(buf []byte) (int, error) {
	n, err := er.r.Read(buf)
	if n > 0 { er.cipher.XORKeyStream(buf[:n], buf[:n]) }
	return n, err
}

func NewCompressReader(r io.Reader) (io.Reader, error) {
	r, err := zlib.NewReader(r)	// this call will be blocked at io.ReadFull
	if err != nil { return nil, err }
	return r, nil
}

type EncryptWriter struct {
	w io.Writer
	cipher *rc4.Cipher
}

func NewEncryptWriter(w io.Writer, key []byte) (*EncryptWriter, error) {
	cipher, err := rc4.NewCipher(key)
	if err != nil { return nil, err }
	return &EncryptWriter{
		w: w,
		cipher: cipher,
	}, nil
}

func (ew *EncryptWriter) Write(buf []byte) (int, error) {
	ew.cipher.XORKeyStream(buf, buf)
	bufSize := len(buf)
	writeSize := 0
	for {
		n, err := ew.w.Write(buf)
		writeSize += n
		if err == syscall.EAGAIN {
			time.Sleep(WriteWaitTime)
			continue
		}
		if err != nil {
			return writeSize, err
		}
		if n < len(buf) {
			buf = buf[n:]
		} else {
			break
		}
	}
	return bufSize, nil
}

type CompressWriter struct {
	w *zlib.Writer
}

func NewCompressWriter(w io.Writer) io.Writer {
	z := zlib.NewWriter(w)
	return &CompressWriter{
		w: z,
	}
}

func (cw *CompressWriter) Write(buf []byte) (int, error) {
	n, err := cw.w.Write(buf)
	cw.w.Flush()
	return n, err
}
