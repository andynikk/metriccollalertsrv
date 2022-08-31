package compression

import (
	"bytes"
	"compress/gzip"
	"fmt"
)

func Compress(data []byte) ([]byte, error) {
	var valByte bytes.Buffer
	writer := gzip.NewWriter(&valByte)
	if _, err := writer.Write(data); err != nil {
		return nil, fmt.Errorf("failed write data to compress temporary buffer: %v", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed compress data: %v", err)
	}
	return valByte.Bytes(), nil
}

func Decompress(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed init decompress reader: %v", err)
	}
	defer reader.Close()

	var valByte bytes.Buffer
	if _, err := valByte.ReadFrom(reader); err != nil {
		return nil, fmt.Errorf("failed decompress data: %v", err)
	}

	return valByte.Bytes(), nil
}
