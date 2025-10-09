package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"os"
	"path"
	"path/filepath"
)

type JSONFileRepository[T any] struct{}

func NewJSONFileRepository[T any]() *JSONFileRepository[T] {
	return &JSONFileRepository[T]{}
}

func (t *JSONFileRepository[T]) Walk(filePath string, filename string) iter.Seq2[string, T] {
	return t.walk(filePath, filename, false)
}

func (t *JSONFileRepository[T]) ReverseWalk(filePath string, filename string) iter.Seq2[string, T] {
	return t.walk(filePath, filename, true)
}

func (t *JSONFileRepository[T]) Save(filePath string, entity T) error {
	data, err := json.MarshalIndent(entity, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal entity: %w", err)
	}

	if err := os.MkdirAll(path.Dir(filePath), 0700); err != nil {
		return fmt.Errorf("failed to create entity directory: %w", err)
	}

	tmpFilePath := path.Dir(filePath)
	tmpFile, err := os.CreateTemp(tmpFilePath, "*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}

	tmpName := tmpFile.Name()

	defer func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpName)
	}()

	n, err := tmpFile.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write entity file: %w", err)
	}

	if n != len(data) {
		return fmt.Errorf("failed to write all data to file %s", filePath)
	}

	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync entity file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close entity file: %w", err)
	}

	// nolint:gosec
	if err := os.Rename(tmpName, filePath); err != nil {
		return fmt.Errorf("failed to rename entity file %s: %w", filePath, err)
	}

	return nil
}

func (t *JSONFileRepository[T]) Load(filePath string) (T, error) {
	var entity T

	data, err := os.ReadFile(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return entity, nil
		}

		return entity, fmt.Errorf("failed to read json file %s: %w", filePath, err)
	}

	if len(data) > 0 {
		if err := json.Unmarshal(data, &entity); err != nil {
			return entity, fmt.Errorf("failed to unmarshal json: %w", err)
		}
	}

	return entity, nil
}

func (t *JSONFileRepository[T]) walk(filePath string, filename string, reverse bool) iter.Seq2[string, T] {
	return func(yield func(string, T) bool) {
		entries, err := os.ReadDir(filePath)
		if err != nil {
			return
		}

		if reverse {
			for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			executionEventPath := filepath.Join(filePath, entry.Name(), filename)
			eventData, err := t.Load(executionEventPath)
			if err != nil {
				return
			}

			yield(entry.Name(), eventData)
		}
	}
}
