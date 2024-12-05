package cacher

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"io"
)

type CompressAlg string

const (
	CompressAlgGzip  CompressAlg = "gzip"
	CompressAlgFlate CompressAlg = "flate"
	CompressAlgZlib  CompressAlg = "zlib"
)

func IsValidAlg(val CompressAlg) bool {
	return val == CompressAlgGzip || val == CompressAlgFlate || val == CompressAlgZlib
}

// Gzip
func CompressGzip(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	_, err := gzipWriter.Write(data)
	if err != nil {
		return nil, err
	}
	gzipWriter.Close() // Make sure to close to flush remaining data
	return buf.Bytes(), nil
}

func DecompressGzip(data []byte) ([]byte, error) {
	buf := bytes.NewReader(data)
	gzipReader, err := gzip.NewReader(buf)
	if err != nil {
		return nil, err
	}
	defer gzipReader.Close()

	var decompressed bytes.Buffer
	if _, err := io.Copy(&decompressed, gzipReader); err != nil {
		return nil, err
	}
	return decompressed.Bytes(), nil
}

// Zlib
func CompressZlib(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	zlibWriter := zlib.NewWriter(&buf)
	_, err := zlibWriter.Write(data)
	if err != nil {
		return nil, err
	}
	zlibWriter.Close() // Close to flush remaining data
	return buf.Bytes(), nil
}

func DecompressZlib(data []byte) ([]byte, error) {
	buf := bytes.NewReader(data)
	zlibReader, err := zlib.NewReader(buf)
	if err != nil {
		return nil, err
	}
	defer zlibReader.Close()

	var decompressed bytes.Buffer
	if _, err := io.Copy(&decompressed, zlibReader); err != nil {
		return nil, err
	}
	return decompressed.Bytes(), nil
}

// Flate

func CompressFlate(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	flateWriter, err := flate.NewWriter(&buf, flate.BestCompression)
	if err != nil {
		return nil, err
	}
	_, err = flateWriter.Write(data)
	if err != nil {
		return nil, err
	}
	flateWriter.Close()
	return buf.Bytes(), nil
}

func DecompressFlate(data []byte) ([]byte, error) {
	buf := bytes.NewReader(data)
	flateReader := flate.NewReader(buf)
	defer flateReader.Close()

	var decompressed bytes.Buffer
	if _, err := io.Copy(&decompressed, flateReader); err != nil {
		return nil, err
	}
	return decompressed.Bytes(), nil
}
