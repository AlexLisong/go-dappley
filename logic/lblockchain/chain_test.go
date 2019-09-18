package lblockchain

import (
	"errors"
	"github.com/dappley/go-dappley/common/hash"
	"github.com/dappley/go-dappley/core/block"
	"github.com/dappley/go-dappley/core/blockchain"
	"github.com/dappley/go-dappley/logic/lblockchain/mocks"
	"github.com/dappley/go-dappley/storage"
	"github.com/dappley/go-dappley/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestChain_GetBlockByHash(t *testing.T) {
	tests := []struct {
		name         string
		serializedBp string
		libBlkHash   string
		libBlkHeight uint64
		blkHash      hash.Hash
		blocksInDb   []*block.Block
		expectedBlk  *block.Block
		expectedErr  error
	}{
		{
			"block not found",
			"1#2, 1#3, 3#4, 4#5, 4#6, 4#7, 2#8, 2#9, 8#10",
			"1",
			10,
			hash.Hash("d"),
			[]*block.Block{
				blockchain.CreateBlock(hash.Hash("a"), nil, 7),
				blockchain.CreateBlock(hash.Hash("b"), hash.Hash("a"), 8),
				blockchain.CreateBlock(hash.Hash("c"), hash.Hash("b"), 9),
				blockchain.CreateBlock(hash.Hash("1"), hash.Hash("c"), 10),
			},
			nil,
			ErrBlockDoesNotExist,
		},
		{
			"block found in forks",
			"1#2, 1#3, 3#4, 4#5, 4#6, 4#7, 2#8, 2#9, 8#10",
			"1",
			10,
			hash.Hash("2"),
			[]*block.Block{
				blockchain.CreateBlock(hash.Hash("a"), nil, 7),
				blockchain.CreateBlock(hash.Hash("b"), hash.Hash("a"), 8),
				blockchain.CreateBlock(hash.Hash("c"), hash.Hash("b"), 9),
				blockchain.CreateBlock(hash.Hash("1"), hash.Hash("c"), 10),
			},
			blockchain.CreateBlock(hash.Hash("2"), hash.Hash("1"), 11),
			nil,
		},
		{
			"block found in db",
			"1#2, 1#3, 3#4, 4#5, 4#6, 4#7, 2#8, 2#9, 8#10",
			"1",
			10,
			hash.Hash("b"),
			[]*block.Block{
				blockchain.CreateBlock(hash.Hash("a"), nil, 7),
				blockchain.CreateBlock(hash.Hash("b"), hash.Hash("a"), 8),
				blockchain.CreateBlock(hash.Hash("c"), hash.Hash("b"), 9),
				blockchain.CreateBlock(hash.Hash("1"), hash.Hash("c"), 10),
			},
			blockchain.CreateBlock(hash.Hash("b"), hash.Hash("a"), 8),
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp, _ := blockchain.DeserializeBlockPool(tt.serializedBp, blockchain.CreateBlock(hash.Hash(tt.libBlkHash), nil, tt.libBlkHeight))
			db := &mocks.Storage{}
			for _, blk := range tt.blocksInDb {
				db.On("Get", []byte(blk.GetHash())).Return(blk.Serialize(), nil)
			}
			db.On("Get", mock.MatchedBy(
				func(blkHash hash.Hash) bool {
					return true
				})).Return(nil, errors.New("Error"))

			bc := &Chain{forks: bp, db: db}
			blk, err := bc.GetBlockByHash(tt.blkHash)
			assert.Equal(t, tt.expectedBlk, blk)
			assert.Equal(t, tt.expectedErr, err)

		})
	}
}

func TestChain_GetBlockByHeight(t *testing.T) {
	tests := []struct {
		name         string
		serializedBp string
		libBlkHash   string
		libBlkHeight uint64
		blkHeight    uint64
		blocksInDb   []*block.Block
		expectedBlk  *block.Block
		expectedErr  error
	}{
		{
			"block not found",
			"1#2, 1#3, 3#4, 4#5, 4#6, 4#7, 2#8, 2#9, 8#10",
			"1",
			10,
			17,
			[]*block.Block{
				blockchain.CreateBlock(hash.Hash("a"), nil, 7),
				blockchain.CreateBlock(hash.Hash("b"), hash.Hash("a"), 8),
				blockchain.CreateBlock(hash.Hash("c"), hash.Hash("b"), 9),
				blockchain.CreateBlock(hash.Hash("1"), hash.Hash("c"), 10),
			},
			nil,
			ErrBlockDoesNotExist,
		},
		{
			"block found in forks",
			"1#2, 1#3, 3#4, 4#5, 4#6, 4#7, 2#8, 2#9, 8#10",
			"1",
			10,
			11,
			[]*block.Block{
				blockchain.CreateBlock(hash.Hash("a"), nil, 7),
				blockchain.CreateBlock(hash.Hash("b"), hash.Hash("a"), 8),
				blockchain.CreateBlock(hash.Hash("c"), hash.Hash("b"), 9),
				blockchain.CreateBlock(hash.Hash("1"), hash.Hash("c"), 10),
			},
			blockchain.CreateBlock(hash.Hash("2"), hash.Hash("1"), 11),
			nil,
		},
		{
			"block found in db",
			"1#2, 1#3, 3#4, 4#5, 4#6, 4#7, 2#8, 2#9, 8#10",
			"1",
			10,
			8,
			[]*block.Block{
				blockchain.CreateBlock(hash.Hash("a"), nil, 7),
				blockchain.CreateBlock(hash.Hash("b"), hash.Hash("a"), 8),
				blockchain.CreateBlock(hash.Hash("c"), hash.Hash("b"), 9),
				blockchain.CreateBlock(hash.Hash("1"), hash.Hash("c"), 10),
			},
			blockchain.CreateBlock(hash.Hash("b"), hash.Hash("a"), 8),
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp, _ := blockchain.DeserializeBlockPool(tt.serializedBp, blockchain.CreateBlock(hash.Hash(tt.libBlkHash), nil, tt.libBlkHeight))
			db := &mocks.Storage{}
			for _, blk := range tt.blocksInDb {
				db.On("Get", util.UintToHex(blk.GetHeight())).Return([]byte(blk.GetHash()), nil)
				db.On("Get", []byte(blk.GetHash())).Return(blk.Serialize(), nil)
			}
			db.On("Get", mock.MatchedBy(
				func(hash hash.Hash) bool {
					return true
				})).Return(nil, errors.New("Error"))
			db.On("Put", mock.Anything, mock.Anything).Return(nil)

			policy := &mocks.LIBPolicy{}
			policy.On("GetMinConfirmationNum").Return(3)

			bc := NewChain(
				blockchain.CreateBlock(hash.Hash(tt.libBlkHash), nil, tt.libBlkHeight),
				db,
				policy)
			bc.forks = bp
			bc.SetTailBlockHash(bp.GetHighestBlock().GetHash())
			bc.updateForkHeightCache()

			blk, err := bc.GetBlockByHeight(tt.blkHeight)
			assert.Equal(t, tt.expectedBlk, blk)
			assert.Equal(t, tt.expectedErr, err)

		})
	}
}

func TestChain_AddBlock(t *testing.T) {
	tests := []struct {
		name                  string
		serializedBp          string
		libBlk                *block.Block
		libMinNumOfBlocks     int
		blocksInDb            []*block.Block
		newBlk                *block.Block
		expectedErr           error
		expectedTailBlockHash string
		expectedLIBHash       string
		expectedBlksInDb      []*block.Block
		expectedForkHashes    []string
	}{
		{
			"Empty Blockpool",
			"",
			blockchain.CreateBlock(hash.Hash("1"), hash.Hash("c"), 10),
			3,
			[]*block.Block{
				blockchain.CreateBlock(hash.Hash("a"), nil, 7),
				blockchain.CreateBlock(hash.Hash("b"), hash.Hash("a"), 8),
				blockchain.CreateBlock(hash.Hash("c"), hash.Hash("b"), 9),
				blockchain.CreateBlock(hash.Hash("1"), hash.Hash("c"), 10),
			},
			blockchain.CreateBlock(hash.Hash("2"), hash.Hash("1"), 11),
			nil,
			"2",
			"1",
			[]*block.Block{
				blockchain.CreateBlock(hash.Hash("a"), nil, 7),
				blockchain.CreateBlock(hash.Hash("b"), hash.Hash("a"), 8),
				blockchain.CreateBlock(hash.Hash("c"), hash.Hash("b"), 9),
				blockchain.CreateBlock(hash.Hash("1"), hash.Hash("c"), 10),
			},
			[]string{"1", "2"},
		},
		{
			"Add Block To Tail",
			"1#2, 1#3, 3#4, 4#5, 4#6, 4#7, 2#8, 2#9, 8#10",
			blockchain.CreateBlock(hash.Hash("1"), hash.Hash("c"), 10),
			3,
			[]*block.Block{
				blockchain.CreateBlock(hash.Hash("a"), nil, 7),
				blockchain.CreateBlock(hash.Hash("b"), hash.Hash("a"), 8),
				blockchain.CreateBlock(hash.Hash("c"), hash.Hash("b"), 9),
				blockchain.CreateBlock(hash.Hash("1"), hash.Hash("c"), 10),
			},
			blockchain.CreateBlock(hash.Hash("11"), hash.Hash("10"), 14),
			nil,
			"11",
			"2",
			[]*block.Block{
				blockchain.CreateBlock(hash.Hash("a"), nil, 7),
				blockchain.CreateBlock(hash.Hash("b"), hash.Hash("a"), 8),
				blockchain.CreateBlock(hash.Hash("c"), hash.Hash("b"), 9),
				blockchain.CreateBlock(hash.Hash("1"), hash.Hash("c"), 10),
				blockchain.CreateBlock(hash.Hash("2"), hash.Hash("1"), 11),
			},
			[]string{"2", "8", "10", "11"},
		},
		{
			"Add orphan block",
			"1#2, 1#3, 3#4, 4#5, 4#6, 4#7, 2#8, 2#9, 8#10",
			blockchain.CreateBlock(hash.Hash("1"), hash.Hash("c"), 10),
			3,
			[]*block.Block{
				blockchain.CreateBlock(hash.Hash("a"), nil, 7),
				blockchain.CreateBlock(hash.Hash("b"), hash.Hash("a"), 8),
				blockchain.CreateBlock(hash.Hash("c"), hash.Hash("b"), 9),
				blockchain.CreateBlock(hash.Hash("1"), hash.Hash("c"), 10),
			},
			blockchain.CreateBlock(hash.Hash("11"), hash.Hash("d"), 14),
			nil,
			"10",
			"1",
			[]*block.Block{
				blockchain.CreateBlock(hash.Hash("a"), nil, 7),
				blockchain.CreateBlock(hash.Hash("b"), hash.Hash("a"), 8),
				blockchain.CreateBlock(hash.Hash("c"), hash.Hash("b"), 9),
				blockchain.CreateBlock(hash.Hash("1"), hash.Hash("c"), 10),
			},
			[]string{"1", "2", "8", "10"},
		},
		{
			"Add a low block",
			"1#2, 1#3, 3#4, 4#5, 4#6, 4#7, 2#8, 2#9, 8#10",
			blockchain.CreateBlock(hash.Hash("1"), hash.Hash("c"), 10),
			3,
			[]*block.Block{
				blockchain.CreateBlock(hash.Hash("a"), nil, 7),
				blockchain.CreateBlock(hash.Hash("b"), hash.Hash("a"), 8),
				blockchain.CreateBlock(hash.Hash("c"), hash.Hash("b"), 9),
				blockchain.CreateBlock(hash.Hash("1"), hash.Hash("c"), 10),
			},
			blockchain.CreateBlock(hash.Hash("11"), hash.Hash("d"), 8),
			ErrBlockHeightTooLow,
			"10",
			"1",
			[]*block.Block{
				blockchain.CreateBlock(hash.Hash("a"), nil, 7),
				blockchain.CreateBlock(hash.Hash("b"), hash.Hash("a"), 8),
				blockchain.CreateBlock(hash.Hash("c"), hash.Hash("b"), 9),
				blockchain.CreateBlock(hash.Hash("1"), hash.Hash("c"), 10),
			},
			[]string{"1", "2", "8", "10"},
		},
		{
			"Link an orphan to chain",
			"1#2, 1#3, 3#4, 4#5, 4#6, 4#7, 2#8, 2#9, 8#10, 14^11#12",
			blockchain.CreateBlock(hash.Hash("1"), hash.Hash("c"), 10),
			3,
			[]*block.Block{
				blockchain.CreateBlock(hash.Hash("a"), nil, 7),
				blockchain.CreateBlock(hash.Hash("b"), hash.Hash("a"), 8),
				blockchain.CreateBlock(hash.Hash("c"), hash.Hash("b"), 9),
				blockchain.CreateBlock(hash.Hash("1"), hash.Hash("c"), 10),
			},
			blockchain.CreateBlock(hash.Hash("11"), hash.Hash("9"), 13),
			nil,
			"12",
			"2",
			[]*block.Block{
				blockchain.CreateBlock(hash.Hash("a"), nil, 7),
				blockchain.CreateBlock(hash.Hash("b"), hash.Hash("a"), 8),
				blockchain.CreateBlock(hash.Hash("c"), hash.Hash("b"), 9),
				blockchain.CreateBlock(hash.Hash("1"), hash.Hash("c"), 10),
				blockchain.CreateBlock(hash.Hash("2"), hash.Hash("1"), 11),
			},
			[]string{"2", "9", "11", "12"},
		},
		{
			"Link a long orphan to chain",
			"1#2, 1#3, 3#4, 4#5, 4#6, 4#7, 2#8, 2#9, 8#10, 14^11#12, 12#13",
			blockchain.CreateBlock(hash.Hash("1"), hash.Hash("c"), 10),
			3,
			[]*block.Block{
				blockchain.CreateBlock(hash.Hash("a"), nil, 7),
				blockchain.CreateBlock(hash.Hash("b"), hash.Hash("a"), 8),
				blockchain.CreateBlock(hash.Hash("c"), hash.Hash("b"), 9),
				blockchain.CreateBlock(hash.Hash("1"), hash.Hash("c"), 10),
			},
			blockchain.CreateBlock(hash.Hash("11"), hash.Hash("9"), 13),
			nil,
			"13",
			"9",
			[]*block.Block{
				blockchain.CreateBlock(hash.Hash("a"), nil, 7),
				blockchain.CreateBlock(hash.Hash("b"), hash.Hash("a"), 8),
				blockchain.CreateBlock(hash.Hash("c"), hash.Hash("b"), 9),
				blockchain.CreateBlock(hash.Hash("1"), hash.Hash("c"), 10),
				blockchain.CreateBlock(hash.Hash("2"), hash.Hash("1"), 11),
				blockchain.CreateBlock(hash.Hash("9"), hash.Hash("2"), 12),
			},
			[]string{"9", "11", "12", "13"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			//Prepare blockchain
			bp, _ := blockchain.DeserializeBlockPool(tt.serializedBp, tt.libBlk)
			db := storage.NewRamStorage()

			libPolicy := &mocks.LIBPolicy{}
			libPolicy.On("GetMinConfirmationNum").Return(tt.libMinNumOfBlocks)

			bc := NewChain(
				tt.libBlk,
				db,
				libPolicy)
			bc.forks = bp
			bc.SetTailBlockHash(bp.GetHighestBlock().GetHash())

			for _, blk := range tt.blocksInDb {
				bc.AddBlockToDb(blk)
			}

			bc.updateForkHeightCache()

			//Add a new block to blockchain
			err := bc.AddBlock(tt.newBlk)

			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, hash.Hash(tt.expectedTailBlockHash), bc.GetTailBlockHash())
			assert.Equal(t, hash.Hash(tt.expectedLIBHash), bc.GetLIBBlockHash())

			//check if the new LIBs are stored in db
			for _, blk := range tt.expectedBlksInDb {
				rawBytes, err := db.Get(blk.GetHash())
				assert.Nil(t, err)
				assert.Equal(t, blk, block.Deserialize(rawBytes))
			}

			//Check forkhash cache. The cache should contain they key value pair height-blkhash for all blocks in the main fork
			currHeight := bc.GetLIBHeight()
			tailBlkHeight := bc.GetMaxHeight()
			count := 0
			for currHeight <= tailBlkHeight {
				blkHash, exists := bc.forkHash.Get(currHeight)
				assert.True(t, exists)
				assert.Equal(t, hash.Hash(tt.expectedForkHashes[count]), blkHash)
				currHeight++
				count++
			}

		})
	}
}
