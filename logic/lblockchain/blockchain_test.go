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
	"os"
	"testing"

	"github.com/dappley/go-dappley/core/scState"
	"github.com/dappley/go-dappley/core/transaction"
	"github.com/dappley/go-dappley/logic/lScState"
	"github.com/dappley/go-dappley/logic/ltransaction"
	"github.com/dappley/go-dappley/logic/lutxo"
	"github.com/dappley/go-dappley/logic/transactionpool"

	"github.com/dappley/go-dappley/common/hash"
	"github.com/dappley/go-dappley/core"
	"github.com/dappley/go-dappley/core/block"
	"github.com/dappley/go-dappley/logic/lblock"

	"github.com/dappley/go-dappley/common"
	"github.com/dappley/go-dappley/core/account"
	bcMocks "github.com/dappley/go-dappley/logic/lblockchain/mocks"
	"github.com/dappley/go-dappley/storage"

	logger "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	logger.SetLevel(logger.WarnLevel)
	retCode := m.Run()
	os.Exit(retCode)
}

func TestCreateBlockchain(t *testing.T) {
	//create a new block chain
	s := storage.NewRamStorage()
	defer s.Close()

	addr := account.NewAddress("16PencPNnF8CiSx2EBGEd1axhf7vuHCouj")
	policy := &bcMocks.LIBPolicy{}
	policy.On("GetMinConfirmationNum").Return(3)

	bc := CreateBlockchain(addr, s, policy, transactionpool.NewTransactionPool(nil, 128), nil, 1000000)

	//find next block. This block should be the genesis block and its prev hash should be empty
	blk, err := bc.Next()
	assert.Nil(t, err)
	assert.Empty(t, blk.GetPrevHash())
}

func TestBlockchain_SetTailBlockHash(t *testing.T) {
	s := storage.NewRamStorage()
	defer s.Close()

	addr := account.NewAddress("16PencPNnF8CiSx2EBGEd1axhf7vuHCouj")
	policy := &bcMocks.LIBPolicy{}
	policy.On("GetMinConfirmationNum").Return(3)
	bc := CreateBlockchain(addr, s, policy, transactionpool.NewTransactionPool(nil, 128), nil, 1000000)

	tailHash := hash.Hash("TestHash")
	bc.SetTailBlockHash(tailHash)
	assert.Equal(t, tailHash, bc.GetTailBlockHash())

	newTailHash := hash.Hash("NewTestHash")
	bc.SetTailBlockHash(newTailHash)
	assert.NotEqual(t, tailHash, bc.GetTailBlockHash())
}

func TestBlockchain_HigherThanBlockchainTestHigher(t *testing.T) {
	//create a new block chain
	s := storage.NewRamStorage()
	defer s.Close()

	addr := account.NewAddress("16PencPNnF8CiSx2EBGEd1axhf7vuHCouj")
	policy := &bcMocks.LIBPolicy{}
	policy.On("GetMinConfirmationNum").Return(3)
	bc := CreateBlockchain(addr, s, policy, transactionpool.NewTransactionPool(nil, 128), nil, 1000000)
	blk := block.GenerateMockBlock()
	blk.SetHeight(1)
	assert.True(t, blk.GetHeight() > bc.GetMaxHeight())
}

func TestBlockchain_HigherThanBlockchainTestLower(t *testing.T) {
	//create a new block chain
	s := storage.NewRamStorage()
	defer s.Close()

	addr := account.NewAddress("16PencPNnF8CiSx2EBGEd1axhf7vuHCouj")
	policy := &bcMocks.LIBPolicy{}
	policy.On("GetMinConfirmationNum").Return(3)
	bc := CreateBlockchain(addr, s, policy, transactionpool.NewTransactionPool(nil, 128), nil, 1000000)
	tailblk, _ := bc.GetTailBlock()
	blk := ltransaction.GenerateBlockWithCbtx(addr, tailblk)
	blk.SetHeight(1)
	bc.AddBlockWithContext(PrepareBlockContext(bc, blk))

	assert.False(t, blk.GetHeight() > bc.GetMaxHeight())

}

func TestBlockchain_IsInBlockchain(t *testing.T) {
	//create a new block chain
	s := storage.NewRamStorage()
	defer s.Close()

	addr := account.NewAddress("16PencPNnF8CiSx2EBGEd1axhf7vuHCouj")
	policy := &bcMocks.LIBPolicy{}
	policy.On("GetMinConfirmationNum").Return(3)
	bc := CreateBlockchain(addr, s, policy, transactionpool.NewTransactionPool(nil, 128), nil, 100000)

	blk := core.GenerateUtxoMockBlockWithoutInputs(bc.GetTailBlockHash())
	bc.AddBlockWithContext(PrepareBlockContext(bc, blk))

	isFound := bc.IsInBlockchain([]byte("hash"))
	assert.True(t, isFound)

	isFound = bc.IsInBlockchain([]byte("hash2"))
	assert.False(t, isFound)
}

func TestBlockchain_RollbackToABlock(t *testing.T) {
	//create a mock blockchain with max height of 5
	bc := GenerateMockBlockchainWithCoinbaseTxOnly(5)
	defer bc.db.Close()

	//find the hash at height 3
	blk, err := bc.GetBlockByHeight(3)
	assert.Nil(t, err)

	//rollback to height 3
	bc.Rollback(blk.GetHash(), lutxo.NewUTXOIndex(bc.GetUtxoDBIO()), scState.NewScState())

	//the height 3 block should be the new tail block
	newTailBlk, err := bc.GetTailBlock()
	assert.Nil(t, err)
	assert.Equal(t, blk.GetHash(), newTailBlk.GetHash())

}

func TestBlockchain_AddBlockToTail(t *testing.T) {

	db := storage.NewRamStorage()

	// Create a blockchain for testing
	addr := account.NewAddress("dGDrVKjCG3sdXtDUgWZ7Fp3Q97tLhqWivf")
	policy := &bcMocks.LIBPolicy{}
	policy.On("GetMinConfirmationNum").Return(3)
	bc := CreateBlockchain(addr, db, policy, transactionpool.NewTransactionPool(nil, 128), nil, 1000000)

	tailBlk, _ := bc.GetTailBlock()
	// Add new block
	blk := block.NewBlock([]*transaction.Transaction{}, tailBlk, "")
	blk.SetHash([]byte("hash1"))

	blk.SetHeight(1)
	err := bc.AddBlockWithContext(PrepareBlockContext(bc, blk))

	// Expect the coinbase tx to go through
	assert.Equal(t, nil, err)
	// Expect that the block added is the blockchain tail
	assert.Equal(t, blk.GetHash(), hash.Hash(bc.GetTailBlockHash()))
}

func BenchmarkBlockchain_AddBlockToTail(b *testing.B) {
	//create a new block chain

	s := storage.NewRamStorage()
	addr := account.NewAddress("16PencPNnF8CiSx2EBGEd1axhf7vuHCouj")
	policy := &bcMocks.LIBPolicy{}
	policy.On("GetMinConfirmationNum").Return(3)
	bc := CreateBlockchain(addr, s, policy, transactionpool.NewTransactionPool(nil, 1280000), nil, 100000)
	var accounts []*account.Account
	for i := 0; i < 10; i++ {
		acc := account.NewAccount()
		accounts = append(accounts, acc)
	}

	for i := 0; i < b.N; i++ {

		tailBlk, _ := bc.GetTailBlock()
		txs := []*transaction.Transaction{}
		utxo := lutxo.NewUTXOIndex(bc.GetUtxoDBIO())
		cbtx := transaction.NewCoinbaseTX(accounts[0].GetAddress(), "", uint64(i+1), common.NewAmount(0))
		utxo.UpdateUtxo(&cbtx)
		txs = append(txs, &cbtx)
		for j := 0; j < 10; j++ {
			sendTxParam := transaction.NewSendTxParam(accounts[0].GetAddress(), accounts[0].GetKeyPair(), accounts[i%10].GetAddress(), common.NewAmount(1), common.NewAmount(0), common.NewAmount(0), common.NewAmount(0), "")
			tx, _ := transaction.NewUTXOTransaction(utxo.GetAllUTXOsByPubKeyHash(accounts[0].GetPubKeyHash()).GetAllUtxos(), sendTxParam)
			utxo.UpdateUtxo(&tx)
			txs = append(txs, &tx)
		}

		b := block.NewBlock(txs, tailBlk, "")
		b.SetHash(lblock.CalculateHash(b))
		state := lScState.LoadScStateFromDatabase(bc.GetDb())
		bc.AddBlockWithContext(&BlockContext{Block: b, UtxoIndex: utxo, State: state})

	}
}

func GenerateMockBlockchain(size int) *Blockchain {
	//create a new block chain
	s := storage.NewRamStorage()

	addr := account.NewAddress("16PencPNnF8CiSx2EBGEd1axhf7vuHCouj")
	policy := &bcMocks.LIBPolicy{}
	policy.On("GetMinConfirmationNum").Return(3)
	bc := CreateBlockchain(addr, s, policy, transactionpool.NewTransactionPool(nil, 128000), nil, 100000)

	for i := 0; i < size; i++ {
		tailBlk, _ := bc.GetTailBlock()
		b := block.NewBlock([]*transaction.Transaction{core.MockTransaction()}, tailBlk, "16PencPNnF8CiSx2EBGEd1axhf7vuHCouj")
		b.SetHash(lblock.CalculateHash(b))
		bc.AddBlockWithContext(PrepareBlockContext(bc, b))
	}
	return bc
}
