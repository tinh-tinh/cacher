package cacher

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"errors"
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

func Compress[M any](data M, alg CompressAlg) ([]byte, error) {
	input, err := ToBytes(data)
	if err != nil {
		return nil, err
	}
	switch alg {
	case CompressAlgGzip:
		return CompressGzip(input)
	case CompressAlgFlate:
		return CompressFlate(input)
	case CompressAlgZlib:
		return CompressZlib(input)
	default:
		return input, nil
	}
}

func Decompress[M any](input interface{}, alg CompressAlg) (M, error) {
	dataByte, ok := input.([]byte)
	if !ok {
		return *new(M), errors.New("assert type failed")
	}

	var output []byte
	var err error

	switch alg {
	case CompressAlgGzip:
		output, err = DecompressGzip(dataByte)
		if err != nil {
			return *new(M), err
		}
	case CompressAlgFlate:
		output, err = DecompressFlate(dataByte)
		if err != nil {
			return *new(M), err
		}
	case CompressAlgZlib:
		output, err = DecompressZlib(dataByte)
		if err != nil {
			return *new(M), err
		}
	default:
		return *new(M), err
	}

	dataRaw, err := FromBytes[M](output)
	if err != nil {
		return *new(M), err
	}

	data, ok := dataRaw.(M)
	if !ok {
		return *new(M), errors.New("assert type failed")
	}
	return data, nil
}
