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

package blockchain

import (
	"github.com/dappley/go-dappley/common/hash"
	"github.com/dappley/go-dappley/core/block"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dappley/go-dappley/common"
)

func TestLRUCacheWithIntKeyAndValue(t *testing.T) {
	bp := NewBlockPool(nil)
	assert.Equal(t, 0, bp.blkCache.Len())
	const addCount = 200
	for i := 0; i < addCount; i++ {
		if bp.blkCache.Len() == ForkCacheLRUCacheLimit {
			bp.blkCache.RemoveOldest()
		}
		bp.blkCache.Add(i, i)
	}
	//test blkCache is full
	assert.Equal(t, ForkCacheLRUCacheLimit, bp.blkCache.Len())
	//test blkCache contains last added key
	assert.Equal(t, true, bp.blkCache.Contains(199))
	//test blkCache oldest key = addcount - BlockPoolLRUCacheLimit
	assert.Equal(t, addCount-ForkCacheLRUCacheLimit, bp.blkCache.Keys()[0])
}

func TestBlockPool_ForkHeadRange(t *testing.T) {
	bp := NewBlockPool(nil)

	parent := block.NewBlockWithRawInfo(hash.Hash("parent"), []byte{0}, 0, 0, 1, nil)
	blk := block.NewBlockWithRawInfo(hash.Hash("blk"), parent.GetHash(), 0, 0, 2, nil)
	child := block.NewBlockWithRawInfo(hash.Hash("child"), blk.GetHash(), 0, 0, 3, nil)

	// cache a blk
	bp.AddBlock(blk)
	readBlk, isFound := bp.blkCache.Get(blk.GetHash().String())
	require.Equal(t, blk, readBlk.(*common.TreeNode).GetValue().(*block.Block))
	require.True(t, isFound)
	require.ElementsMatch(t, []string{blk.GetHash().String()}, testGetForkHeadHashes(bp))

	// attach child to blk
	bp.AddBlock(child)
	require.ElementsMatch(t, []string{blk.GetHash().String()}, testGetForkHeadHashes(bp))

	// attach parent to blk
	bp.AddBlock(parent)
	require.ElementsMatch(t, []string{parent.GetHash().String()}, testGetForkHeadHashes(bp))

	// cache extraneous block
	unrelatedBlk := block.NewBlockWithRawInfo(hash.Hash("unrelated"), []byte{0}, 0, 0, 1, nil)
	bp.AddBlock(unrelatedBlk)
	require.ElementsMatch(t, []string{parent.GetHash().String(), unrelatedBlk.GetHash().String()}, testGetForkHeadHashes(bp))

	// remove parent
	bp.RemoveFork([]*block.Block{parent})
	require.ElementsMatch(t, []string{unrelatedBlk.GetHash().String()}, testGetForkHeadHashes(bp))

	// remove unrelated
	bp.RemoveFork([]*block.Block{unrelatedBlk})
	require.Nil(t, testGetForkHeadHashes(bp))
}

func TestBlockPool_isBlockValid(t *testing.T) {
	tests := []struct {
		name     string
		rootBlk  *block.Block
		blk      *block.Block
		expected bool
	}{
		{
			"Empty Block",
			CreateBlock(hash.Hash("child"), nil, 0),
			nil,
			false,
		},
		{
			"No rootBlkNode",
			nil,
			CreateBlock(hash.Hash("child"), hash.Hash("parent"), 0),
			true,
		},
		{
			"rootBlk is parent and input block is 1 block higher",
			CreateBlock(hash.Hash("1"), hash.Hash("0"), 0),
			CreateBlock(hash.Hash("2"), hash.Hash("1"), 1),
			true,
		},
		{
			"rootBlk is not parent and input block is 1 block higher",
			CreateBlock(hash.Hash("1"), hash.Hash("0"), 0),
			CreateBlock(hash.Hash("2"), hash.Hash("3"), 1),
			false,
		},
		{
			"input block is more than 1 block higher than rootBlk",
			CreateBlock(hash.Hash("1"), hash.Hash("0"), 0),
			CreateBlock(hash.Hash("2"), hash.Hash("3"), 2),
			true,
		},
		{
			"input block is same height as rootBlk",
			CreateBlock(hash.Hash("1"), hash.Hash("0"), 0),
			CreateBlock(hash.Hash("2"), hash.Hash("3"), 2),
			true,
		},
		{
			"input block is lower than rootBlk",
			CreateBlock(hash.Hash("1"), hash.Hash("0"), 5),
			CreateBlock(hash.Hash("2"), hash.Hash("3"), 2),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := NewBlockPool(tt.rootBlk)
			assert.Equal(t, tt.expected, bp.isBlockValid(tt.blk))
		})
	}
}

func TestBlockPool_removeTree(t *testing.T) {
	/*  BLOCK FORK STRUCTURE
	MAIN FORK:		     1           ORPHANS:(3 orphan forks)
				    2        3
				  8  9     4                              15
				10	     5 6 7              11          16
	                                      12  13					17
	                                            14
	*/

	tests := []struct {
		name                   string
		serializedBp           string
		rootBlkHash            string
		treeRoot               string
		expectedNumOfNodesLeft int
	}{
		{
			"Remove from main fork",
			"1#2, 1#3, 3#4, 4#5, 4#6, 4#7, 2#8, 2#9, 8#10, 3^11, 11#12, 11#13, 13#14, 2^15, 15#16, 4^17",
			"1",
			"3",
			12,
		},
		{
			"Remove from orphan 1",
			"1#2, 1#3, 3#4, 4#5, 4#6, 4#7, 2#8, 2#9, 8#10, 3^11, 11#12, 11#13, 13#14, 2^15, 15#16, 4^17",
			"1",
			"13",
			15,
		},
		{
			"Remove all from orphan 2",
			"1#2, 1#3, 3#4, 4#5, 4#6, 4#7, 2#8, 2#9, 8#10, 3^11, 11#12, 11#13, 13#14, 2^15, 15#16, 4^17",
			"1",
			"15",
			15,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp, _ := DeserializeBlockPool(tt.serializedBp, tt.rootBlkHash, 0)
			node, ok := bp.blkCache.Get(hash.Hash(tt.treeRoot).String())
			assert.True(t, ok)
			bp.removeTree(node.(*common.TreeNode))
			assert.Equal(t, tt.expectedNumOfNodesLeft, bp.blkCache.Len())
		})
	}
}

func TestBlockPool_GetGetHighestBlockBlock(t *testing.T) {
	/*  BLOCK FORK STRUCTURE
	MAIN FORK:		     1
				    2        3
				  8  9     4
				10	     5 6 7

	*/

	tests := []struct {
		name            string
		serializedBp    string
		rootBlkHash     string
		expectedBlkHash string
	}{
		{
			"Single node",
			"",
			"1",
			"1",
		},
		{
			"Normal case",
			"1#2, 1#3, 3#4, 4#5, 4#6, 4#7, 2#8, 2#9, 8#10",
			"1",
			"10",
		},
		{
			"Normal case",
			"1#2, 1#3, 3#4, 4#5, 4#6, 4#7, 2#8, 2#9, 8#10, 7#11",
			"1",
			"11",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp, _ := DeserializeBlockPool(tt.serializedBp, tt.rootBlkHash, 0)
			bp.PrintInfo()
			assert.Equal(t, hash.Hash(tt.expectedBlkHash), bp.GetHighestBlock().GetHash())
		})
	}
}

func TestBlockPool_AddBlock(t *testing.T) {
	/*  BLOCK FORK STRUCTURE
	MAIN FORK:		     1
				    2        3
				  8  9     4
				10	     5 6 7

	*/

	tests := []struct {
		name                  string
		serializedBp          string
		rootBlkHash           string
		rootBlkHeight         uint64
		newBlk                *block.Block
		expectedTailBlockHash string
		expectedLIBHash       string
		expectedNumOfNodes    int
		expectedNumOfOrphans  int
	}{
		{
			"Add Block To Tail",
			"1#2, 1#3, 3#4, 4#5, 4#6, 4#7, 2#8, 2#9, 8#10",
			"1",
			0,
			CreateBlock(hash.Hash("new"), hash.Hash("10"), 4),
			"new",
			"1",
			11,
			0,
		},
		{
			"Add Block To an orphan",
			"1#2, 1#3, 3#4, 4#5, 4#6, 4#7, 2#8, 2#9, 8#10, 3^11",
			"1",
			0,
			CreateBlock(hash.Hash("new"), hash.Hash("11"), 4),
			"10",
			"1",
			12,
			1,
		},
		{
			"Add an orphan block",
			"1#2, 1#3, 3#4, 4#5, 4#6, 4#7, 2#8, 2#9, 8#10, 3^11",
			"1",
			0,
			CreateBlock(hash.Hash("new"), hash.Hash("12"), 4),
			"10",
			"1",
			12,
			2,
		},
		{
			"Add block to the top of an orphan fork",
			"1#2, 1#3, 3#4, 4#5, 4#6, 4#7, 2#8, 2#9, 8#10, 3^13#11",
			"1",
			0,
			CreateBlock(hash.Hash("13"), hash.Hash("12"), 2),
			"10",
			"1",
			12,
			1,
		},
		{
			"Link an orphan to main fork",
			"1#2, 1#3, 3#4, 4#5, 4#6, 4#7, 2#8, 2#9, 8#10, 3^13#11",
			"1",
			0,
			CreateBlock(hash.Hash("13"), hash.Hash("2"), 2),
			"10",
			"1",
			12,
			0,
		},
		{
			"Link an higher height orphan to main fork",
			"1#2, 1#3, 3#4, 4#5, 4#6, 4#7, 2#8, 2#9, 8#10, 3^13#11, 11#14, 14#15",
			"1",
			0,
			CreateBlock(hash.Hash("13"), hash.Hash("2"), 2),
			"15",
			"1",
			14,
			0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp, _ := DeserializeBlockPool(tt.serializedBp, tt.rootBlkHash, tt.rootBlkHeight)
			bp.AddBlock(tt.newBlk)
			assert.Equal(t, hash.Hash(tt.expectedLIBHash).String(), getKey(bp.root))
			assert.Equal(t, hash.Hash(tt.expectedTailBlockHash).String(), bp.GetHighestBlock().GetHash().String())
			assert.Equal(t, tt.expectedNumOfNodes, bp.blkCache.Len())
			assert.Equal(t, tt.expectedNumOfOrphans, len(bp.orphans))
		})
	}
}

func TestBlockPool_SetRootBlock(t *testing.T) {
	/*  Test Block Pool Structure
		Blkgheight 				MAIN FORK:		     	ORPHANS:(3 orphan forks)
	         0							1
			 1						2        3
			 2					  8  9     4                              15
			 3					10	     5 6 7              11          16
			 4											  12  13						17
			 5													14
	*/
	tests := []struct {
		name                 string
		serializedBp         string
		rootBlkHash          string
		newRootBlkHash       string
		expectedNumOfNodes   int
		expectedNumOfOrphans int
	}{
		/*  Expected Result
		Blkgheight 				MAIN FORK:		     	ORPHANS:
			 0
			 1						       3
			 2					         4
			 3						   5 6 7              11
			 4											12  13						17
			 5												  14
		*/
		{
			"Set rootBlkHash to a descendant in main fork upper section",
			"1#2, 1#3, 3#4, 4#5, 4#6, 4#7, 2#8, 2#9, 8#10, 3^11, 11#12, 11#13, 13#14, 2^15, 15#16, 4^17",
			"1",
			"3",
			10,
			2,
		},

		/*  Expected Result
		Blkgheight 				MAIN FORK:		     	ORPHANS:
			 0
			 1
			 2					           4
			 3						     5 6 7
			 4											    						17
			 5
		*/
		{
			"Set rootBlkHash to a descendant in main fork middle section",
			"1#2, 1#3, 3#4, 4#5, 4#6, 4#7, 2#8, 2#9, 8#10, 3^11, 11#12, 11#13, 13#14, 2^15, 15#16, 4^17",
			"1",
			"4",
			5,
			1,
		},

		/*  Expected Result
		Blkgheight 				MAIN FORK:		     	ORPHANS:
			 0
			 1
			 2
			 3					10
			 4
			 5
		*/
		{
			"Set rootBlkHash to a descendant in main fork bottom section",
			"1#2, 1#3, 3#4, 4#5, 4#6, 4#7, 2#8, 2#9, 8#10, 3^11, 11#12, 11#13, 13#14, 2^15, 15#16, 4^17",
			"1",
			"10",
			1,
			0,
		},

		/*  Expected Result
		Blkgheight 				MAIN FORK:		     	ORPHANS:(3 orphan forks)
			 0
			 1
			 2
			 3								                11
			 4											  12  13
			 5													14
		*/
		{
			"Set rootBlkHash to the root in orphan fork 1",
			"1#2, 1#3, 3#4, 4#5, 4#6, 4#7, 2#8, 2#9, 8#10, 3^11, 11#12, 11#13, 13#14, 2^15, 15#16, 4^17",
			"1",
			"11",
			4,
			0,
		},

		/*  Expected Result
		Blkgheight 				MAIN FORK:		     	ORPHANS:(3 orphan forks)
			 0
			 1
			 2
			 3
			 4											      13
			 5													14
		*/
		{
			"Set rootBlkHash to a node in orphan fork 1",
			"1#2, 1#3, 3#4, 4#5, 4#6, 4#7, 2#8, 2#9, 8#10, 3^11, 11#12, 11#13, 13#14, 2^15, 15#16, 4^17",
			"1",
			"13",
			2,
			0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp, _ := DeserializeBlockPool(tt.serializedBp, tt.rootBlkHash, 0)
			node, ok := bp.blkCache.Get(hash.Hash(tt.newRootBlkHash).String())
			assert.True(t, ok)
			bp.UpdateRootBlock(node.(*common.TreeNode).GetValue().(*block.Block))
			assert.Equal(t, tt.expectedNumOfNodes, bp.blkCache.Len())
			assert.Equal(t, node, bp.root)
		})
	}
}

func testGetForkHeadHashes(bp *BlockPool) []string {
	var hashes []string
	bp.ForkHeadRange(func(blkHash string, tree *common.TreeNode) {
		hashes = append(hashes, blkHash)
	})
	return hashes
}
