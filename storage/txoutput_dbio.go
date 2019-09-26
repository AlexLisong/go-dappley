package storage

import (
	"errors"

	"github.com/dappley/go-dappley/core/transaction"
	"github.com/dappley/go-dappley/core/transactionbase"
)

var (
	ErrUTXONotFound   = errors.New("utxo not found when trying to remove from cache")
	ErrTXInputInvalid = errors.New("txInput refers to non-existing transaction")
	ErrVoutNotFound   = errors.New("vout not found in current transaction")
)

type TXOutPutDBIO struct {
	db Storage
}

func NewTXOutPutDBIO(db Storage) *TXOutPutDBIO {
	return &TXOutPutDBIO{db: db}
}

func (dbio *TXOutPutDBIO) GetTxOutput(vin transactionbase.TXInput) (transactionbase.TXOutput, error) {
	key := GetStorageKey(vin.Txid)
	value, err := dbio.db.Get(key)
	if err != nil {
		return transactionbase.TXOutput{}, err
	}
	txJournal, err := transaction.DeserializeJournal(value)
	if err != nil {
		return transactionbase.TXOutput{}, err
	}
	if vin.Vout >= len(txJournal.Vout) {
		return transactionbase.TXOutput{}, ErrVoutNotFound
	}
	return txJournal.Vout[vin.Vout], nil
}

// generate storage key in database
func GetStorageKey(txid []byte) []byte {
	key := "tx_journal_" + string(txid)
	return []byte(key)
}
