package domain

type EntityRepository[T any] interface {
	Save(filePath string, entity T) error
	Load(filePath string) (T, error)
}
