// Copyright (C) 2018 go-dappley authors
//
// This file is part of the go-dappley library.
//
// the go-dappley library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-dappley library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-dappley library.  If not, see <http://www.gnu.org/licenses/>.
//
package rpc

import (
	"context"
	"strings"

	"github.com/dappley/go-dappley/client"
	"github.com/dappley/go-dappley/crypto/byteutils"

	"github.com/golang/protobuf/proto"
	logger "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/dappley/go-dappley/common"
	"github.com/dappley/go-dappley/core"
	corepb "github.com/dappley/go-dappley/core/pb"
	"github.com/dappley/go-dappley/logic"
	"github.com/dappley/go-dappley/network"
	rpcpb "github.com/dappley/go-dappley/rpc/pb"
	"github.com/dappley/go-dappley/vm"
)

const (
	ProtoVersion                   = "1.0.0"
	MaxGetBlocksCount       int32  = 500
	MinUtxoBlockHeaderCount uint64 = 6
)

type RpcService struct {
	node *network.Node
}

func (rpcService *RpcService) RpcGetVersion(ctx context.Context, in *rpcpb.GetVersionRequest) (*rpcpb.GetVersionResponse, error) {
	clientProtoVersions := strings.Split(in.GetProtoVersion(), ".")

	if len(clientProtoVersions) != 3 {
		return nil, status.Error(codes.InvalidArgument, "proto version not supported")
	}

	serverProtoVersions := strings.Split(ProtoVersion, ".")

	// Major version must equal
	if serverProtoVersions[0] != clientProtoVersions[0] {
		return nil, status.Error(codes.Unimplemented, "major version mismatch")
	}

	return &rpcpb.GetVersionResponse{ProtoVersion: ProtoVersion, ServerVersion: ""}, nil
}

func (rpcService *RpcService) RpcGetBalance(ctx context.Context, in *rpcpb.GetBalanceRequest) (*rpcpb.GetBalanceResponse, error) {
	address := in.GetAddress()
	if !client.NewAddress(address).IsValid() {
		return nil, status.Error(codes.InvalidArgument, client.ErrInvalidAddress.Error())
	}

	amount, err := logic.GetBalance(client.NewAddress(address), rpcService.node.GetBlockchain())
	if err != nil {
		switch err {
		case logic.ErrInvalidAddress:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Unknown, err.Error())
		}
	}
	return &rpcpb.GetBalanceResponse{Amount: amount.Int64()}, nil
}

func (rpcService *RpcService) RpcGetBlockchainInfo(ctx context.Context, in *rpcpb.GetBlockchainInfoRequest) (*rpcpb.GetBlockchainInfoResponse, error) {
	tailBlock, err := rpcService.node.GetBlockchain().GetTailBlock()
	if err != nil {
		switch err {
		case core.ErrBlockDoesNotExist:
			return nil, status.Error(codes.Internal, err.Error())
		default:
			return nil, status.Error(codes.Unknown, err.Error())
		}
	}

	return &rpcpb.GetBlockchainInfoResponse{
		TailBlockHash: rpcService.node.GetBlockchain().GetTailBlockHash(),
		BlockHeight:   rpcService.node.GetBlockchain().GetMaxHeight(),
		Producers:     rpcService.node.GetBlockchain().GetConsensus().GetProducers(),
		Timestamp:     tailBlock.GetTimestamp(),
	}, nil
}

func (rpcService *RpcService) RpcGetUTXO(ctx context.Context, in *rpcpb.GetUTXORequest) (*rpcpb.GetUTXOResponse, error) {
	utxoIndex := core.NewUTXOIndex(rpcService.node.GetBlockchain().GetUtxoCache())
	utxoIndex.UpdateUtxoState(rpcService.node.GetBlockchain().GetTxPool().GetAllTransactions())

	publicKeyHash, ok := client.NewAddress(in.GetAddress()).GetPubKeyHash()
	if !ok {
		return nil, status.Error(codes.InvalidArgument, logic.ErrInvalidAddress.Error())
	}

	utxos := utxoIndex.GetAllUTXOsByPubKeyHash(publicKeyHash)
	response := rpcpb.GetUTXOResponse{}
	for _, utxo := range utxos.Indices {
		response.Utxos = append(response.Utxos, utxo.ToProto().(*corepb.Utxo))
	}

	//TODO Race condition Blockchain update after GetUTXO
	getHeaderCount := MinUtxoBlockHeaderCount
	if int(getHeaderCount) < len(rpcService.node.GetBlockchain().GetConsensus().GetProducers()) {
		getHeaderCount = uint64(len(rpcService.node.GetBlockchain().GetConsensus().GetProducers()))
	}

	tailHeight := rpcService.node.GetBlockchain().GetMaxHeight()
	if getHeaderCount > tailHeight {
		getHeaderCount = tailHeight
	}

	for i := uint64(0); i < getHeaderCount; i++ {
		block, err := rpcService.node.GetBlockchain().GetBlockByHeight(tailHeight - uint64(i))
		if err != nil {
			break
		}

		response.BlockHeaders = append(response.BlockHeaders, block.GetHeader().ToProto().(*corepb.BlockHeader))
	}

	return &response, nil
}

// RpcGetBlocks Get blocks in blockchain from head to tail
func (rpcService *RpcService) RpcGetBlocks(ctx context.Context, in *rpcpb.GetBlocksRequest) (*rpcpb.GetBlocksResponse, error) {
	block := rpcService.findBlockInRequestHash(in.GetStartBlockHashes())

	// Reach the blockchain's tail
	if block.GetHeight() >= rpcService.node.GetBlockchain().GetMaxHeight() {
		return &rpcpb.GetBlocksResponse{}, nil
	}

	var blocks []*core.Block
	maxBlockCount := in.GetMaxCount()
	if maxBlockCount > MaxGetBlocksCount {
		return nil, status.Error(codes.InvalidArgument, "block count overflow")
	}

	block, err := rpcService.node.GetBlockchain().GetBlockByHeight(block.GetHeight() + 1)
	for i := int32(0); i < maxBlockCount && err == nil; i++ {
		blocks = append(blocks, block)
		block, err = rpcService.node.GetBlockchain().GetBlockByHeight(block.GetHeight() + 1)
	}

	result := &rpcpb.GetBlocksResponse{}

	for _, block = range blocks {
		result.Blocks = append(result.Blocks, block.ToProto().(*corepb.Block))
	}

	return result, nil
}

func (rpcService *RpcService) findBlockInRequestHash(startBlockHashes [][]byte) *core.Block {
	for _, hash := range startBlockHashes {
		// hash in blockchain, return
		if block, err := rpcService.node.GetBlockchain().GetBlockByHash(hash); err == nil {
			return block
		}
	}

	// Return Genesis Block
	block, _ := rpcService.node.GetBlockchain().GetBlockByHeight(0)
	return block
}

// RpcGetBlockByHash Get single block in blockchain by hash
func (rpcService *RpcService) RpcGetBlockByHash(ctx context.Context, in *rpcpb.GetBlockByHashRequest) (*rpcpb.GetBlockByHashResponse, error) {
	block, err := rpcService.node.GetBlockchain().GetBlockByHash(in.GetHash())

	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &rpcpb.GetBlockByHashResponse{Block: block.ToProto().(*corepb.Block)}, nil
}

// RpcGetBlockByHeight Get single block in blockchain by height
func (rpcService *RpcService) RpcGetBlockByHeight(ctx context.Context, in *rpcpb.GetBlockByHeightRequest) (*rpcpb.GetBlockByHeightResponse, error) {
	block, err := rpcService.node.GetBlockchain().GetBlockByHeight(in.GetHeight())

	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &rpcpb.GetBlockByHeightResponse{Block: block.ToProto().(*corepb.Block)}, nil
}

// RpcSendTransaction Send transaction to blockchain created by account client
func (rpcService *RpcService) RpcSendTransaction(ctx context.Context, in *rpcpb.SendTransactionRequest) (*rpcpb.SendTransactionResponse, error) {
	tx := core.Transaction{nil, nil, nil, common.NewAmount(0), common.NewAmount(0), common.NewAmount(0)}
	tx.FromProto(in.GetTransaction())

	if tx.IsCoinbase() {
		return nil, status.Error(codes.InvalidArgument, "cannot send coinbase transaction")
	}

	utxoIndex := core.NewUTXOIndex(rpcService.node.GetBlockchain().GetUtxoCache())
	utxoIndex.UpdateUtxoState(rpcService.node.GetBlockchain().GetTxPool().GetAllTransactions())

	if result, err := tx.Verify(utxoIndex, 0); !result {
		logger.Warn(err.Error())
		return nil, status.Error(codes.FailedPrecondition, core.ErrTransactionVerifyFailed.Error())
	}

	engine := vm.NewV8Engine()
	defer engine.DestroyEngine()
	err := tx.CheckContractSyntax(engine)

	if err != nil {
		logger.WithFields(logger.Fields{
			"error": err,
		}).Error("Smart Contract Deployed Failed!")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	rpcService.node.GetBlockchain().GetTxPool().Push(tx)
	rpcService.node.TxBroadcast(&tx)
	if tx.IsContract() {
		contractAddr := tx.GetContractAddress()
		message := contractAddr.String()
		logger.WithFields(logger.Fields{
			"contractAddr": message,
		}).Info("Smart Contract Deployed Successful!")
	}

	return &rpcpb.SendTransactionResponse{}, nil
}

// RpcSendBatchTransaction sends a batch of transactions to blockchain created by account client
func (rpcService *RpcService) RpcSendBatchTransaction(ctx context.Context, in *rpcpb.SendBatchTransactionRequest) (*rpcpb.SendBatchTransactionResponse, error) {
	statusCode := codes.OK
	var details []proto.Message
	utxoIndex := core.NewUTXOIndex(rpcService.node.GetBlockchain().GetUtxoCache())
	utxoIndex.UpdateUtxoState(rpcService.node.GetBlockchain().GetTxPool().GetAllTransactions())

	txMap := make(map[int]core.Transaction, len(in.Transactions))
	txs := []core.Transaction{}
	for key, txInReq := range in.Transactions {
		tx := core.Transaction{nil, nil, nil, common.NewAmount(0), common.NewAmount(0), common.NewAmount(0)}
		tx.FromProto(txInReq)
		txs = append(txs, tx)
		txMap[key] = tx
	}

	// verify dependent transactions within batch of transactions
	lastTxsLen := 0
	verifiedTxs := []core.Transaction{}
	for len(txMap) != lastTxsLen {
		lastTxsLen = len(txMap)
		for key, tx := range txs {
			if _, ok := txMap[key]; !ok {
				continue
			}

			if tx.IsCoinbase() {
				if statusCode == codes.OK {
					statusCode = codes.Unknown
				}
				details = append(details, &rpcpb.SendTransactionStatus{
					Txid:    tx.ID,
					Code:    uint32(codes.InvalidArgument),
					Message: "cannot send coinbase transaction",
				})
				delete(txMap, key)
				continue
			}

			if result, _ := tx.Verify(utxoIndex, 0); !result {
				continue
			}

			utxoIndex.UpdateUtxo(&tx)
			rpcService.node.GetBlockchain().GetTxPool().Push(tx)
			verifiedTxs = append(verifiedTxs, tx)

			details = append(details, &rpcpb.SendTransactionStatus{
				Txid:    tx.ID,
				Code:    uint32(codes.OK),
				Message: "",
			})
			delete(txMap, key)
		}
	}

	rpcService.node.BatchTxBroadcast(verifiedTxs)

	st := status.New(codes.OK, "")
	// add invalid transactions to response details if exists
	if statusCode == codes.Unknown || len(txMap) > 0 {
		st = status.New(codes.Unknown, "one or more transactions are invalid")
		for _, tx := range txMap {
			details = append(details, &rpcpb.SendTransactionStatus{
				Txid:    tx.ID,
				Code:    uint32(codes.FailedPrecondition),
				Message: core.ErrTransactionVerifyFailed.Error(),
			})

		}
	}
	st, _ = st.WithDetails(details...)

	return &rpcpb.SendBatchTransactionResponse{}, st.Err()
}

func (rpcService *RpcService) RpcGetNewTransaction(in *rpcpb.GetNewTransactionRequest, stream rpcpb.RpcService_RpcGetNewTransactionServer) error {
	var txHandler interface{}

	quitCh := make(chan bool, 1)

	txHandler = func(tx *core.Transaction) {
		response := &rpcpb.GetNewTransactionResponse{Transaction: tx.ToProto().(*corepb.Transaction)}
		err := stream.Send(response)
		if err != nil {
			logger.WithError(err).Info("RPCService: failed to send transaction to client.")
			rpcService.node.GetBlockchain().GetTxPool().EventBus.Unsubscribe(core.NewTransactionTopic, txHandler)
			quitCh <- true
		}
	}

	rpcService.node.GetBlockchain().GetTxPool().EventBus.SubscribeAsync(core.NewTransactionTopic, txHandler, false)
	<-quitCh
	return nil
}

func (rpcService *RpcService) RpcSubscribe(in *rpcpb.SubscribeRequest, stream rpcpb.RpcService_RpcSubscribeServer) error {
	quitCh := make(chan bool, 1)
	var cb interface{}
	cb = func(event *core.Event) {
		response := &rpcpb.SubscribeResponse{Data: event.GetData()}
		err := stream.Send(response)
		if err != nil {
			logger.WithError(err).WithFields(logger.Fields{
				"topic": event.GetTopic(),
				"data":  event.GetData(),
			}).Info("RPCService: failed to send published data")
			rpcService.node.GetBlockchain().GetEventManager().Unsubscribe(event.GetTopic(), cb)
			quitCh <- true
		}
	}
	rpcService.node.GetBlockchain().GetEventManager().SubscribeMultiple(in.Topics, cb)
	<-quitCh
	return nil
}

func (rpcService *RpcService) IsPrivate() bool { return false }

// RpcGetAllTransactionsFromTxPool get all transactions from transaction_pool
func (rpcService *RpcService) RpcGetAllTransactionsFromTxPool(ctx context.Context, in *rpcpb.GetAllTransactionsRequest) (*rpcpb.GetAllTransactionsResponse, error) {
	txs := rpcService.node.GetBlockchain().GetTxPool().GetTransactions()
	result := &rpcpb.GetAllTransactionsResponse{}
	for _, tx := range txs {
		result.Transactions = append(result.Transactions, tx.ToProto().(*corepb.Transaction))
	}
	return result, nil
}

// RpcGetLastIrreversibleBlock get last irreversible block
func (rpcService *RpcService) RpcGetLastIrreversibleBlock(ctx context.Context, in *rpcpb.GetLastIrreversibleBlockRequest) (*rpcpb.GetLastIrreversibleBlockResponse, error) {
	block, err := rpcService.node.GetBlockchain().GetLIB()

	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &rpcpb.GetLastIrreversibleBlockResponse{Block: block.ToProto().(*corepb.Block)}, nil
}

// RpcEstimateGas estimate gas value of contract deploy and execution.
func (rpcService *RpcService) RpcEstimateGas(ctx context.Context, in *rpcpb.EstimateGasRequest) (*rpcpb.EstimateGasResponse, error) {
	tx := core.Transaction{nil, nil, nil, common.NewAmount(0), common.NewAmount(0), common.NewAmount(0)}
	tx.FromProto(in.GetTransaction())

	if tx.IsCoinbase() {
		return nil, status.Error(codes.InvalidArgument, "cannot send coinbase transaction")
	}
	contractTx := tx.ToContractTx()
	if contractTx == nil {
		return nil, status.Error(codes.FailedPrecondition, "cannot estimate normal transaction")
	}
	utxoIndex := core.NewUTXOIndex(rpcService.node.GetBlockchain().GetUtxoCache())
	utxoIndex.UpdateUtxoState(rpcService.node.GetBlockchain().GetTxPool().GetTransactions())

	err := contractTx.VerifyInEstimate(utxoIndex)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	tx.GasLimit = common.NewAmount(vm.MaxLimitsOfExecutionInstructions)
	gasCount, error := vm.EstimateGas(rpcService.node.GetBlockchain(), &tx)
	return &rpcpb.EstimateGasResponse{GasCount: byteutils.FromUint64(gasCount)}, error
}

// RpcGasPrice returns current gas price.
func (rpcService *RpcService) RpcGasPrice(ctx context.Context, in *rpcpb.GasPriceRequest) (*rpcpb.GasPriceResponse, error) {
	gasPrice := rpcService.node.GetBlockchain().GasPrice()
	return &rpcpb.GasPriceResponse{GasPrice: byteutils.FromUint64(gasPrice)}, nil
}
