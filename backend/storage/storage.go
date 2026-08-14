package storage

type Storage interface {
	Get(key string) ([]byte, error)
	Set(key string, data []byte) error
	Delete(key string) error
}
