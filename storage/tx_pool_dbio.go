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
