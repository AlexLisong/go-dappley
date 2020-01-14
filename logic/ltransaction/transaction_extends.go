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

package ltransaction

import (
	"crypto/ecdsa"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/dappley/go-dappley/common"
	"github.com/dappley/go-dappley/core/account"
	"github.com/dappley/go-dappley/core/transaction"
	"github.com/dappley/go-dappley/core/transactionbase"
	"github.com/dappley/go-dappley/core/utxo"
	"github.com/dappley/go-dappley/logic/lutxo"
	logger "github.com/sirupsen/logrus"
	"strings"
	"time"
)

const (
	scheduleFuncName = "dapp_schedule"
)

// Normal transaction
type TxNormal struct {
	*transaction.Transaction
}

// TxContract contains contract value
type TxContract struct {
	*transaction.Transaction
	Address account.Address
}

// Coinbase transaction, rewards to miner
type TxCoinbase struct {
	*transaction.Transaction
}

// GasReward transaction, rewards to miner during vm execution
type TxGasReward struct {
	*transaction.Transaction
}

// GasChange transaction, change value to from user
type TxGasChange struct {
	*transaction.Transaction
}

// Reward transaction, step reward
type TxReward struct {
	*transaction.Transaction
}

// Returns decorator of transaction
func NewTxDecorator(tx *transaction.Transaction) TxDecorator {
	// old data adapter
	adaptedTx := transaction.NewTxAdapter(tx)
	tx = adaptedTx.Transaction
	switch tx.Type {
	case transaction.TxTypeNormal:
		return &TxNormal{tx}
	case transaction.TxTypeContract:
		return NewTxContract(tx)
	case transaction.TxTypeCoinbase:
		return &TxCoinbase{tx}
	case transaction.TxTypeGasReward:
		return &TxGasReward{tx}
	case transaction.TxTypeGasChange:
		return &TxGasChange{tx}
	case transaction.TxTypeReward:
		return &TxReward{tx}
	}
	return nil
}

func (tx *TxNormal) Sign(privKey ecdsa.PrivateKey, prevUtxos []*utxo.UTXO) error {
	return tx.Transaction.Sign(privKey, prevUtxos)
}

func (tx *TxNormal) IsNeedVerify() bool {
	return true
}

func (tx *TxNormal) Verify(utxoIndex *lutxo.UTXOIndex, blockHeight uint64) error {
	prevUtxos, err := lutxo.FindVinUtxosInUtxoPool(utxoIndex, tx.Transaction)
	if err != nil {
		logger.WithError(err).WithFields(logger.Fields{
			"txid": hex.EncodeToString(tx.ID),
		}).Warn("Verify: cannot find vin while verifying normal tx")
		return err
	}
	return tx.Transaction.Verify(prevUtxos)
}

func (tx *TxContract) Sign(privKey ecdsa.PrivateKey, prevUtxos []*utxo.UTXO) error {
	return tx.Transaction.Sign(privKey, prevUtxos)
}

func (tx *TxContract) IsNeedVerify() bool {
	return true
}

func (tx *TxContract) Verify(utxoIndex *lutxo.UTXOIndex, blockHeight uint64) error {
	prevUtxos, err := lutxo.FindVinUtxosInUtxoPool(utxoIndex, tx.Transaction)
	if err != nil {
		return err
	}
	err = tx.verifyInEstimate(utxoIndex, prevUtxos)
	if err != nil {
		return err
	}
	totalBalance, err := tx.GetTotalBalance(prevUtxos)
	if err != nil {
		return err
	}
	return tx.VerifyGas(totalBalance)
}

func (tx *TxCoinbase) Sign(privKey ecdsa.PrivateKey, prevUtxos []*utxo.UTXO) error {
	return nil
}

func (tx *TxCoinbase) IsNeedVerify() bool {
	return true
}

func (tx *TxCoinbase) Verify(utxoIndex *lutxo.UTXOIndex, blockHeight uint64) error {
	//TODO coinbase vout check need add tip
	if tx.Vout[0].Value.Cmp(transaction.Subsidy) < 0 {
		return errors.New("Transaction: subsidy check failed")
	}
	bh := binary.BigEndian.Uint64(tx.Vin[0].Signature)
	if blockHeight != bh {
		return fmt.Errorf("Transaction: block height check failed expected=%v actual=%v", blockHeight, bh)
	}
	return nil
}

func (tx *TxGasReward) Sign(privKey ecdsa.PrivateKey, prevUtxos []*utxo.UTXO) error {
	return nil
}

func (tx *TxGasReward) IsNeedVerify() bool {
	return false
}

func (tx *TxGasReward) Verify(utxoIndex *lutxo.UTXOIndex, blockHeight uint64) error {
	return nil
}

func (tx *TxGasChange) Sign(privKey ecdsa.PrivateKey, prevUtxos []*utxo.UTXO) error {
	return nil
}

func (tx *TxGasChange) IsNeedVerify() bool {
	return false
}

func (tx *TxGasChange) Verify(utxoIndex *lutxo.UTXOIndex, blockHeight uint64) error {
	return nil
}

func (tx *TxReward) Sign(privKey ecdsa.PrivateKey, prevUtxos []*utxo.UTXO) error {
	return nil
}

func (tx *TxReward) IsNeedVerify() bool {
	return false
}

func (tx *TxReward) Verify(utxoIndex *lutxo.UTXOIndex, blockHeight uint64) error {
	return nil
}

func NewTxContract(tx *transaction.Transaction) *TxContract {
	adaptedTx := transaction.NewTxAdapter(tx)
	if adaptedTx.IsContract() {
		address := tx.Vout[transaction.ContractTxouputIndex].GetAddress()
		return &TxContract{tx, address}
	}
	return nil
}

// IsScheduleContract returns if the contract contains 'dapp_schedule'
func (ctx *TxContract) IsScheduleContract() bool {
	if !strings.Contains(ctx.GetContract(), scheduleFuncName) {
		return true
	}
	return false
}

//GetContract returns the smart contract code in a transaction
func (ctx *TxContract) GetContract() string {
	return ctx.Vout[transaction.ContractTxouputIndex].Contract
}

//GetContractPubKeyHash returns the smart contract pubkeyhash in a transaction
func (ctx *TxContract) GetContractPubKeyHash() account.PubKeyHash {
	return ctx.Vout[transaction.ContractTxouputIndex].PubKeyHash
}

// GasCountOfTxBase calculate the actual amount for a tx with data
func (ctx *TxContract) GasCountOfTxBase() (*common.Amount, error) {
	txGas := transaction.MinGasCountPerTransaction
	if dataLen := ctx.DataLen(); dataLen > 0 {
		dataGas := common.NewAmount(uint64(dataLen)).Mul(transaction.GasCountPerByte)
		baseGas := txGas.Add(dataGas)
		txGas = baseGas
	}
	return txGas, nil
}

// DataLen return the length of payload
func (ctx *TxContract) DataLen() int {
	return len([]byte(ctx.GetContract()))
}

// VerifyGas verifies if the transaction has the correct GasLimit and GasPrice
func (ctx *TxContract) VerifyGas(totalBalance *common.Amount) error {
	baseGas, err := ctx.GasCountOfTxBase()
	if err == nil {
		if ctx.GasLimit.Cmp(baseGas) < 0 {
			logger.WithFields(logger.Fields{
				"limit":       ctx.GasLimit,
				"acceptedGas": baseGas,
			}).Warn("Failed to check GasLimit >= txBaseGas.")
			// GasLimit is smaller than based tx gas, won't giveback the tx
			return transaction.ErrOutOfGasLimit
		}
	}

	limitedFee := ctx.GasLimit.Mul(ctx.GasPrice)
	if totalBalance.Cmp(limitedFee) < 0 {
		return transaction.ErrInsufficientBalance
	}
	return nil
}

// ToContractTx Returns structure of ContractTx
func ToContractTx(tx *transaction.Transaction) *TxContract {
	address := tx.Vout[transaction.ContractTxouputIndex].GetAddress()
	if tx.IsContract() {
		return &TxContract{tx, address}
	}
	if tx.Type != transaction.TxTypeDefault {
		return nil
	}
	txAdapter := transaction.NewTxAdapter(tx)
	if txAdapter.IsContract() {
		return &TxContract{tx, address}
	}
	return nil
}

//GetContractAddress gets the smart contract's address if a transaction deploys a smart contract
func (tx *TxContract) GetContractAddress() account.Address {
	return tx.Address
}

// VerifyInEstimate returns whether the current tx in estimate mode is valid.
func (tx *TxContract) VerifyInEstimate(utxoIndex *lutxo.UTXOIndex) error {
	prevUtxos, err := lutxo.FindVinUtxosInUtxoPool(utxoIndex, tx.Transaction)
	if err != nil {
		return err
	}
	return tx.verifyInEstimate(utxoIndex, prevUtxos)
}

func (tx *TxContract) verifyInEstimate(utxoIndex *lutxo.UTXOIndex, prevUtxos []*utxo.UTXO) error {
	if tx.IsScheduleContract() && !IsContractDeployed(utxoIndex, tx) {
		return errors.New("Transaction: contract state check failed")
	}
	err := tx.Transaction.Verify(prevUtxos)
	return err
}

//NewRewardTx creates a new transaction that gives reward to addresses according to the input rewards
func NewRewardTx(blockHeight uint64, rewards map[string]string) transaction.Transaction {

	bh := make([]byte, 8)
	binary.BigEndian.PutUint64(bh, uint64(blockHeight))

	txin := transactionbase.TXInput{nil, -1, bh, transaction.RewardTxData}
	txOutputs := []transactionbase.TXOutput{}
	for address, amount := range rewards {
		amt, err := common.NewAmountFromString(amount)
		if err != nil {
			logger.WithError(err).WithFields(logger.Fields{
				"address": address,
				"amount":  amount,
			}).Warn("Transaction: failed to parse reward amount")
		}
		acc := account.NewContractAccountByAddress(account.NewAddress(address))
		txOutputs = append(txOutputs, *transactionbase.NewTXOutput(amt, acc))
	}
	tx := transaction.Transaction{nil, []transactionbase.TXInput{txin}, txOutputs, common.NewAmount(0), common.NewAmount(0), common.NewAmount(0), time.Now().UnixNano() / 1e6, transaction.TxTypeReward}

	tx.ID = tx.Hash()

	return tx
}

// NewGasRewardTx returns a reward to miner, earned for contract execution gas fee
func NewGasRewardTx(to *account.TransactionAccount, blockHeight uint64, actualGasCount *common.Amount, gasPrice *common.Amount, uniqueNum int) (transaction.Transaction, error) {
	fee := actualGasCount.Mul(gasPrice)
	txin := transactionbase.TXInput{nil, -1, getUniqueByte(blockHeight, uniqueNum), transaction.GasRewardData}
	txout := transactionbase.NewTXOutput(fee, to)
	tx := transaction.Transaction{nil, []transactionbase.TXInput{txin}, []transactionbase.TXOutput{*txout}, common.NewAmount(0), common.NewAmount(0), common.NewAmount(0), time.Now().UnixNano() / 1e6, transaction.TxTypeGasReward}
	tx.ID = tx.Hash()
	return tx, nil
}

// NewGasChangeTx returns a change to contract invoker, pay for the change of unused gas
func NewGasChangeTx(to *account.TransactionAccount, blockHeight uint64, actualGasCount *common.Amount, gasLimit *common.Amount, gasPrice *common.Amount, uniqueNum int) (transaction.Transaction, error) {
	if gasLimit.Cmp(actualGasCount) <= 0 {
		return transaction.Transaction{}, transaction.ErrNoGasChange
	}
	change, err := gasLimit.Sub(actualGasCount)

	if err != nil {
		return transaction.Transaction{}, err
	}
	changeValue := change.Mul(gasPrice)

	txin := transactionbase.TXInput{nil, -1, getUniqueByte(blockHeight, uniqueNum), transaction.GasChangeData}
	txout := transactionbase.NewTXOutput(changeValue, to)
	tx := transaction.Transaction{nil, []transactionbase.TXInput{txin}, []transactionbase.TXOutput{*txout}, common.NewAmount(0), common.NewAmount(0), common.NewAmount(0), time.Now().UnixNano() / 1e6, transaction.TxTypeGasChange}

	tx.ID = tx.Hash()
	return tx, nil
}

// NewCoinbaseTX creates a new coinbase transaction
func NewCoinbaseTX(to account.Address, data string, blockHeight uint64, tip *common.Amount) transaction.Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}
	bh := make([]byte, 8)
	binary.BigEndian.PutUint64(bh, uint64(blockHeight))
	toAccount := account.NewContractAccountByAddress(to)
	txin := transactionbase.TXInput{nil, -1, bh, []byte(data)}
	txout := transactionbase.NewTXOutput(transaction.Subsidy.Add(tip), toAccount)
	tx := transaction.Transaction{nil, []transactionbase.TXInput{txin}, []transactionbase.TXOutput{*txout}, common.NewAmount(0), common.NewAmount(0), common.NewAmount(0), time.Now().UnixNano() / 1e6, transaction.TxTypeCoinbase}
	tx.ID = tx.Hash()

	return tx
}

// NewUTXOTransaction creates a new transaction
func NewUTXOTransaction(utxos []*utxo.UTXO, sendTxParam transaction.SendTxParam) (transaction.Transaction, error) {
	fromAccount := account.NewContractAccountByAddress(sendTxParam.From)
	toAccount := account.NewContractAccountByAddress(sendTxParam.To)
	sum := transaction.CalculateUtxoSum(utxos)
	change, err := transaction.CalculateChange(sum, sendTxParam.Amount, sendTxParam.Tip, sendTxParam.GasLimit, sendTxParam.GasPrice)
	if err != nil {
		return transaction.Transaction{}, err
	}
	txType := transaction.TxTypeNormal
	if sendTxParam.Contract != "" {
		txType = transaction.TxTypeContract
	}
	tx := transaction.Transaction{
		nil,
		prepareInputLists(utxos, sendTxParam.SenderKeyPair.GetPublicKey(), nil),
		prepareOutputLists(fromAccount, toAccount, sendTxParam.Amount, change, sendTxParam.Contract),
		sendTxParam.Tip,
		sendTxParam.GasLimit,
		sendTxParam.GasPrice,
		time.Now().UnixNano() / 1e6,
		txType,
	}
	tx.ID = tx.Hash()

	err = tx.Sign(sendTxParam.SenderKeyPair.GetPrivateKey(), utxos)
	if err != nil {
		return transaction.Transaction{}, err
	}

	return tx, nil
}

func NewSmartContractDestoryTX(utxos []*utxo.UTXO, contractAddr account.Address, sourceTXID []byte) transaction.Transaction {
	sum := transaction.CalculateUtxoSum(utxos)
	tips := common.NewAmount(0)
	gasLimit := common.NewAmount(0)
	gasPrice := common.NewAmount(0)

	tx, _ := NewContractTransferTX(utxos, contractAddr, account.NewAddress(transaction.SCDestroyAddress), sum, tips, gasLimit, gasPrice, sourceTXID)
	return tx
}

func NewContractTransferTX(utxos []*utxo.UTXO, contractAddr, toAddr account.Address, amount, tip *common.Amount, gasLimit *common.Amount, gasPrice *common.Amount, sourceTXID []byte) (transaction.Transaction, error) {
	contractAccount := account.NewContractAccountByAddress(contractAddr)
	toAccount := account.NewContractAccountByAddress(toAddr)
	if !contractAccount.IsValid() {
		return transaction.Transaction{}, account.ErrInvalidAddress
	}
	if isContract, err := contractAccount.GetPubKeyHash().IsContract(); !isContract {
		return transaction.Transaction{}, err
	}

	sum := transaction.CalculateUtxoSum(utxos)
	change, err := transaction.CalculateChange(sum, amount, tip, gasLimit, gasPrice)
	if err != nil {
		return transaction.Transaction{}, err
	}

	// Intentionally set PubKeyHash as PubKey (to recognize it is from contract) and sourceTXID as signature in Vin
	tx := transaction.Transaction{
		nil,
		prepareInputLists(utxos, contractAccount.GetPubKeyHash(), sourceTXID),
		prepareOutputLists(contractAccount, toAccount, amount, change, ""),
		tip,
		gasLimit,
		gasPrice,
		time.Now().UnixNano() / 1e6,
		transaction.TxTypeNormal,
	}
	tx.ID = tx.Hash()

	return tx, nil
}

//prepareInputLists prepares a list of txinputs for a new transaction
func prepareInputLists(utxos []*utxo.UTXO, publicKey []byte, signature []byte) []transactionbase.TXInput {
	var inputs []transactionbase.TXInput

	// Build a list of inputs
	for _, utxo := range utxos {
		input := transactionbase.TXInput{utxo.Txid, utxo.TxIndex, signature, publicKey}
		inputs = append(inputs, input)
	}

	return inputs
}

//prepareOutPutLists prepares a list of txoutputs for a new transaction
func prepareOutputLists(from, to *account.TransactionAccount, amount *common.Amount, change *common.Amount, contract string) []transactionbase.TXOutput {
	var outputs []transactionbase.TXOutput
	toAddr := to

	if toAddr.GetAddress().String() == "" {
		toAddr = account.NewContractTransactionAccount()
	}

	if contract != "" {
		outputs = append(outputs, *transactionbase.NewContractTXOutput(toAddr, contract))
	}

	outputs = append(outputs, *transactionbase.NewTXOutput(amount, toAddr))
	if !change.IsZero() {
		outputs = append(outputs, *transactionbase.NewTXOutput(change, from))
	}
	return outputs
}

func (tx *TxContract) GetTotalBalance(prevUtxos []*utxo.UTXO) (*common.Amount, error) {
	totalPrev := transaction.CalculateUtxoSum(prevUtxos)
	totalVoutValue, _ := tx.CalculateTotalVoutValue()
	totalBalance, err := totalPrev.Sub(totalVoutValue)
	if err != nil {
		return nil, transaction.ErrInsufficientBalance
	}
	totalBalance, _ = totalBalance.Sub(tx.Tip)
	return totalBalance, nil
}

func getUniqueByte(height uint64, uniqueNum int) []byte {
	bh := make([]byte, 8)
	binary.BigEndian.PutUint64(bh, uint64(height))
	bUnique := make([]byte, 2)
	binary.BigEndian.PutUint16(bUnique, uint16(uniqueNum))
	bh[0] = bUnique[0]
	bh[1] = bUnique[1]
	return bh
}