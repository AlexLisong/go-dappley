package storage

type UTXODB struct {
	db Storage
}

func NewUTXODB(db Storage) UTXODB {
	return UTXODB{db}
}

func (tdb UTXODB) Delete(pubKeyHash []byte) error {
	return tdb.db.Del(pubKeyHash)
}

func (tdb UTXODB) Save(key []byte, value []byte) error {
	return tdb.db.Put(key, value)
}

func (tdb UTXODB) Get(key []byte) ([]byte, error) {
	return tdb.db.Get(key)
}
