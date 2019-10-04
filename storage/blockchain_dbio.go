package storage

import (
	"github.com/dappley/go-dappley/core/block"
	"github.com/dappley/go-dappley/util"
	logger "github.com/sirupsen/logrus"
)

type BlockChainDBIO struct {
	TxPoolDBIO    *TXPoolDBIO
	ScStateDBIO   *SCStateDBIO
	TxJournalBDIO *TXJournalDBIO
	UtxoDBIO      *UTXODBIO
	ChainDBIO     *ChainDBIO
	db            Storage
}

func NewBlockChainDBIO(db Storage) *BlockChainDBIO {
	blockChainDBIO := &BlockChainDBIO{}
	blockChainDBIO.TxPoolDBIO = NewTXPoolDBIO(db)
	blockChainDBIO.UtxoDBIO = NewUTXODBIO(db)
	blockChainDBIO.TxJournalBDIO = NewTXJournalDBIO(db)
	blockChainDBIO.ScStateDBIO = NewSCStateDBIO(db)
	blockChainDBIO.ChainDBIO = NewChainDBIO(db)
	blockChainDBIO.db = db
	return blockChainDBIO
}

func (dbio *BlockChainDBIO) EnableBatch() {
	dbio.db.EnableBatch()
}
func (dbio *BlockChainDBIO) DisableBatch() {
	dbio.db.DisableBatch()
}
func (dbio *BlockChainDBIO) Flush() error {
	return dbio.db.Flush()
}

func (dbio *BlockChainDBIO) Close() {
	dbio.db.Close()
}

func (dbio *BlockChainDBIO) AddBlockToDb(blk *block.Block) error {

	err := dbio.db.Put(blk.GetHash(), blk.Serialize())
	if err != nil {
		logger.WithError(err).Warn("Blockchain: failed to add blk to database!")
		return err
	}

	err = dbio.db.Put(util.UintToHex(blk.GetHeight()), blk.GetHash())
	if err != nil {
		logger.WithError(err).Warn("Blockchain: failed to index the blk by blk height in database!")
		return err
	}
	// add transaction journals
	for _, tx := range blk.GetTransactions() {
		err = dbio.TxJournalBDIO.PutTxJournal(*tx)
		if err != nil {
			logger.WithError(err).Warn("Blockchain: failed to add blk transaction journals into database!")
			return err
		}
	}
	return nil
}
