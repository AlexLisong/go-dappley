package lblockchain

import (
	"errors"
	"github.com/dappley/go-dappley/common/hash"
	"github.com/dappley/go-dappley/core/block"
	"github.com/dappley/go-dappley/core/blockchain"
	"github.com/dappley/go-dappley/util"
	"github.com/hashicorp/golang-lru"
	"github.com/sirupsen/logrus"
)

var (
	tipKey = []byte("tailBlockHash")
	libKey = []byte("lastIrreversibleBlockHash")
)

var (
	ErrBlockDoesNotExist = errors.New("block does not exist")
	ErrBlockHeightTooLow = errors.New("block height too low")
)

type Chain struct {
	bc        blockchain.Blockchain
	forks     *blockchain.BlockPool
	db        Storage
	LIBPolicy LIBPolicy
	forkHash  *lru.Cache //key: block height; value: block hash
}

func NewChain(tailBlk *block.Block, LIBBlk *block.Block, db Storage, policy LIBPolicy) *Chain {
	if tailBlk == nil {
		return nil
	}

	if LIBBlk == nil {
		return nil
	}

	chain := &Chain{
		bc:        blockchain.NewBlockchain(tailBlk.GetHash(), LIBBlk.GetHash()),
		forks:     blockchain.NewBlockPool(LIBBlk),
		db:        db,
		LIBPolicy: policy,
	}

	chain.forkHash, _ = lru.New(1000)
	return chain
}

func (bc *Chain) GetTailBlockHash() hash.Hash {
	return bc.bc.GetTailBlockHash()
}

func (bc *Chain) GetLIBBlockHash() hash.Hash {
	return bc.bc.GetLIBHash()
}

func (bc *Chain) SetTailBlockHash(hash hash.Hash) error {
	err := bc.db.Put(tipKey, hash)
	if err != nil {
		return err
	}
	bc.bc.SetTailBlockHash(hash)
	return nil
}

func (bc *Chain) SetLIBHash(hash hash.Hash) error {
	err := bc.db.Put(libKey, hash)
	if err != nil {
		return err
	}
	bc.bc.SetLIBHash(hash)
	return nil
}

func (bc *Chain) GetBlockByHash(hash hash.Hash) (*block.Block, error) {
	blk := bc.forks.GetBlockByHash(hash)
	if blk != nil {
		return blk, nil
	}

	rawBytes, err := bc.db.Get(hash)
	if err != nil {
		return nil, ErrBlockDoesNotExist
	}
	return block.Deserialize(rawBytes), nil

}

func (bc *Chain) GetTailBlock() (*block.Block, error) {
	return bc.GetBlockByHash(bc.GetTailBlockHash())
}

func (bc *Chain) GetLIBBlock() (*block.Block, error) {
	return bc.GetBlockByHash(bc.GetLIBBlockHash())
}

func (bc *Chain) GetMaxHeight() uint64 {
	blk, err := bc.GetTailBlock()
	if err != nil {
		return 0
	}
	return blk.GetHeight()
}

func (bc *Chain) GetLIBHeight() uint64 {
	blk, err := bc.GetLIBBlock()
	if err != nil {
		return 0
	}
	return blk.GetHeight()
}

func (bc *Chain) GetBlockByHeight(height uint64) (*block.Block, error) {

	var blkHash hash.Hash
	var err error

	if height >= bc.GetLIBHeight() {
		v, exists := bc.forkHash.Get(height)
		if !exists {
			return nil, ErrBlockDoesNotExist
		}
		blkHash = v.(hash.Hash)
	} else {
		blkHash, err = bc.db.Get(util.UintToHex(height))
		if err != nil {
			return nil, ErrBlockDoesNotExist
		}
	}

	return bc.GetBlockByHash(blkHash)
}

func (bc *Chain) AddBlock(blk *block.Block) error {
	if blk == nil {
		return ErrBlockDoesNotExist
	}

	if bc.IsBlockTooLow(blk) {
		return ErrBlockHeightTooLow
	}

	oldTailBlkhash := bc.GetTailBlockHash()

	bc.forks.AddBlock(blk)

	newTailBlk := bc.forks.GetHighestBlock()
	if newTailBlk.GetHash().Equals(oldTailBlkhash) {
		return nil
	}

	//TODO: Save LIB Block

	return bc.updateBlockchainInfo()
}

func (bc *Chain) updateBlockchainInfo() error {
	err := bc.updateTailBlockInfo()
	if err != nil {

		return err
	}
	err = bc.updateLIBBlockInfo()
	if err != nil {
		return err
	}
	logrus.Warn(3)

	return bc.updateForkHeightCache()
}

func (bc *Chain) updateTailBlockInfo() error {
	tailBlk := bc.forks.GetHighestBlock()
	return bc.SetTailBlockHash(tailBlk.GetHash())
}

func (bc *Chain) updateLIBBlockInfo() error {
	newLIBHeight := calculateLIBHeight(bc.GetMaxHeight(), bc.LIBPolicy.GetMinConfirmationNum())

	currBlk, err := bc.GetTailBlock()
	if err != nil {
		return err
	}

	for currBlk.GetHeight() > newLIBHeight {
		currBlk, err = bc.GetBlockByHash(currBlk.GetPrevHash())
		if err != nil {
			return err
		}
	}

	err = bc.SetLIBHash(currBlk.GetHash())
	if err != nil {
		return err
	}

	bc.forks.UpdateRootBlock(currBlk)
	return nil
}

func (bc *Chain) updateForkHeightCache() error {
	currBlk, err := bc.GetTailBlock()

	if err != nil {
		return err
	}

	for !currBlk.GetHash().Equals(bc.GetLIBBlockHash()) {
		bc.forkHash.Add(currBlk.GetHeight(), currBlk.GetHash())

		prevHash := currBlk.GetPrevHash()
		if prevHash == nil {
			return nil
		}
		//logrus.Info(currBlk.GetHash())

		currBlk, err = bc.GetBlockByHash(prevHash)
		if err != nil {

			return err
		}
	}
	return nil
}

func (bc *Chain) IsBlockTooLow(blk *block.Block) bool {
	return blk.GetHeight() <= bc.GetLIBHeight()
}

func calculateLIBHeight(tailBlkHeight uint64, minConfirmationNum int) uint64 {
	LIBHeight := uint64(0)
	if tailBlkHeight > uint64(minConfirmationNum) {
		LIBHeight = tailBlkHeight - uint64(minConfirmationNum)
	}
	return LIBHeight
}
