package storage

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

// func (dbio *TXPoolDBIO) LoadTxPoolFromDatabase(netService NetService, txPoolSize uint32) *transactionpool.TransactionPool {
// 	rawBytes, err := db.Get([]byte(TxPoolDbKey))
// 	if err != nil && err.Error() == storage.ErrKeyInvalid.Error() || len(rawBytes) == 0 {
// 		return NewTransactionPool(netService, txPoolSize)
// 	}
// 	txPool := deserializeTxPool(rawBytes)
// 	txPool.sizeLimit = txPoolSize
// 	txPool.netService = netService
// 	return txPool
// }
