package util

import (
	"github.com/dappley/go-dappley/common"
	"github.com/dappley/go-dappley/core/account"
	"github.com/dappley/go-dappley/core/transaction"
	transactionpb "github.com/dappley/go-dappley/core/transaction/pb"
	"github.com/dappley/go-dappley/sdk"
	logger "github.com/sirupsen/logrus"
)

type InsufficientBalanceTxSender struct {
	TxSender
}

func NewInsufficientBalanceTxSender(dappSdk *sdk.DappSdk, account *sdk.DappSdkAccount) *InsufficientBalanceTxSender {
	return &InsufficientBalanceTxSender{
		TxSender{
			dappSdk: dappSdk,
			account: account,
		},
	}
}

func (txSender *InsufficientBalanceTxSender) Generate(params transaction.SendTxParam) {
	pkh, err := account.NewUserPubKeyHash(params.SenderKeyPair.GetPublicKey())

	if err != nil {
		logger.WithError(err).Panic("InsufficientBalanceTx: Unable to hash sender public key")
	}

	prevUtxos, err := txSender.account.GetUtxoIndex().GetUTXOsByAmount(pkh, params.Amount)

	if err != nil {
		logger.WithError(err).Panic("InsufficientBalanceTx: Unable to get UTXOs to match the amount")
	}

	vouts := prepareOutputLists(prevUtxos, params.From, params.To, params.Amount, params.Tip)
	vouts[0].Value = vouts[0].Value.Add(common.NewAmount(1))

	txSender.tx = NewTransaction(prevUtxos, vouts, params.Tip, params.SenderKeyPair)
}

func (txSender *InsufficientBalanceTxSender) Send() {

	_, err := txSender.dappSdk.SendTransaction(txSender.tx.ToProto().(*transactionpb.Transaction))

	if err != nil {
		logger.WithError(err).Error("InsufficientBalanceTx: Sending transaction failed!")
	}
}

func (txSender *InsufficientBalanceTxSender) Print() {
	logger.Info("Sending a transaction with insufficient balance...")
}
