package utils

import (
	"compress/gzip"
	"compress/lzw"
	"compress/zlib"
	"errors"
	"fmt"
	"io"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/andybalholm/brotli"
)

// Поддерживаемые стандарты кодирования и сжатия данных (HTTP Content-Encoding / AMQP Headers).
const (
	EncodingGzip     = "gzip"     // Стандартный алгоритм сжатия GZIP (RFC 1952)
	EncodingCompress = "compress" // Алгоритм LZW (Lempel-Ziv-Welch), используемый в старых UNIX системах
	EncodingDeflate  = "deflate"  // Формат сжатия ZLIB (RFC 1950)
	EncodingBrotli   = "br"       // Современный высокопроизводительный алгоритм Brotli от Google
)

// DecompressReader реализует потоковый интерфейс io.ReadCloser (Адаптер / Wrapper).
// Прозрачно распаковывает входящий бинарный поток данных «на лету» на основе переданного типа кодирования.
type DecompressReader struct {
	compressedSource io.ReadCloser // Исходный сжатый сетевой поток или файл (например, body ответа)
	decompressor     io.Reader     // Конкретный бинарный распаковщик, инжектированный фабрикой
}

// Гарантируем сквозное соответствие интерфейсам ввода-вывода Go на этапе компиляции
var _ io.Reader = (*DecompressReader)(nil)
var _ io.Closer = (*DecompressReader)(nil)
var _ io.ReadCloser = (*DecompressReader)(nil)

// NewDecompressReader — фабричный конструктор декомпрессора.
// Принимает строковый маркер кодирования (например, "gzip" или "br") и оборачивает исходный поток.
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

// Read вычитывает распакованные байты в буфер p.
// Пропускает системный сигнал io.EOF без модификации, транслируя его наверх для корректного завершения циклов чтения.
func (dr *DecompressReader) Read(p []byte) (n int, err error) {
	cnt, err := dr.decompressor.Read(p)
	if err != nil && !errors.Is(err, io.EOF) {
		return cnt, errs.NewUtlError("DecompressReader.Read", "streaming decompression read failure", err)
	}

	return cnt, err
}

// Close осуществляет каскадное закрытие всех задействованных ресурсов (Ресурсный Менеджер).
// Гарантирует одновременное освобождение буферов памяти декомпрессора и закрытие базового сокета.
func (dr *DecompressReader) Close() error {
	var err error
	// Если внутренний декомпрессор библиотеки поддерживает io.Closer (например, gzip.Reader) — гасим оба
	if rc, ok := dr.decompressor.(io.Closer); ok {
		err = errors.Join(rc.Close(), dr.compressedSource.Close())
	} else {
		// Если декомпрессор не требует явного закрытия — закрываем только базовый источник
		err = dr.compressedSource.Close()
	}
	if err != nil {
		return errs.NewUtlError("DecompressReader.Close", "failed to cascade close decompression streams", err)
	}

	return nil
}

// decompressorFactory — классическая фабрика (Factory Method), инициализирующая
// нужный потоковый алгоритм распаковки на основе входящего заголовка.
func decompressorFactory(encoding string, compressedSource io.ReadCloser) (io.Reader, error) {
	var res io.Reader
	var err error

	switch encoding {
	case EncodingGzip:
		res, err = gzip.NewReader(compressedSource)
	case EncodingDeflate:
		res, err = zlib.NewReader(compressedSource)
	case EncodingCompress:
		// Алгоритм LZW инициализируется с порядком байт LSB (Least Significant Bit) и шириной кода 8 бит
		res, err = lzw.NewReader(compressedSource, lzw.LSB, 8), nil
	case EncodingBrotli:
		res, err = brotli.NewReader(compressedSource), nil
	default:
		res, err = nil, errs.NewUtlError("decompressorFactory", fmt.Sprintf("unsupported decompression encoding algorithm [%s]", encoding), nil)
	}
	if err != nil {
		err = errs.NewUtlError("decompressorFactory", "failed to initialize target decompression reader instance", err)
	}

	return res, err
}
