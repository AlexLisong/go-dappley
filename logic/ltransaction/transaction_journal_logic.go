package ltransaction

import (
	"github.com/dappley/go-dappley/core/transaction"
	"github.com/dappley/go-dappley/logic/lutxo"
	"github.com/dappley/go-dappley/storage"
)

// Add new log
func PutTxJournal(tx transaction.Transaction, db storage.Storage) error {
	txJournal := transaction.NewTxJournal(tx.ID, tx.Vout)
	return SaveTxJournal(db, txJournal)
}

// Save TxJournal into database
func SaveTxJournal(db storage.Storage, txJournal *transaction.TxJournal) error {
	bytes, err := txJournal.SerializeJournal()
	if err != nil {
		return err
	}
	err = db.Put(lutxo.GetStorageKey(txJournal.Txid), bytes)
	return err
}
