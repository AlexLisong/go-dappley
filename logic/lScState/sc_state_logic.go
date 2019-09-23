package lScState

import (
	"github.com/dappley/go-dappley/common/hash"

	"github.com/dappley/go-dappley/core/scState"
	"github.com/dappley/go-dappley/storage"
)

const (
	scStateLogKey = "scLog"
	scStateMapKey = "scState"
)

//LoadScStateFromDatabase loads states from database
func LoadScStateFromDatabase(db storage.Storage) *scState.ScState {

	rawBytes, err := db.Get([]byte(scStateMapKey))

	if err != nil && err.Error() == storage.ErrKeyInvalid.Error() || len(rawBytes) == 0 {
		return scState.NewScState()
	}
	return scState.DeserializeScState(rawBytes)
}

func Save(db storage.Storage, blkHash hash.Hash, ss *scState.ScState) error {
	scStateOld := LoadScStateFromDatabase(db)
	change := scState.NewChangeLog()
	change.Log = findChangedValue(ss, scStateOld)

	err := db.Put([]byte(scStateLogKey+blkHash.String()), change.SerializeChangeLog())
	if err != nil {
		return err
	}

	err = db.Put([]byte(scStateMapKey), ss.Serialize())
	if err != nil {
		return err
	}

	return err
}

//SaveToDatabase saves states to database
func SaveToDatabase(db storage.Storage, ss *scState.ScState) error {
	return db.Put([]byte(scStateMapKey), ss.Serialize())
}

func findChangedValue(newScState *scState.ScState, oldScState *scState.ScState) map[string]map[string]string {
	change := make(map[string]map[string]string)

	newStates := newScState.GetStates()
	oldStates := oldScState.GetStates()
	for address, newMap := range newStates {
		if oldMap, ok := oldStates[address]; !ok {
			change[address] = nil
		} else {
			ls := make(map[string]string)
			for key, value := range oldMap {
				if newValue, ok := newMap[key]; ok {
					if newValue != value {
						ls[key] = value
					}
				} else {
					ls[key] = value
				}
			}

			for key, value := range newMap {
				if oldMap[key] != value {
					ls[key] = oldMap[key]
				}
			}

			if len(ls) > 0 {
				change[address] = ls
			}
		}
	}

	for address, oldMap := range oldStates {
		if _, ok := newStates[address]; !ok {
			change[address] = oldMap
		}
	}

	return change
}

func RevertState(db storage.Storage, prevHash hash.Hash, ss *scState.ScState) error {
	changelog := getChangeLog(db, prevHash)
	if len(changelog) < 1 {
		return nil
	}
	revertState(changelog, ss)
	//err := deleteLog(db, prevHash)
	//if err != nil {
	//	return err
	//}

	return nil
}

func getChangeLog(db storage.Storage, prevHash hash.Hash) map[string]map[string]string {
	change := make(map[string]map[string]string)

	rawBytes, err := db.Get([]byte(scStateLogKey + prevHash.String()))

	if err != nil && err.Error() == storage.ErrKeyInvalid.Error() || len(rawBytes) == 0 {
		return change
	}
	change = scState.DeserializeChangeLog(rawBytes).Log

	return change
}

func deleteLog(db storage.Storage, prevHash hash.Hash) error {
	err := db.Del([]byte(scStateLogKey + prevHash.String()))
	return err
}

func revertState(changelog map[string]map[string]string, ss *scState.ScState) {
	for address, pair := range changelog {
		state := ss.GetStates()
		if pair == nil {
			delete(state, address)
		} else {
			if _, ok := state[address]; !ok {
				state[address] = pair
			} else {
				for key, value := range pair {
					if value != "" {
						state[address][key] = value
					} else {
						delete(state[address], key)
					}
				}
			}
		}
	}
}
