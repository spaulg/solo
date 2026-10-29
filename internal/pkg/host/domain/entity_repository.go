package domain

import (
	"iter"
)

type EntityRepository[T any] interface {
	Walk(filePath string, filename string) iter.Seq2[string, T]
	ReverseWalk(filePath string, filename string) iter.Seq2[string, T]
	Save(filePath string, entity T) error
	Load(filePath string) (T, error)
}
