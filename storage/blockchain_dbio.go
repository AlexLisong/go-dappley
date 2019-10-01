package storage

type BlockChainDBIO struct {
	txPoolDBIO    *TXPoolDBIO
	scStateDBIO   *SCStateDBIO
	txJournalBDIO *TXJournalDBIO
	utxoDBIO      *UTXODBIO
}

func NewBlockChainDBIO(db Storage) *BlockChainDBIO {
	blockChainDBIO := &BlockChainDBIO{}
	blockChainDBIO.txPoolDBIO = NewTXPoolDBIO(db)
	blockChainDBIO.utxoDBIO = NewUTXODBIO(db)
	blockChainDBIO.txJournalBDIO = NewTXJournalDBIO(db)
	blockChainDBIO.scStateDBIO = NewSCStateDBIO(db)
	return blockChainDBIO
}
