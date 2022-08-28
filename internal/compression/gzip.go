package compression

import (
	"bytes"
	"compress/flate"
	"fmt"
)

func Compress(data []byte) ([]byte, error) {
	var valByte bytes.Buffer
	writer, err := flate.NewWriter(&valByte, flate.BestCompression)
	if err != nil {
		return nil, fmt.Errorf("failed init compress writer: %v", err)
	}
	_, err = writer.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed write data to compress temporary buffer: %v", err)
	}
	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("failed compress data: %v", err)
	}
	return valByte.Bytes(), nil
}

func Decompress(data []byte) ([]byte, error) {
	reader := flate.NewReader(bytes.NewReader(data))
	defer reader.Close()

	var valByte bytes.Buffer
	_, err := valByte.ReadFrom(reader)
	if err != nil {
		return nil, fmt.Errorf("failed decompress data: %v", err)
	}

	return valByte.Bytes(), nil
}
