package storage

import (
	"errors"

	"github.com/dappley/go-dappley/common/hash"
	"github.com/dappley/go-dappley/core/block"
	"github.com/dappley/go-dappley/core/blockchain"
	"github.com/dappley/go-dappley/util"
	logger "github.com/sirupsen/logrus"
)

var (
	tipKey = []byte("tailBlockHash")
	libKey = []byte("lastIrreversibleBlockHash")
)

var (
	ErrBlockDoesNotExist = errors.New("block does not exist")
	ErrBlockHeightTooLow = errors.New("block height too low")
)

type ChainDBIO struct {
	db Storage
}

func (dbio *ChainDBIO) SetTailBlockHash(hash hash.Hash) error {
	err := dbio.db.Put(tipKey, hash)
	if err != nil {
		return err
	}
	return nil
}

func (dbio *ChainDBIO) SetLIBHash(hash hash.Hash) error {
	err := dbio.db.Put(libKey, hash)
	if err != nil {
		return err
	}
	return nil
}

func (dbio *ChainDBIO) GetBlockByHash(hash hash.Hash) (*block.Block, error) {
	rawBytes, err := dbio.db.Get(hash)
	if err != nil {
		return nil, ErrBlockDoesNotExist
	}
	return block.Deserialize(rawBytes), nil
}

func (dbio *ChainDBIO) GetBlockHashByHeight(height uint64) (hash.Hash, error) {

	var blkHash hash.Hash
	var err error

	blkHash, err = dbio.db.Get(util.UintToHex(height))
	if err != nil {
		return nil, ErrBlockDoesNotExist
	}

	return blkHash, nil
}

//AddBlockToDb record the new block in the database
func (dbio *ChainDBIO) AddBlockToDb(blk *block.Block) error {
	logger.WithFields(logger.Fields{
		"blkheight": blk.GetHeight(),
		"blkHash":   blk.GetHash(),
		"prevHash":  blk.GetPrevHash(),
	}).Debug("AddBlockToDb")

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

	return nil
}

func (dbio *ChainDBIO) LoadBlockchainFromDb() (blockchain.Blockchain, *blockchain.BlockPool) {

	LIBHash, err := dbio.db.Get(libKey)
	if err != nil {
		logger.WithError(err).Error("Blockchain: Getting last irreversible block hash failed during loading blockchain from database")
		return blockchain.Blockchain{}, nil
	}

	LIBRawBytes, err := dbio.db.Get(LIBHash)
	if err != nil {
		logger.WithError(err).Error("Blockchain: Getting last irreversible block failed during loading blockchain from database")
		return blockchain.Blockchain{}, nil
	}

	bc := blockchain.NewBlockchain(LIBHash, LIBHash)
	forks := blockchain.NewBlockPool(block.Deserialize(LIBRawBytes))

	return bc, forks
}
