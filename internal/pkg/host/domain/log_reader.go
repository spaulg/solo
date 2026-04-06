package domain

import "io"

type LogReader interface {
	NewReader(filePath string) (io.ReadCloser, error)
}
