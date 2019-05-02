package sdk

import (
	"context"
	"github.com/dappley/go-dappley/common"
	"github.com/dappley/go-dappley/core"
	"github.com/dappley/go-dappley/core/pb"
	"github.com/dappley/go-dappley/rpc/pb"
	logger "github.com/sirupsen/logrus"
)

type DappSdk struct {
	conn *DappSdkGrpcClient
}

//NewDappSdk creates a new DappSdk instance
func NewDappSdk(conn *DappSdkGrpcClient) *DappSdk {
	return &DappSdk{conn}
}

//GetBlockHeight requests the height of currnet tail block from the server
func (sdk *DappSdk) GetBlockHeight() (uint64, error) {
	resp, err := sdk.conn.rpcClient.RpcGetBlockchainInfo(
		context.Background(),
		&rpcpb.GetBlockchainInfoRequest{},
	)

	if err != nil || resp == nil {
		return 0, err
	}

	return resp.BlockHeight, nil
}

//GetBlockHeight requests the balance of the input address from the server
func (sdk *DappSdk) GetBalance(address string) (int64, error) {
	response, err := sdk.conn.rpcClient.RpcGetBalance(context.Background(), &rpcpb.GetBalanceRequest{Address: address})
	if err != nil {
		return 0, err
	}
	return response.Amount, err
}

//SendBatchTransactions sends a batch of transactions to the server
func (sdk *DappSdk) SendBatchTransactions(txs []*corepb.Transaction) error {
	_, err := sdk.conn.rpcClient.RpcSendBatchTransaction(
		context.Background(),
		&rpcpb.SendBatchTransactionRequest{
			Transactions: txs,
		},
	)

	if err != nil {
		return err
	}

	logger.WithFields(logger.Fields{
		"num_of_txs": len(txs),
	}).Info("DappSDK: Batch Transactions are sent!")

	return nil
}

//RequestFund sends a fund request to the server
func (sdk *DappSdk) RequestFund(fundAddr string, amount *common.Amount) {
	sendFromMinerRequest := &rpcpb.SendFromMinerRequest{To: fundAddr, Amount: amount.Bytes()}
	sdk.conn.adminClient.RpcSendFromMiner(context.Background(), sendFromMinerRequest)
}

//GetUtxoByAddr gets all utxos related to an address from the server
func (sdk *DappSdk) GetUtxoByAddr(addr core.Address) ([]*corepb.Utxo, error) {

	resp, err := sdk.conn.rpcClient.RpcGetUTXO(context.Background(), &rpcpb.GetUTXORequest{
		Address: addr.String(),
	})

	if err != nil || resp == nil {
		return nil, err
	}

	return resp.Utxos, nil
}
