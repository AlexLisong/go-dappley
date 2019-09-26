package storage

import (
	"github.com/dappley/go-dappley/core/scState"
)

const (
	scStateLogKey = "scLog"
	scStateMapKey = "scState"
)

type SCStateDBIO struct {
	db Storage
}

func NewSCStateDBIO(db Storage) *SCStateDBIO {
	return &SCStateDBIO{db: db}
}

//LoadScStateFromDatabase loads states from database
func (dbio *SCStateDBIO) LoadScStateFromDatabase() *scState.ScState {

	rawBytes, err := dbio.db.Get([]byte(scStateMapKey))

	if err != nil && err.Error() == ErrKeyInvalid.Error() || len(rawBytes) == 0 {
		return scState.NewScState()
	}
	return scState.DeserializeScState(rawBytes)
}

//SaveToDatabase saves states to database
func (dbio *SCStateDBIO) SaveToDatabase(ss *scState.ScState) error {
	return dbio.db.Put([]byte(scStateMapKey), ss.Serialize())
}
