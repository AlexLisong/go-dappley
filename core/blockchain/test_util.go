package blockchain

import (
	"github.com/dappley/go-dappley/common/hash"
	"github.com/dappley/go-dappley/core/block"
	logger "github.com/sirupsen/logrus"
	"strconv"
)

func CreateBlock(currentHash hash.Hash, prevHash hash.Hash, height uint64) *block.Block {
	return block.NewBlockWithRawInfo(currentHash, prevHash, 0, 0, height, nil)
}

//DeserializeBlockPool creates a block pool by deserializing the input string. return the root of the tree
func DeserializeBlockPool(s string, rootBlkHash string, rootBlkHeight uint64) (*BlockPool, map[string]*block.Block) {
	/* "0^1, 1#2, 1#3, 3#4, 0^5, 1^6" describes a block pool like following"
				1      5
			   2 3			6
	              4
	*/
	if rootBlkHash == "" {
		return NewBlockPool(nil), nil
	}

	s += ","
	rootBlk := CreateBlock(hash.Hash(rootBlkHash), nil, rootBlkHeight)
	bp := NewBlockPool(rootBlk)

	var parentBlk *block.Block
	currStr := ""
	parentBlkHash := ""
	blkHeight := -1
	blocks := make(map[string]*block.Block)
	blocks[hash.Hash(rootBlkHash).String()] = rootBlk

	for _, c := range s {
		switch c {
		case ',':

			if currStr == rootBlkHash {
				currStr = ""
				continue
			}

			var blk *block.Block
			if parentBlk == nil {
				blk = CreateBlock(hash.Hash(currStr), hash.Hash(parentBlkHash), uint64(blkHeight))
			} else {
				if blkHeight == -1 {
					blkHeight = int(parentBlk.GetHeight() + 1)
				}
				blk = CreateBlock(hash.Hash(currStr), parentBlk.GetHash(), uint64(blkHeight))
			}
			bp.AddBlock(blk)
			blocks[hash.Hash(currStr).String()] = blk
			if parentBlk == nil {
				logger.WithFields(logger.Fields{
					"hash":   hash.Hash(currStr).String(),
					"height": blk.GetHeight(),
				}).Debug("Add a new head block")
			} else {
				logger.WithFields(logger.Fields{
					"hash":   hash.Hash(currStr).String(),
					"parent": parentBlk.GetHash().String(),
					"height": blk.GetHeight(),
				}).Debug("Add a new block")
			}
			currStr = ""
			parentBlkHash = ""
			parentBlk = nil
			blkHeight = -1
		case '#':
			parentBlkHash = currStr
			if _, isFound := blocks[hash.Hash(currStr).String()]; !isFound {
				currStr = ""
				continue
			}
			parentBlk = blocks[hash.Hash(currStr).String()]
			currStr = ""
		case '^':
			num, err := strconv.Atoi(currStr)
			if err != nil {
				logger.WithError(err).Panic("deserialize block pool failed while converting string to int")
			}
			blkHeight = num
			currStr = ""
		case ' ':
			continue
		default:
			currStr = currStr + string(c)
		}
	}

	return bp, blocks
}
