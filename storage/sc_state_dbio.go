package storage

import (
	"github.com/dappley/go-dappley/common/hash"

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

func (dbio *SCStateDBIO) GetChangeLog(prevHash hash.Hash) map[string]map[string]string {
	change := make(map[string]map[string]string)

	rawBytes, err := dbio.db.Get([]byte(scStateLogKey + prevHash.String()))

	if err != nil && err.Error() == ErrKeyInvalid.Error() || len(rawBytes) == 0 {
		return change
	}
	change = scState.DeserializeChangeLog(rawBytes).Log

	return change
}

func (dbio *SCStateDBIO) DeleteChangeLog(prevHash hash.Hash) error {
	err := dbio.db.Del([]byte(scStateLogKey + prevHash.String()))
	return err
}

func (dbio *SCStateDBIO) SaveChangeLog(change *scState.ChangeLog, blkHash hash.Hash) error {
	err := dbio.db.Put([]byte(scStateLogKey+blkHash.String()), change.SerializeChangeLog())
	return err
}
