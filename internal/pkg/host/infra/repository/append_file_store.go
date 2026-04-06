package repository

import (
	"fmt"
	"io"
	"os"
)

type AppendFileStore struct{}

func NewAppendFileStore() *AppendFileStore {
	return &AppendFileStore{}
}

func (t *AppendFileStore) Append(filePath string, data []byte) error {
	if len(data) == 0 {
		return nil
	}

	outputFile, err := os.OpenFile(filePath, os.O_SYNC|os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}

	defer func(outputFile *os.File) {
		_ = outputFile.Close()
	}(outputFile)

	n, err := outputFile.Write(data)
	if err != nil {
		return fmt.Errorf("failed to append to file %s: %w", filePath, err)
	}

	if n != len(data) {
		return fmt.Errorf("failed to append all data to file %s", filePath)
	}

	return nil
}

func (t *AppendFileStore) NewReader(filePath string) (io.ReadCloser, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s for reading: %w", filePath, err)
	}
	return f, nil
}
