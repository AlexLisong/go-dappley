package lblockchain

import (
	"errors"
	"github.com/jinzhu/copier"

	"github.com/dappley/go-dappley/common/hash"
	"github.com/dappley/go-dappley/core/block"
	"github.com/dappley/go-dappley/core/blockchain"
	"github.com/dappley/go-dappley/storage"
	lru "github.com/hashicorp/golang-lru"
	logger "github.com/sirupsen/logrus"
)

var (
	ErrBlockDoesNotExist = errors.New("block does not exist")
	ErrBlockHeightTooLow = errors.New("block height too low")
)

type Chain struct {
	bc        blockchain.Blockchain
	forks     *blockchain.BlockPool
	dbio      *storage.ChainDBIO
	LIBPolicy LIBPolicy
	forkHash  *lru.Cache //key: block height; value: block hash
}

func NewChain(genesis *block.Block, dbio *storage.ChainDBIO, policy LIBPolicy) *Chain {

	if genesis == nil {
		return nil
	}

	chain := &Chain{
		bc:        blockchain.NewBlockchain(genesis.GetHash(), genesis.GetHash()),
		forks:     blockchain.NewBlockPool(genesis),
		dbio:      dbio,
		LIBPolicy: policy,
	}

	chain.forkHash, _ = lru.New(policy.GetMinConfirmationNum() + 1)
	chain.SaveBlockToDb(genesis)
	return chain
}

func LoadBlockchainFromDb(dbio *storage.ChainDBIO, policy LIBPolicy) *Chain {
	bc, forks := dbio.LoadBlockchainFromDb()

	chain := &Chain{
		bc:        bc,
		forks:     forks,
		dbio:      dbio,
		LIBPolicy: policy,
	}
	chain.forkHash, _ = lru.New(policy.GetMinConfirmationNum() + 1)
	return chain
}

func (bc *Chain) GetTailBlockHash() hash.Hash {
	return bc.bc.GetTailBlockHash()
}

func (bc *Chain) GetLIBHash() hash.Hash {
	return bc.bc.GetLIBHash()
}

func (bc *Chain) SetTailBlockHash(hash hash.Hash) error {

	if !bc.IsInBlockchain(hash) {
		return ErrBlockDoesNotExist
	}

	err := bc.dbio.SetTailBlockHash(hash)
	if err != nil {
		return err
	}
	bc.bc.SetTailBlockHash(hash)
	return nil
}

func (bc *Chain) SetLIBHash(hash hash.Hash) error {

	newLIB, err := bc.GetBlockByHash(hash)

	if err != nil {
		return ErrBlockDoesNotExist
	}

	oldLIB, err := bc.GetLIB()
	if err != nil {
		logger.WithError(err).Panic("Blockchain: Failed during saving LIB")
	}

	if newLIB.GetHeight() <= oldLIB.GetHeight() {
		return ErrBlockHeightTooLow
	}

	currBlk := newLIB
	for !currBlk.GetHash().Equals(oldLIB.GetHash()) {
		bc.SaveBlockToDb(currBlk)

		currBlk, err = bc.GetBlockByHash(currBlk.GetPrevHash())
		if err != nil {
			logger.WithError(errors.New("Can not get block")).Panic("Blockchain: Failed during saving LIB")
		}
	}

	bc.forks.UpdateRootBlock(newLIB)

	err = bc.dbio.SetLIBHash(hash)
	if err != nil {
		logger.WithError(err).Panic("Blockchain: Filed to save LIB hash")
	}

	bc.bc.SetLIBHash(hash)

	return nil
}

func (bc *Chain) GetBlockByHash(hash hash.Hash) (*block.Block, error) {
	blk := bc.forks.GetBlockByHash(hash)
	if blk != nil {
		return blk, nil
	}

	return bc.dbio.GetBlockByHash(hash)

}

func (bc *Chain) GetTailBlock() (*block.Block, error) {
	return bc.GetBlockByHash(bc.GetTailBlockHash())
}

func (bc *Chain) GetLIB() (*block.Block, error) {
	return bc.GetBlockByHash(bc.GetLIBHash())
}

func (bc *Chain) GetMaxHeight() uint64 {
	blk, err := bc.GetTailBlock()
	if err != nil {
		return 0
	}
	return blk.GetHeight()
}

func (bc *Chain) GetLIBHeight() uint64 {
	blk, err := bc.GetLIB()
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
		blkHash, err = bc.dbio.GetBlockHashByHeight(height)
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

	//blockLogger := logger.WithFields(logger.Fields{
	//	"new_blk_height":           blk.GetHeight(),
	//	"new_blk_hash":             blk.GetHash().String(),
	//	"new_blk_num_of_tx":        len(blk.GetTransactions()),
	//	"lib_hash_before":          bc.GetLIBHash(),
	//	"lib_height_before":        bc.GetLIBHeight(),
	//	"tail_blk_hash_before":     bc.GetTailBlockHash(),
	//	"blockchain_height_before": bc.GetMaxHeight(),
	//})

	//oldTailBlkhash := bc.GetTailBlockHash()

	bc.forks.AddBlock(blk)

	//newTailBlk := bc.forks.GetHighestBlock()
	//if newTailBlk.GetHash().Equals(oldTailBlkhash) {
	//	blockLogger.Info("Chain: added a new block to forks")
	//	return nil
	//}
	//
	//bc.updateBlockchainInfo()
	//
	//blockLogger.WithFields(logger.Fields{
	//	"lib_hash_after":          bc.GetLIBHash(),
	//	"lib_height_after":        bc.GetLIBHeight(),
	//	"tail_blk_hash_after":     bc.GetTailBlockHash(),
	//	"blockchain_height_after": bc.GetMaxHeight(),
	//}).Info("Chain: added a new block to tail")
	return nil
}

func (bc *Chain) GetHighestBlock() *block.Block {
	return bc.forks.GetHighestBlock()
}

func (bc *Chain) updateBlockchainInfo() {
	bc.updateTailBlockInfo()
	bc.updateLIBInfo()
	bc.updateForkHeightCache()
}

func (bc *Chain) updateTailBlockInfo() {
	tailBlk := bc.forks.GetHighestBlock()
	err := bc.SetTailBlockHash(tailBlk.GetHash())
	if err != nil {
		logger.WithError(err).Panic("Blockchain: set tail block failed during tail block information update")
	}
}

func (bc *Chain) updateLIBInfo() {
	newLIBHeight := calculateLIBHeight(bc.GetMaxHeight(), bc.LIBPolicy.GetMinConfirmationNum())

	if newLIBHeight < bc.GetLIBHeight() {
		newLIBHeight = bc.GetLIBHeight()
	}

	currBlk, err := bc.GetTailBlock()
	if err != nil {
		logger.WithError(err).Panic("Blockchain: get tail block failed during LIB information update")
	}

	for currBlk.GetHeight() > newLIBHeight {

		prevHash := currBlk.GetPrevHash()

		if prevHash == nil {
			return
		}

		currBlk, err = bc.GetBlockByHash(currBlk.GetPrevHash())
		if err != nil {
			logger.WithError(err).Panic("Blockchain: get previous block failed during LIB information update")
		}
	}

	if currBlk.GetHash().Equals(bc.GetLIBHash()) {
		return
	}

	bc.saveLIB(currBlk)
}

func (bc *Chain) saveLIB(newLIB *block.Block) {
	oldLIB, err := bc.GetLIB()
	if err != nil {
		logger.WithError(err).Panic("Blockchain: Failed during saving LIB")
	}

	currBlk := newLIB
	for !currBlk.GetHash().Equals(oldLIB.GetHash()) {
		bc.SaveBlockToDb(currBlk)

		currBlk, err = bc.GetBlockByHash(currBlk.GetPrevHash())
		if err != nil {
			logger.WithError(errors.New("Can not get block")).Panic("Blockchain: Failed during saving LIB")
		}
	}

	bc.forks.UpdateRootBlock(newLIB)

	err = bc.SetLIBHash(newLIB.GetHash())
	if err != nil {
		logger.WithError(err).Panic("Blockchain: Failed during saving LIB")
	}
}

func (bc *Chain) updateForkHeightCache() {
	currBlk, err := bc.GetTailBlock()

	if err != nil {
		logger.WithError(err).Panic("Blockchain: get tail block failed during fork height cache update")
	}

	for currBlk.GetHeight() >= bc.GetLIBHeight() {
		bc.forkHash.Add(currBlk.GetHeight(), currBlk.GetHash())

		prevHash := currBlk.GetPrevHash()
		if prevHash == nil {
			return
		}

		currBlk, err = bc.GetBlockByHash(prevHash)
		if err != nil {
			logger.WithFields(logger.Fields{
				"prevHash": prevHash,
			}).WithError(err).Panic("Blockchain: get previous block failed during fork height cache update")
		}
	}
}

//SaveBlockToDb record the new block in the database
func (bc *Chain) SaveBlockToDb(blk *block.Block) error {
	return bc.dbio.AddBlockToDb(blk)
}

func (bc *Chain) IsBlockTooLow(blk *block.Block) bool {
	return blk.GetHeight() <= bc.GetLIBHeight()
}

func (bc *Chain) Iterator() *Chain {
	return &Chain{
		bc:    bc.bc,
		dbio:  bc.dbio,
		forks: bc.forks,
	}
}

func (bc *Chain) Next() (*block.Block, error) {
	var blk *block.Block

	blk, err := bc.GetTailBlock()
	if err != nil {
		return nil, err
	}

	bc.bc.SetTailBlockHash(blk.GetPrevHash())

	return blk, nil
}

func (bc *Chain) IsInBlockchain(hash hash.Hash) bool {
	_, err := bc.GetBlockByHash(hash)
	return err == nil
}

func (bc *Chain) GetNumOfForks() int64 {
	return bc.forks.GetNumOfForks()
}

func (bc *Chain) FindCommonAncestor(blk1 *block.Block, blk2 *block.Block) *block.Block {
	if !bc.IsInBlockchain(blk1.GetHash()) || blk1.GetHeight() < bc.GetLIBHeight() {
		return nil
	}

	if !bc.IsInBlockchain(blk2.GetHash()) || blk2.GetHeight() < bc.GetLIBHeight() {
		return nil
	}

	var err error
	fork1Hashes := make(map[string]bool)

	currBlk := blk1
	for currBlk.GetHeight() >= bc.GetLIBHeight() {
		fork1Hashes[currBlk.GetHash().String()] = true

		if currBlk.GetHeight() == 0 {
			break
		}
		currBlk, err = bc.GetBlockByHash(currBlk.GetPrevHash())
		if err != nil {
			return nil
		}
	}

	currBlk = blk2

	for currBlk.GetHeight() >= bc.GetLIBHeight() {

		if fork1Hashes[currBlk.GetHash().String()] {
			return currBlk
		}

		if currBlk.GetHeight() == 0 {
			return nil
		}

		currBlk, err = bc.GetBlockByHash(currBlk.GetPrevHash())
		if err != nil {
			return nil
		}
	}

	return nil
}

func (bc *Chain) DeepCopy() *Chain {
	newCopy := &Chain{}
	copier.Copy(newCopy, bc)
	return newCopy
}

func calculateLIBHeight(tailBlkHeight uint64, minConfirmationNum int) uint64 {
	LIBHeight := uint64(0)
	if tailBlkHeight > uint64(minConfirmationNum) {
		LIBHeight = tailBlkHeight - uint64(minConfirmationNum)
	}
	return LIBHeight
}
