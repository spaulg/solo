package domain

type LogWriter interface {
	Append(filePath string, data []byte) error
}
