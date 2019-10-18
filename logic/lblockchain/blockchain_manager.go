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
	"github.com/pkg/errors"

	"github.com/dappley/go-dappley/common/hash"
	"github.com/dappley/go-dappley/common/pubsub"
	"github.com/dappley/go-dappley/core/block"
	blockpb "github.com/dappley/go-dappley/core/block/pb"
	"github.com/dappley/go-dappley/core/blockchain"
	"github.com/dappley/go-dappley/logic/lblock"

	lblockchainpb "github.com/dappley/go-dappley/logic/lblockchain/pb"
	"github.com/dappley/go-dappley/network/networkmodel"
	"github.com/golang/protobuf/proto"
	logger "github.com/sirupsen/logrus"
)

const (
	HeightDiffThreshold = 10
	SendBlock           = "SendBlockByHash"
	RequestBlock        = "requestBlock"
)

var (
	bmSubscribedTopics = []string{
		SendBlock,
		RequestBlock,
	}
)

var (
	ErrParentBlockNotFound = errors.New("Not able to find parent block in blockchain")
)

type BlockchainManager struct {
	blockchain        *Blockchain
	blockPool         *blockchain.BlockPool
	consensus         Consensus
	downloadRequestCh chan chan bool
	netService        NetService
}

func NewBlockchainManager(blockchain *Blockchain, blockpool *blockchain.BlockPool, service NetService, consensus Consensus) *BlockchainManager {
	bm := &BlockchainManager{
		blockchain: blockchain,
		blockPool:  blockpool,
		netService: service,
		consensus:  consensus,
	}
	bm.ListenToNetService()
	return bm
}

func (bm *BlockchainManager) SetDownloadRequestCh(requestCh chan chan bool) {
	bm.downloadRequestCh = requestCh
}

func (bm *BlockchainManager) RequestDownloadBlockchain() {
	go func() {
		finishChan := make(chan bool, 1)

		bm.Getblockchain().SetState(blockchain.BlockchainDownloading)

		select {
		case bm.downloadRequestCh <- finishChan:
		default:
			logger.Warn("BlockchainManager: Request download failed! download request channel is full!")
		}

		<-finishChan
		bm.Getblockchain().SetState(blockchain.BlockchainReady)
	}()
}

func (bm *BlockchainManager) ListenToNetService() {
	if bm.netService == nil {
		return
	}

	bm.netService.Listen(bm)
}

func (bm *BlockchainManager) GetSubscribedTopics() []string {
	return bmSubscribedTopics
}

func (bm *BlockchainManager) GetTopicHandler(topic string) pubsub.TopicHandler {

	switch topic {
	case SendBlock:
		return bm.SendBlockHandler
	case RequestBlock:
		return bm.RequestBlockHandler
	}
	return nil
}

func (bm *BlockchainManager) Getblockchain() *Blockchain {
	return bm.blockchain
}

func (bm *BlockchainManager) GetblockPool() *blockchain.BlockPool {
	return bm.blockchain.bc.forks
}

func (bm *BlockchainManager) VerifyBlock(blk *block.Block) bool {
	if !lblock.VerifyHash(blk) {
		logger.Warn("BlockchainManager: Block hash verification failed!")
		return false
	}
	//TODO: Verify double spending transactions in the same blk
	if !(bm.consensus.Validate(blk)) {
		logger.Warn("BlockchainManager: blk is invalid according to libPolicy!")
		return false
	}
	logger.Debug("BlockchainManager: blk is valid according to libPolicy.")
	return true
}

func (bm *BlockchainManager) Push(blk *block.Block, pid networkmodel.PeerInfo) {
	logger.WithFields(logger.Fields{
		"from":   pid.PeerId.String(),
		"hash":   blk.GetHash().String(),
		"height": blk.GetHeight(),
	}).Info("BlockChainManager: received a new block.")

	if bm.blockchain.GetState() != blockchain.BlockchainReady {
		logger.Info("BlockchainManager: Blockchain not ready, discard received blk")
		return
	}
	if !bm.VerifyBlock(blk) {
		return
	}

	receiveBlockHeight := blk.GetHeight()
	ownBlockHeight := bm.Getblockchain().GetMaxHeight()
	if receiveBlockHeight-ownBlockHeight >= HeightDiffThreshold &&
		bm.blockchain.GetState() == blockchain.BlockchainReady {
		logger.Info("The height of the received blk is higher than the height of its own blk,to start download blockchain")
		bm.RequestDownloadBlockchain()
		return
	}

	bm.blockPool.AddBlock(blk)
	forkHeadBlk := bm.blockPool.GetForkHead(blk)
	if forkHeadBlk == nil {
		return
	}

	if !bm.blockchain.IsInBlockchain(forkHeadBlk.GetPrevHash()) {
		logger.WithFields(logger.Fields{
			"parent_hash": forkHeadBlk.GetPrevHash(),
			"from":        pid,
		}).Info("BlockchainManager: cannot find the parent of the received blk from blockchain. Requesting the parent...")
		bm.RequestBlock(forkHeadBlk.GetPrevHash(), pid)
		return
	}

	fork := bm.blockPool.GetFork(forkHeadBlk.GetPrevHash())
	if fork == nil {
		return
	}

	if fork[0].GetHeight() <= bm.Getblockchain().GetMaxHeight() {
		return
	}

	bm.blockchain.SetState(blockchain.BlockchainSync)
	_ = bm.blockchain.SwitchFork(fork, forkHeadBlk.GetPrevHash())
	bm.blockPool.RemoveFork(fork)
	bm.blockchain.SetState(blockchain.BlockchainReady)
	return
}

//RequestBlock sends a requestBlock command to its peer with pid through network module
func (bm *BlockchainManager) RequestBlock(hash hash.Hash, pid networkmodel.PeerInfo) {
	request := &lblockchainpb.RequestBlock{Hash: hash}

	bm.netService.UnicastHighProrityCommand(RequestBlock, request, pid)
}

//RequestBlockhandler handles when blockchain manager receives a requestBlock command from its peers
func (bm *BlockchainManager) RequestBlockHandler(input interface{}) {

	var command *networkmodel.DappRcvdCmdContext
	command = input.(*networkmodel.DappRcvdCmdContext)

	request := &lblockchainpb.RequestBlock{}

	if err := proto.Unmarshal(command.GetData(), request); err != nil {
		logger.WithFields(logger.Fields{
			"name": command.GetCommandName(),
		}).Info("BlockchainManager: parse data failed.")
	}

	block, err := bm.Getblockchain().GetBlockByHash(request.Hash)
	if err != nil {
		logger.WithError(err).Warn("BlockchainManager: failed to get the requested block.")
		return
	}

	bm.SendBlockToPeer(block, command.GetSource())
}

//SendBlockToPeer unicasts a block to the peer with peer id "pid"
func (bm *BlockchainManager) SendBlockToPeer(block *block.Block, pid networkmodel.PeerInfo) {

	bm.netService.UnicastNormalPriorityCommand(SendBlock, block.ToProto(), pid)
}

//BroadcastBlock broadcasts a block to all peers
func (bm *BlockchainManager) BroadcastBlock(block *block.Block) {
	bm.netService.BroadcastHighProrityCommand(SendBlock, block.ToProto())
}

//SendBlockHandler handles when blockchain manager receives a sendBlock command from its peers
func (bm *BlockchainManager) SendBlockHandler(input interface{}) {

	var command *networkmodel.DappRcvdCmdContext
	command = input.(*networkmodel.DappRcvdCmdContext)

	blockpb := &blockpb.Block{}

	//unmarshal byte to proto
	if err := proto.Unmarshal(command.GetData(), blockpb); err != nil {
		logger.WithError(err).Warn("BlockchainManager: parse data failed.")
		return
	}

	blk := &block.Block{}
	blk.FromProto(blockpb)
	bm.Push(blk, command.GetSource())

	if command.IsBroadcast() {
		//relay the original command
		bm.netService.Relay(command.GetCommand(), networkmodel.PeerInfo{}, networkmodel.HighPriorityCommand)
	}
}

/* NumForks returns the number of forks in the BlockPool and the height of the current longest fork */
func (bm *BlockchainManager) NumForks() (int64, int64) {

	numOfForks := bm.blockchain.bc.GetNumOfForks()
	maxHeight := bm.blockchain.bc.GetMaxHeight()

	return numOfForks, int64(maxHeight)
}
