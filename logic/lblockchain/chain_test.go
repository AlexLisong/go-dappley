package lblockchain

import (
	"errors"
	"github.com/dappley/go-dappley/common/hash"
	"github.com/dappley/go-dappley/core/block"
	"github.com/dappley/go-dappley/core/blockchain"
	"github.com/dappley/go-dappley/logic/lblockchain/mocks"
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
			},
			blockchain.CreateBlock(hash.Hash("b"), hash.Hash("a"), 8),
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp, _ := blockchain.DeserializeBlockPool(tt.serializedBp, tt.libBlkHash, tt.libBlkHeight)
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

func TestChain_AddBlock(t *testing.T) {
	tests := []struct {
		name                  string
		serializedBp          string
		libBlkHash            string
		libBlkHeight          uint64
		blocksInDb            []*block.Block
		newBlk                *block.Block
		expectedErr           error
		expectedTailBlockHash string
		expectedLIBHash       string
	}{
		{
			"Add Block To Tail",
			"1#2, 1#3, 3#4, 4#5, 4#6, 4#7, 2#8, 2#9, 8#10",
			"1",
			10,
			[]*block.Block{
				blockchain.CreateBlock(hash.Hash("a"), nil, 7),
				blockchain.CreateBlock(hash.Hash("b"), hash.Hash("a"), 8),
				blockchain.CreateBlock(hash.Hash("c"), hash.Hash("b"), 9),
			},
			blockchain.CreateBlock(hash.Hash("11"), hash.Hash("10"), 14),
			nil,
			"11",
			"2",
		},
		{
			"Add orphan block",
			"1#2, 1#3, 3#4, 4#5, 4#6, 4#7, 2#8, 2#9, 8#10",
			"1",
			10,
			[]*block.Block{
				blockchain.CreateBlock(hash.Hash("a"), nil, 7),
				blockchain.CreateBlock(hash.Hash("b"), hash.Hash("a"), 8),
				blockchain.CreateBlock(hash.Hash("c"), hash.Hash("b"), 9),
			},
			blockchain.CreateBlock(hash.Hash("11"), hash.Hash("d"), 14),
			nil,
			"10",
			"1",
		},
		{
			"Add a low block",
			"1#2, 1#3, 3#4, 4#5, 4#6, 4#7, 2#8, 2#9, 8#10",
			"1",
			10,
			[]*block.Block{
				blockchain.CreateBlock(hash.Hash("a"), nil, 7),
				blockchain.CreateBlock(hash.Hash("b"), hash.Hash("a"), 8),
				blockchain.CreateBlock(hash.Hash("c"), hash.Hash("b"), 9),
			},
			blockchain.CreateBlock(hash.Hash("11"), hash.Hash("d"), 8),
			ErrBlockHeightTooLow,
			"10",
			"1",
		},
		{
			"Link an orphan to chain",
			"1#2, 1#3, 3#4, 4#5, 4#6, 4#7, 2#8, 2#9, 8#10, 14^11#12",
			"1",
			10,
			[]*block.Block{
				blockchain.CreateBlock(hash.Hash("a"), nil, 7),
				blockchain.CreateBlock(hash.Hash("b"), hash.Hash("a"), 8),
				blockchain.CreateBlock(hash.Hash("c"), hash.Hash("b"), 9),
			},
			blockchain.CreateBlock(hash.Hash("11"), hash.Hash("9"), 13),
			nil,
			"12",
			"2",
		},
		{
			"Link a long orphan to chain",
			"1#2, 1#3, 3#4, 4#5, 4#6, 4#7, 2#8, 2#9, 8#10, 14^11#12, 12#13",
			"1",
			10,
			[]*block.Block{
				blockchain.CreateBlock(hash.Hash("a"), nil, 7),
				blockchain.CreateBlock(hash.Hash("b"), hash.Hash("a"), 8),
				blockchain.CreateBlock(hash.Hash("c"), hash.Hash("b"), 9),
			},
			blockchain.CreateBlock(hash.Hash("11"), hash.Hash("9"), 13),
			nil,
			"13",
			"9",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			bp, _ := blockchain.DeserializeBlockPool(tt.serializedBp, tt.libBlkHash, tt.libBlkHeight)
			db := &mocks.Storage{}
			for _, blk := range tt.blocksInDb {
				//Add block to db
				db.On("Get", []byte(blk.GetHash())).Return(blk.Serialize(), nil)
				db.On("Get", util.UintToHex(blk.GetHeight())).Return(blk.GetHash(), nil)
			}
			//any other block return err
			db.On("Get", mock.MatchedBy(
				func(blkHash hash.Hash) bool {
					return true
				})).Return(nil, errors.New("Error"))

			db.On("Put", mock.Anything, mock.Anything).Return(nil)

			libPolicy := &mocks.LIBPolicy{}
			libPolicy.On("GetMinConfirmationNum").Return(3)

			bc := NewChain(
				bp.GetHighestBlock(),
				blockchain.CreateBlock(hash.Hash(tt.libBlkHash), nil, tt.libBlkHeight),
				db,
				libPolicy)
			bc.forks = bp
			bc.updateForkHeightCache()

			err := bc.AddBlock(tt.newBlk)
			bc.forks.PrintInfo()

			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, hash.Hash(tt.expectedTailBlockHash), bc.GetTailBlockHash())
			assert.Equal(t, hash.Hash(tt.expectedLIBHash), bc.GetLIBBlockHash())

		})
	}
}
