package lblockchain

import (
	"testing"

	"github.com/dappley/go-dappley/logic/lblockchain/mocks"
	"github.com/dappley/go-dappley/logic/ltransactionpool"

	"github.com/dappley/go-dappley/common"
	"github.com/dappley/go-dappley/common/hash"
	"github.com/dappley/go-dappley/core"
	"github.com/dappley/go-dappley/core/block"
	"github.com/dappley/go-dappley/core/transaction"
	"github.com/dappley/go-dappley/core/transactionbase"
	"github.com/dappley/go-dappley/core/utxo"
	"github.com/dappley/go-dappley/logic/lblock"
	"github.com/dappley/go-dappley/logic/lutxo"
	logger "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/dappley/go-dappley/core/account"
	"github.com/dappley/go-dappley/storage"
)

func TestGetUTXOIndexAtBlockHash(t *testing.T) {
	genesisAddr := account.NewAddress("##@@")
	genesisBlock := NewGenesisBlock(genesisAddr, transaction.Subsidy)

	// prepareBlockchainWithBlocks returns a blockchain that contains the given blocks with correct utxoIndex in RAM
	prepareBlockchainWithBlocks := func(blks []*block.Block) *Blockchain {
		policy := &mocks.LIBPolicy{}
		policy.On("GetMinConfirmationNum").Return(3)
		bc := CreateBlockchain(genesisAddr, storage.NewRamStorage(), policy, ltransactionpool.NewTransactionPoolLogic(nil, 128000), nil, 100000)
		for _, blk := range blks {
			err := bc.AddBlockWithContext(PrepareBlockContext(bc, blk))
			if err != nil {
				logger.Fatal("TestGetUTXOIndexAtBlockHash: cannot add the blocks to blockchain.")
			}
		}
		return bc
	}

	// utxoIndexFromTXs creates a utxoIndex containing all vout of transactions in txs
	utxoIndexFromTXs := func(txs []*transaction.Transaction, cache *storage.UTXODBIO) *lutxo.UTXOIndex {
		utxoIndex := lutxo.NewUTXOIndex(cache)
		utxosMap := make(map[string]*utxo.UTXOTx)
		for _, tx := range txs {
			for i, vout := range tx.Vout {
				utxos, ok := utxosMap[vout.Account.GetPubKeyHash().String()]
				if !ok {
					newUtxos := utxo.NewUTXOTx()
					utxos = &newUtxos
				}
				utxos.PutUtxo(utxo.NewUTXO(vout, tx.ID, i, utxo.UtxoNormal))
				utxosMap[vout.Account.GetPubKeyHash().String()] = utxos
			}
		}
		utxoIndex.SetIndex(utxosMap)
		return utxoIndex
	}
	acc := account.NewAccount()

	normalTX := transaction.NewCoinbaseTX(acc.GetAddress(), "", 1, common.NewAmount(5))
	normalTX2 := transaction.Transaction{
		hash.Hash("normal2"),
		[]transactionbase.TXInput{{normalTX.ID, 0, nil, acc.GetKeyPair().GetPublicKey()}},
		[]transactionbase.TXOutput{{common.NewAmount(5), acc.GetTransactionAccount(), ""}},
		common.NewAmount(0),
		common.NewAmount(0),
		common.NewAmount(0),
		0,
	}
	abnormalTX := transaction.Transaction{
		hash.Hash("abnormal"),
		[]transactionbase.TXInput{{normalTX.ID, 1, nil, nil}},
		[]transactionbase.TXOutput{{common.NewAmount(5), account.NewContractAccountByPubKeyHash(account.PubKeyHash([]byte("pkh"))), ""}},
		common.NewAmount(0),
		common.NewAmount(0),
		common.NewAmount(0),
		0,
	}
	prevBlock := block.NewBlock([]*transaction.Transaction{}, genesisBlock, "")
	prevBlock.SetHash(lblock.CalculateHash(prevBlock))
	emptyBlock := block.NewBlock([]*transaction.Transaction{}, prevBlock, "")
	emptyBlock.SetHash(lblock.CalculateHash(emptyBlock))
	normalBlock := block.NewBlock([]*transaction.Transaction{&normalTX}, genesisBlock, "")
	normalBlock.SetHash(lblock.CalculateHash(normalBlock))
	normalBlock2 := block.NewBlock([]*transaction.Transaction{&normalTX2}, normalBlock, "")
	normalBlock2.SetHash(lblock.CalculateHash(normalBlock2))
	abnormalBlock := block.NewBlock([]*transaction.Transaction{&abnormalTX}, normalBlock, "")
	abnormalBlock.SetHash(lblock.CalculateHash(abnormalBlock))
	corruptedUTXOBlockchain := prepareBlockchainWithBlocks([]*block.Block{normalBlock, normalBlock2})
	err := utxoIndexFromTXs([]*transaction.Transaction{&normalTX}, corruptedUTXOBlockchain.GetDBIO().UtxoDBIO).Save()
	if err != nil {
		logger.Fatal("TestGetUTXOIndexAtBlockHash: cannot corrupt the utxoIndex in database.")
	}

	policy := &mocks.LIBPolicy{}
	policy.On("GetMinConfirmationNum").Return(3)

	bcs := []*Blockchain{
		prepareBlockchainWithBlocks([]*block.Block{normalBlock}),
		prepareBlockchainWithBlocks([]*block.Block{normalBlock, normalBlock2}),
		CreateBlockchain(account.NewAddress(""), storage.NewRamStorage(), policy, ltransactionpool.NewTransactionPoolLogic(nil, 128000), nil, 100000),
		prepareBlockchainWithBlocks([]*block.Block{prevBlock, emptyBlock}),
		prepareBlockchainWithBlocks([]*block.Block{normalBlock, normalBlock2}),
		prepareBlockchainWithBlocks([]*block.Block{normalBlock, abnormalBlock}),
		corruptedUTXOBlockchain,
	}
	tests := []struct {
		name     string
		bc       *Blockchain
		hash     hash.Hash
		expected *lutxo.UTXOIndex
		err      error
	}{
		{
			name:     "current block",
			bc:       bcs[0],
			hash:     normalBlock.GetHash(),
			expected: utxoIndexFromTXs([]*transaction.Transaction{&normalTX}, bcs[0].GetDBIO().UtxoDBIO),
			err:      nil,
		},
		{
			name:     "previous block",
			bc:       bcs[1],
			hash:     normalBlock.GetHash(),
			expected: utxoIndexFromTXs([]*transaction.Transaction{&normalTX}, bcs[1].GetDBIO().UtxoDBIO), // should not have utxo from normalTX2
			err:      nil,
		},
		{
			name:     "block not found",
			bc:       bcs[2],
			hash:     hash.Hash("not there"),
			expected: lutxo.NewUTXOIndex(bcs[2].GetDBIO().UtxoDBIO),
			err:      ErrBlockDoesNotExist,
		},
		{
			name:     "no txs in blocks",
			bc:       bcs[3],
			hash:     emptyBlock.GetHash(),
			expected: utxoIndexFromTXs(genesisBlock.GetTransactions(), bcs[3].GetDBIO().UtxoDBIO),
			err:      nil,
		},
		{
			name:     "genesis block",
			bc:       bcs[4],
			hash:     genesisBlock.GetHash(),
			expected: utxoIndexFromTXs(genesisBlock.GetTransactions(), bcs[4].GetDBIO().UtxoDBIO),
			err:      nil,
		},
		{
			name:     "utxo not found",
			bc:       bcs[5],
			hash:     normalBlock.GetHash(),
			expected: lutxo.NewUTXOIndex(bcs[5].GetDBIO().UtxoDBIO),
			err:      lutxo.ErrUTXONotFound,
		},
		{
			name:     "corrupted utxoIndex",
			bc:       bcs[6],
			hash:     normalBlock.GetHash(),
			expected: lutxo.NewUTXOIndex(bcs[6].GetDBIO().UtxoDBIO),
			err:      lutxo.ErrUTXONotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utxoIndex := lutxo.NewUTXOIndex(tt.bc.GetDBIO().UtxoDBIO)
			logger.WithFields(logger.Fields{
				"normalBlkHash": utxoIndex.GetAllUTXOsByPubKeyHash(acc.GetPubKeyHash()),
			}).Error(tt.bc.GetTailBlockHash())
			_, _, err = RevertUtxoAndScStateAtBlockHash(tt.bc.GetDBIO(), tt.bc, tt.hash)
			if !assert.Equal(t, tt.err, err) {
				return
			}
		})
	}
}

func TestCopyAndRevertUtxos(t *testing.T) {
	db := storage.NewRamStorage()
	defer db.Close()

	coinbaseAddr := account.NewAddress("testaddress")
	policy := &mocks.LIBPolicy{}
	policy.On("GetMinConfirmationNum").Return(3)
	bc := CreateBlockchain(coinbaseAddr, db, policy, ltransactionpool.NewTransactionPoolLogic(nil, 128000), nil, 100000)

	blk1 := core.GenerateUtxoMockBlockWithoutInputs(bc.GetTailBlockHash()) // contains 2 UTXOs for address1
	blk2 := core.GenerateUtxoMockBlockWithInputs()                         // contains tx that transfers address1's UTXOs to address2 with a change

	bc.AddBlockWithContext(PrepareBlockContext(bc, blk1))
	bc.AddBlockWithContext(PrepareBlockContext(bc, blk2))

	utxoIndex := lutxo.NewUTXOIndex(bc.GetDBIO().UtxoDBIO)

	var address1Bytes = []byte("address1000000000000000000000000")
	var address2Bytes = []byte("address2000000000000000000000000")
	var ta1 = account.NewTransactionAccountByPubKey(address1Bytes)
	var ta2 = account.NewTransactionAccountByPubKey(address2Bytes)

	addr1UTXOs := utxoIndex.GetAllUTXOsByPubKeyHash([]byte(ta1.GetPubKeyHash()))
	addr2UTXOs := utxoIndex.GetAllUTXOsByPubKeyHash([]byte(ta2.GetPubKeyHash()))
	// Expect address1 to have 1 utxo of $4
	assert.Equal(t, 1, addr1UTXOs.Size())
	utxo1 := addr1UTXOs.GetAllUtxos()[0]
	assert.Equal(t, common.NewAmount(4), utxo1.Value)

	// Expect address2 to have 2 utxos totaling $8
	assert.Equal(t, 2, addr2UTXOs.Size())

	// Rollback to blk1, address1 has a $5 utxo and a $7 utxo, total $12, and address2 has nothing
	indexSnapshot, _, err := RevertUtxoAndScStateAtBlockHash(bc.GetDBIO(), bc, blk1.GetHash())
	if err != nil {
		panic(err)
	}

	addr1UtxoTx := indexSnapshot.GetAllUTXOsByPubKeyHash(ta1.GetPubKeyHash())
	assert.Equal(t, 2, addr1UtxoTx.Size())

	tx1 := core.MockUtxoTransactionWithoutInputs()

	assert.Equal(t, common.NewAmount(5), addr1UtxoTx.GetUtxo(tx1.ID, 0).Value)
	assert.Equal(t, common.NewAmount(7), addr1UtxoTx.GetUtxo(tx1.ID, 1).Value)
	assert.Equal(t, 0, indexSnapshot.GetAllUTXOsByPubKeyHash(ta2.GetPubKeyHash()).Size())
}
