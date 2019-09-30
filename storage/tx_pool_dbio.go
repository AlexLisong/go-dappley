package storage

import "github.com/dappley/go-dappley/core/transactionpool"

const (
	NewTransactionTopic   = "NewTransaction"
	EvictTransactionTopic = "EvictTransaction"

	TxPoolDbKey = "txpool"

	BroadcastTx       = "BroadcastTx"
	BroadcastBatchTxs = "BraodcastBatchTxs"
)

type TXPoolDBIO struct {
	db Storage
}

func NewTXPoolDBIO(db Storage) *TXPoolDBIO {
	return &TXPoolDBIO{db: db}
}

func (dbio *TXPoolDBIO) SaveTxPool(txPool *transactionpool.TransactionPool) error {
	return dbio.db.Put([]byte(TxPoolDbKey), txPool.Serialize())
}

func (dbio *TXPoolDBIO) GetTxPool() *transactionpool.TransactionPool {
	rawBytes, err := dbio.db.Get([]byte(TxPoolDbKey))
	if err != nil && err.Error() == ErrKeyInvalid.Error() || len(rawBytes) == 0 {
		return transactionpool.NewTransactionPool()
	}
	txPool := transactionpool.DeserializeTxPool(rawBytes)
	return txPool
}
