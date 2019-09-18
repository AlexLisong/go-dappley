package transaction

type Storage interface {
	Put(key []byte, val []byte) error
}
