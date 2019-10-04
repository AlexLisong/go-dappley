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

func Save(dbio *storage.SCStateDBIO, blkHash hash.Hash, ss *scState.ScState) error {
	scStateOld := dbio.LoadScStateFromDatabase()
	change := scState.NewChangeLog()
	change.Log = findChangedValue(ss, scStateOld)

	err := dbio.SaveChangeLog(change, blkHash)
	if err != nil {
		return err
	}

	err = dbio.SaveToDatabase(ss)

	return err
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

func RevertState(dbio *storage.SCStateDBIO, prevHash hash.Hash, ss *scState.ScState) error {
	changelog := dbio.GetChangeLog(prevHash)
	if len(changelog) < 1 {
		return nil
	}
	revertState(changelog, ss)

	return nil
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
