// Copyright (C) 2018 go-dappley authors
//
// This file is part of the go-dappley library.
//
// the go-dappley library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either pubKeyHash 3 of the License, or
// (at your option) any later pubKeyHash.
//
// the go-dappley library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-dappley library.  If not, see <http://www.gnu.org/licenses/>.
//

package lblockchain

import (
	"bytes"
	"errors"
	"fmt"
	"sync"

	"github.com/dappley/go-dappley/core/scState"
	"github.com/dappley/go-dappley/core/transaction"
	"github.com/dappley/go-dappley/logic/lScState"
	"github.com/dappley/go-dappley/logic/ltransactionpool"
	"github.com/dappley/go-dappley/logic/lutxo"

	"github.com/dappley/go-dappley/common/hash"
	"github.com/dappley/go-dappley/core"
	"github.com/dappley/go-dappley/core/block"
	"github.com/dappley/go-dappley/core/blockchain"
	"github.com/dappley/go-dappley/logic/lblock"

	"github.com/dappley/go-dappley/core/account"
	"github.com/dappley/go-dappley/storage"
	"github.com/jinzhu/copier"
	logger "github.com/sirupsen/logrus"
)

var (
	ErrTransactionVerifyFailed = errors.New("transaction verification failed")
	ErrBlockExists             = errors.New("block already exists in blockchain")
	ErrAddBlockFailed          = errors.New("add block to blockchain failed")
	ErrProducerNotEnough       = errors.New("producer number is less than ConsensusSize")
	// DefaultGasPrice default price of per gas
	DefaultGasPrice uint64 = 1
)

type Blockchain struct {
	bc           *Chain
	dbio         *storage.BlockChainDBIO
	libPolicy    LIBPolicy
	txPool       *ltransactionpool.TransactionPoolLogic
	scManager    core.ScEngineManager
	eventManager *scState.EventManager
	blkSizeLimit int
	mutex        *sync.Mutex
}

// CreateBlockchain creates a new blockchain db
func CreateBlockchain(address account.Address, db storage.Storage, libPolicy LIBPolicy, txPool *ltransactionpool.TransactionPoolLogic, scManager core.ScEngineManager, blkSizeLimit int) *Blockchain {
	genesis := NewGenesisBlock(address, transaction.Subsidy)
	dbio := storage.NewBlockChainDBIO(db)
	bc := &Blockchain{
		NewChain(genesis, dbio.ChainDBIO, libPolicy),
		dbio,
		libPolicy,
		txPool,
		scManager,
		scState.NewEventManager(),
		blkSizeLimit,
		&sync.Mutex{},
	}
	utxoIndex := lutxo.NewUTXOIndex(bc.dbio.UtxoDBIO)
	utxoIndex.UpdateUtxoState(genesis.GetTransactions())
	utxoIndex.Save()
	return bc
}

func GetBlockchain(db storage.Storage, libPolicy LIBPolicy, txPool *ltransactionpool.TransactionPoolLogic, scManager core.ScEngineManager, blkSizeLimit int) (*Blockchain, error) {
	dbio := storage.NewBlockChainDBIO(db)
	bc := &Blockchain{
		LoadBlockchainFromDb(dbio.ChainDBIO, libPolicy),
		dbio,
		libPolicy,
		txPool,
		scManager,
		scState.NewEventManager(),
		blkSizeLimit,
		&sync.Mutex{},
	}
	return bc, nil
}

func (bc *Blockchain) GetDBIO() *storage.BlockChainDBIO {
	return bc.dbio
}

func (bc *Blockchain) GetTailBlockHash() hash.Hash {
	return bc.bc.GetTailBlockHash()
}

func (bc *Blockchain) GetLIBHash() hash.Hash {
	return bc.bc.GetLIBHash()
}

func (bc *Blockchain) GetSCManager() core.ScEngineManager {
	return bc.scManager
}

func (bc *Blockchain) GetTxPool() *ltransactionpool.TransactionPoolLogic {
	return bc.txPool
}

func (bc *Blockchain) GetEventManager() *scState.EventManager {
	return bc.eventManager
}

func (bc *Blockchain) SetBlockSizeLimit(limit int) {
	bc.blkSizeLimit = limit
}

func (bc *Blockchain) GetBlockSizeLimit() int {
	return bc.blkSizeLimit
}

func (bc *Blockchain) GetTailBlock() (*block.Block, error) {
	return bc.bc.GetTailBlock()
}

func (bc *Blockchain) GetLIB() (*block.Block, error) {
	return bc.bc.GetLIB()
}

func (bc *Blockchain) GetMaxHeight() uint64 {
	return bc.bc.GetMaxHeight()
}

func (bc *Blockchain) GetLIBHeight() uint64 {
	return bc.bc.GetLIBHeight()
}

func (bc *Blockchain) GetBlockByHash(hash hash.Hash) (*block.Block, error) {
	return bc.bc.GetBlockByHash(hash)
}

func (bc *Blockchain) GetBlockByHeight(height uint64) (*block.Block, error) {
	return bc.bc.GetBlockByHeight(height)
}

func (bc *Blockchain) SetTailBlockHash(tailBlockHash hash.Hash) {
	bc.bc.SetTailBlockHash(tailBlockHash)
}

func (bc *Blockchain) SetState(state blockchain.BlockchainState) {
	bc.bc.bc.SetState(state)
}

func (bc *Blockchain) GetState() blockchain.BlockchainState {
	return bc.bc.bc.GetState()
}

func (bc *Blockchain) MergeFork(forkBlks []*block.Block, forkParentHash hash.Hash) error {

	//find parent block
	if len(forkBlks) == 0 {
		return nil
	}
	forkHeadBlock := forkBlks[len(forkBlks)-1]
	if forkHeadBlock == nil {
		return nil
	}

	//verify transactions in the fork
	utxo, scState, err := RevertUtxoAndScStateAtBlockHash(bc, forkParentHash)
	if err != nil {
		logger.Error("BlockchainManager: blockchain is corrupted! Delete the database file and resynchronize to the network.")
		return err
	}
	rollBackUtxo := utxo.DeepCopy()
	rollScState := scState.DeepCopy()

	parentBlk, err := bc.GetBlockByHash(forkParentHash)
	if err != nil {
		logger.WithFields(logger.Fields{
			"error": err,
			"hash":  forkParentHash.String(),
		}).Error("BlockchainManager: get fork parent block failed.")
	}

	firstCheck := true

	for i := len(forkBlks) - 1; i >= 0; i-- {
		logger.WithFields(logger.Fields{
			"height": forkBlks[i].GetHeight(),
			"hash":   forkBlks[i].GetHash().String(),
		}).Debug("BlockchainManager: is verifying a block in the fork.")

		if !lblock.VerifyTransactions(forkBlks[i], utxo, scState, parentBlk) {
			return ErrTransactionVerifyFailed
		}

		if !bc.CheckLibPolicy(forkBlks[i]) {
			return ErrProducerNotEnough
		}

		if firstCheck {
			firstCheck = false
			bc.Rollback(forkParentHash, rollBackUtxo, rollScState)
		}

		ctx := BlockContext{Block: forkBlks[i], UtxoIndex: utxo, State: scState}
		err = bc.AddBlockWithContext(&ctx)
		if err != nil {
			logger.WithFields(logger.Fields{
				"error":  err,
				"height": forkBlks[i].GetHeight(),
			}).Error("BlockchainManager: add fork to tail failed.")
		}
		parentBlk = forkBlks[i]
	}

	return nil
}

func (bc *Blockchain) AddBlockWithContext(ctx *BlockContext) error {
	// Atomically set tail block hash and update UTXO index in db
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	if bc.bc.IsInBlockchain(ctx.Block.GetHash()) {
		return ErrBlockExists
	}

	blockLogger := logger.WithFields(logger.Fields{
		"num_txs_in_txpol_before": bc.GetTxPool().GetNumOfTxInPool(),
	})

	bcTemp := bc.DeepCopy()
	tailBlk, _ := bc.GetTailBlock()
	bcTemp.runScheduleEvents(ctx, tailBlk)

	bcTemp.dbio.EnableBatch()
	defer bcTemp.dbio.DisableBatch()

	//update Blockchain
	err := bc.bc.AddBlock(ctx.Block)
	if err != nil {
		blockLogger.WithError(err).Warn("Blockchain: Add block failed")
		return ErrAddBlockFailed
	}

	//update UTXOIndex
	err = ctx.UtxoIndex.Save()
	if err != nil {
		blockLogger.Warn("Blockchain: failed to save utxo to database.")
		return err
	}

	//update Transaction pool
	bcTemp.GetTxPool().CleanUpMinedTxs(ctx.Block.GetTransactions())
	bcTemp.GetTxPool().ResetPendingTransactions()
	err = bcTemp.GetTxPool().SaveToDatabase(bc.dbio.TxPoolDBIO)

	if err != nil {
		blockLogger.Warn("Blockchain: failed to save txpool to database.")
		return err
	}

	//update smart contract state
	err = lScState.Save(bc.dbio.ScStateDBIO, ctx.Block.GetHash(), ctx.State)
	if err != nil {
		blockLogger.Warn("Blockchain: failed to save scState to database.")
		return err
	}

	bc.updateTransactionJournal(tailBlk)

	// Flush batch changes to storage
	err = bcTemp.dbio.Flush()
	if err != nil {
		blockLogger.Error("Blockchain: failed to update tail block hash and UTXO index!")
		return err
	}

	// Assign changes to receiver
	*bc = *bcTemp

	blockLogger.WithFields(logger.Fields{
		"num_txs_in_txpol_after": bc.GetTxPool().GetNumOfTxInPool(),
	}).Info("Blockchain: added a new block")

	return nil
}

func (bc *Blockchain) updateTransactionJournal(oldTailBlk *block.Block) error {

	var err error
	currBlk := oldTailBlk
	for currBlk.GetHeight() <= bc.GetMaxHeight() {
		// add transaction journals
		currBlk, err = bc.GetBlockByHeight(currBlk.GetHeight() + 1)
		if err != nil {
			return err
		}

		for _, tx := range currBlk.GetTransactions() {
			err := bc.dbio.TxJournalBDIO.PutTxJournal(*tx)

			if err != nil {
				logger.WithError(err).Warn("Blockchain: failed to add blk transaction journals into database!")
				return err
			}
		}

	}

	return nil
}

func (bc *Blockchain) runScheduleEvents(ctx *BlockContext, parentBlk *block.Block) error {
	if parentBlk == nil {
		//if the current block is genesis block. do not run smart contract
		return nil
	}

	if bc.scManager == nil {
		return nil
	}

	bc.scManager.RunScheduledEvents(ctx.UtxoIndex.GetContractUtxos(), ctx.State, ctx.Block.GetHeight(), parentBlk.GetTimestamp())
	bc.eventManager.Trigger(ctx.State.GetEvents())
	return nil
}

func (bc *Blockchain) Iterator() *Chain {
	return bc.bc.Iterator()
}

func (bc *Blockchain) Next() (*block.Block, error) {
	return bc.bc.Next()
}

func (bc *Blockchain) String() string {
	var buffer bytes.Buffer

	bci := bc.Iterator()
	for {
		block, err := bci.Next()
		if err != nil {
			logger.Error(err)
		}

		buffer.WriteString(fmt.Sprintf("============ Block %x ============\n", block.GetHash()))
		buffer.WriteString(fmt.Sprintf("Height: %d\n", block.GetHeight()))
		buffer.WriteString(fmt.Sprintf("Prev. block: %x\n", block.GetPrevHash()))
		for _, tx := range block.GetTransactions() {
			buffer.WriteString(tx.String())
		}
		buffer.WriteString(fmt.Sprintf("\n\n"))

		if len(block.GetPrevHash()) == 0 {
			break
		}
	}
	return buffer.String()
}

func (bc *Blockchain) IsInBlockchain(hash hash.Hash) bool {
	return bc.bc.IsInBlockchain(hash)
}

//rollback the blockchain to a block with the targetHash
func (bc *Blockchain) Rollback(targetHash hash.Hash, utxo *lutxo.UTXOIndex, scState *scState.ScState) bool {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	if !bc.IsInBlockchain(targetHash) {
		return false
	}
	parentblockHash := bc.GetTailBlockHash()
	//if is child of tail, skip rollback
	if lblock.IsHashEqual(parentblockHash, targetHash) {
		return true
	}

	//keep rolling back blocks until the block with the input hash
	for bytes.Compare(parentblockHash, targetHash) != 0 {

		block, err := bc.GetBlockByHash(parentblockHash)
		logger.WithFields(logger.Fields{
			"height": block.GetHeight(),
			"hash":   parentblockHash.String(),
		}).Info("Blockchain: is about to rollback the block...")
		if err != nil {
			return false
		}
		parentblockHash = block.GetPrevHash()

		for _, tx := range block.GetTransactions() {
			if !tx.IsCoinbase() && !tx.IsRewardTx() && !tx.IsGasRewardTx() && !tx.IsGasChangeTx() {
				bc.txPool.Rollback(*tx)
			}
		}
	}

	bc.dbio.EnableBatch()
	defer bc.dbio.DisableBatch()

	err := bc.setTailBlockHash(parentblockHash)
	if err != nil {
		logger.Error("Blockchain: failed to set tail block hash during rollback!")
		return false
	}

	bc.txPool.SaveToDatabase(bc.dbio.TxPoolDBIO)

	utxo.Save()
	bc.dbio.ScStateDBIO.SaveToDatabase(scState)
	bc.dbio.Flush()

	return true
}

func (bc *Blockchain) setTailBlockHash(hash hash.Hash) error {
	return bc.bc.SetTailBlockHash(hash)
}

func (bc *Blockchain) DeepCopy() *Blockchain {
	newCopy := &Blockchain{}
	copier.Copy(newCopy, bc)
	return newCopy
}

func (bc *Blockchain) SetLIBHash(hash hash.Hash) error {
	return bc.bc.SetLIBHash(hash)
}

// GasPrice returns gas price in current blockchain
func (bc *Blockchain) GasPrice() uint64 {
	return DefaultGasPrice
}

func (bc *Blockchain) CheckLibPolicy(blk *block.Block) bool {
	//Do not check genesis block
	if blk.GetHeight() == 0 {
		return true
	}

	if bc.libPolicy.IsBypassingLibCheck() {
		return true
	}

	if !bc.libPolicy.IsNonRepeatingBlockProducerRequired() {
		return true
	}

	return !bc.checkRepeatingProducer(blk)
}

//checkRepeatingProducer returns true if it found a repeating block between the input block and last irreversible block
func (bc *Blockchain) checkRepeatingProducer(blk *block.Block) bool {
	lib := bc.GetLIBHash()

	libProduerNum := bc.libPolicy.GetMinConfirmationNum()

	existProducers := make(map[string]bool)
	currBlk := blk

	for i := 0; i < libProduerNum; i++ {
		if currBlk.GetHeight() == 0 {
			return false
		}

		if _, ok := existProducers[currBlk.GetProducer()]; ok {
			logger.WithFields(logger.Fields{
				"currBlkHeight": currBlk.GetHeight(),
				"producer":      currBlk.GetProducer(),
			}).Debug("Blockchain: repeating producer")
			return true
		}

		if lib.Equals(currBlk.GetHash()) {
			return false
		}

		existProducers[currBlk.GetProducer()] = true

		newBlock, err := bc.GetBlockByHash(currBlk.GetPrevHash())
		if err != nil {
			logger.WithError(err).Warn("Blockchain: Cant not read parent block while checking repeating producer")
			return true
		}

		currBlk = newBlock
	}
	return false
}

func (bc *Blockchain) updateLIB(currBlkHeight uint64) {
	if bc.libPolicy == nil {
		return
	}

	minConfirmationNum := bc.libPolicy.GetMinConfirmationNum()
	LIBHeight := uint64(0)
	if currBlkHeight > uint64(minConfirmationNum) {
		LIBHeight = currBlkHeight - uint64(minConfirmationNum)
	}

	LIBBlk, err := bc.GetBlockByHeight(LIBHeight)
	if err != nil {
		logger.WithError(err).Warn("Blockchain: Can not find LIB block in database")
		return
	}

	bc.bc.SetLIBHash(LIBBlk.GetHash())
}

// RevertUtxoAndScStateAtBlockHash returns the previous snapshot of UTXOIndex when the block of given hash was the tail block.
func RevertUtxoAndScStateAtBlockHash(bc *Blockchain, hash hash.Hash) (*lutxo.UTXOIndex, *scState.ScState, error) {
	index := lutxo.NewUTXOIndex(bc.dbio.UtxoDBIO)
	scState := bc.dbio.ScStateDBIO.LoadScStateFromDatabase()
	bci := bc.Iterator()

	// Start from the tail of blockchain, compute the previous UTXOIndex by undoing transactions
	// in the block, until the block hash matches.
	for {
		block, err := bci.Next()
		logger.WithFields(logger.Fields{
			"currBlkHeight":      block.GetHeight(),
			"currBlkHash":        block.GetHash(),
			"rollbackTargetHash": hash,
		}).Error("GetNextBlock")
		if bytes.Compare(block.GetHash(), hash) == 0 {
			break
		}

		if err != nil {
			return nil, nil, err
		}

		if len(block.GetPrevHash()) == 0 {
			logger.WithFields(logger.Fields{
				"currBlkHeight":      block.GetHeight(),
				"currBlkHash":        block.GetHash(),
				"rollbackTargetHash": hash,
			}).Error("Reached the beginning of the blockchain")
			return nil, nil, ErrBlockDoesNotExist
		}

		err = index.UndoTxsInBlock(block)

		if err != nil {
			logger.WithError(err).WithFields(logger.Fields{
				"hash": block.GetHash(),
			}).Warn("BlockchainManager: failed to calculate previous state of UTXO index for the block")
			return nil, nil, err
		}

		err = lScState.RevertState(bc.dbio.ScStateDBIO, block.GetHash(), scState)
		if err != nil {
			logger.WithError(err).WithFields(logger.Fields{
				"hash": block.GetHash(),
			}).Warn("BlockchainManager: failed to calculate previous state of scState for the block")
			return nil, nil, err
		}
	}

	return index, scState, nil
}
