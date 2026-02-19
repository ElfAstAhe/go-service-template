package utils

import (
	"compress/flate"
	"compress/gzip"
	"compress/lzw"
	"errors"
	"fmt"
	"io"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/andybalholm/brotli"
)

// Decompressor encodings
const (
	EncodingGzip     = "gzip"     // gzip
	EncodingCompress = "compress" // lzw
	EncodingDeflate  = "deflate"  // zlib
	EncodingBrotli   = "br"       // brotli
)

// DecompressReader implementing ReadCLoser with source as compressed data reader
type DecompressReader struct {
	compressedSource io.ReadCloser
	decompressor     io.Reader
}

func NewDecompressReader(encoding string, compressedSource io.ReadCloser) (*DecompressReader, error) {
	dec, err := decompressorFactory(encoding, compressedSource)
	if err != nil {
		return nil, err
	}

	return &DecompressReader{
		compressedSource: compressedSource,
		decompressor:     dec,
	}, nil
}

func (dr *DecompressReader) Read(p []byte) (n int, err error) {
	cnt, err := dr.decompressor.Read(p)
	if err != nil && !errors.Is(err, io.EOF) {
		return cnt, errs.NewUtlError("DecompressReader.Read", "read error", err)
	}

	return cnt, err
}

func (dr *DecompressReader) Close() error {
	var err error
	if rc, ok := dr.decompressor.(io.Closer); ok {
		err = errors.Join(rc.Close(), dr.compressedSource.Close())
	} else {
		err = dr.compressedSource.Close()
	}
	if err != nil {
		return errs.NewUtlError("DecompressReader.Close", "close error", err)
	}

	return nil
}

func decompressorFactory(encoding string, compressedSource io.ReadCloser) (io.Reader, error) {
	var res io.Reader
	var err error
	switch encoding {
	case EncodingGzip:
		res, err = gzip.NewReader(compressedSource)
	//	case encodingDeflate:
	//		return zlib.NewReader(reader)
	case EncodingDeflate:
		res, err = flate.NewReader(compressedSource), nil
	case EncodingCompress:
		res, err = lzw.NewReader(compressedSource, lzw.LSB, 8), nil
	case EncodingBrotli:
		res, err = brotli.NewReader(compressedSource), nil
	default:
		res, err = nil, errs.NewMiddleWareError(fmt.Sprintf("unknown encoding [%s]", encoding), nil)
	}
	if err != nil {
		err = errs.NewUtlError("newDecompressor", "create decompressed reader", err)
	}

	return res, err
}
