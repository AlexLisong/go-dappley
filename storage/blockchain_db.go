package storage

import (
	"errors"

	"github.com/dappley/go-dappley/common/hash"
	"github.com/dappley/go-dappley/util"
)

type BlockChainDB struct {
	db Storage
}

var (
	tipKey = []byte("tailBlockHash")
	libKey = []byte("lastIrreversibleBlockHash")
)

var (
	ErrBlockDoesNotExist = errors.New("block does not exist")
	ErrBlockHeightTooLow = errors.New("block height too low")
)

func NewBlockChainDB(db Storage) BlockChainDB {
	return BlockChainDB{db}
}

func (bcdb BlockChainDB) GetTips() ([]byte, error) {
	tip, err := bcdb.db.Get(tipKey)
	if err != nil {
		return nil, err
	}
	return tip, err
}

func (bcdb BlockChainDB) GetLib() ([]byte, error) {
	lib, err := bcdb.db.Get(libKey)
	if err != nil {
		return nil, err
	}
	return lib, err
}

func (bcdb BlockChainDB) GetBlockByHash(hash hash.Hash) ([]byte, error) {
	lib, err := bcdb.db.Get(hash)
	if err != nil {
		return nil, ErrBlockDoesNotExist
	}
	return lib, err
}

func (bcdb BlockChainDB) GetBlockByHeight(height uint64) ([]byte, error) {
	hash, err := bcdb.db.Get(util.UintToHex(height))
	if err != nil {
		return nil, ErrBlockDoesNotExist
	}
	return bcdb.GetBlockByHash(hash)
}
