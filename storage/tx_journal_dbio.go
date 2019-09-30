package storage

import (
	"github.com/dappley/go-dappley/core/transaction"
)

type TXJournalDBIO struct {
	db Storage
}

func NewTXJournalDBIO(db Storage) *TXJournalDBIO {
	return &TXJournalDBIO{db: db}
}

func (dbio *TXJournalDBIO) PutTxJournal(tx transaction.Transaction) error {
	txJournal := transaction.NewTxJournal(tx.ID, tx.Vout)

	bytes, err := txJournal.SerializeJournal()
	if err != nil {
		return err
	}
	err = dbio.db.Put(GetStorageKey(txJournal.Txid), bytes)
	return err
}
