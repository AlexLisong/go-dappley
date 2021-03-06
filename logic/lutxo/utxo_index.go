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

package lutxo

import (
	"encoding/hex"
	"errors"

	"github.com/dappley/go-dappley/storage"

	"sync"

	"github.com/dappley/go-dappley/core/transaction"

	"github.com/dappley/go-dappley/core/block"

	"github.com/dappley/go-dappley/common"
	"github.com/dappley/go-dappley/core/account"
	"github.com/dappley/go-dappley/core/transactionbase"
	"github.com/dappley/go-dappley/core/utxo"
	logger "github.com/sirupsen/logrus"
)

var contractUtxoKey = []byte("contractUtxoKey")

var (
	ErrUTXONotFound   = errors.New("utxo not found when trying to remove from cache")
	ErrTXInputInvalid = errors.New("txInput refers to non-existing transaction")
)

// UTXOIndex holds all unspent TXOutputs indexed by public key hash.
type UTXOIndex struct {
	index map[string]*utxo.UTXOTx
	cache *utxo.UTXOCache
	mutex *sync.RWMutex
}

// NewUTXOIndex initializes an UTXOIndex instance
func NewUTXOIndex(cache *utxo.UTXOCache) *UTXOIndex {
	return &UTXOIndex{
		index: make(map[string]*utxo.UTXOTx),
		cache: cache,
		mutex: &sync.RWMutex{},
	}
}

func (utxos *UTXOIndex) SetIndex(index map[string]*utxo.UTXOTx) {
	utxos.index = index
}

func (utxos *UTXOIndex) Save() error {
	utxos.mutex.Lock()
	defer utxos.mutex.Unlock()
	for key, utxoTx := range utxos.index {
		pubKeyHash, err := hex.DecodeString(key)
		if err != nil {
			return err
		}

		err = utxos.cache.Put(pubKeyHash, utxoTx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (utxos *UTXOIndex) Delete() error {
	return nil
}

// GetAllUTXOsByPubKeyHash returns all current UTXOs identified by pubkey.
func (utxos *UTXOIndex) GetAllUTXOsByPubKeyHash(pubkeyHash account.PubKeyHash) *utxo.UTXOTx {
	key := pubkeyHash.String()
	utxos.mutex.RLock()
	utxoTx, ok := utxos.index[key]
	utxos.mutex.RUnlock()

	if ok {
		return utxoTx
	}

	utxoTx = utxos.cache.Get(pubkeyHash)
	utxos.mutex.Lock()
	newUtxoTx := utxoTx.DeepCopy()
	if utxos.index[key] != nil {
		utxo.Free(utxos.index[key])
	}
	utxos.index[key] = newUtxoTx
	utxos.mutex.Unlock()
	return newUtxoTx
}

//SplitContractUtxo
func (utxos *UTXOIndex) SplitContractUtxo(pubkeyHash account.PubKeyHash) (*utxo.UTXO, []*utxo.UTXO) {
	if ok, _ := account.PubKeyHash(pubkeyHash).IsContract(); !ok {
		return nil, nil
	}

	utxoTx := utxos.GetAllUTXOsByPubKeyHash(pubkeyHash)

	var invokeContractUtxos []*utxo.UTXO
	var createContractUtxo *utxo.UTXO

	for _, u := range utxoTx.Indices {
		if u.UtxoType == utxo.UtxoCreateContract {
			createContractUtxo = u
		} else {
			invokeContractUtxos = append(invokeContractUtxos, u)
		}
	}
	return createContractUtxo, invokeContractUtxos
}

// GetUTXOsByAmount returns a number of UTXOs that has a sum more than or equal to the amount
func (utxos *UTXOIndex) GetUTXOsByAmount(pubkeyHash account.PubKeyHash, amount *common.Amount) ([]*utxo.UTXO, error) {
	allUtxos := utxos.GetAllUTXOsByPubKeyHash(pubkeyHash)
	retUtxos, ok := allUtxos.PrepareUtxos(amount)
	if !ok {
		return nil, transaction.ErrInsufficientFund
	}

	return retUtxos, nil
}

// FindUTXOByVin returns the UTXO instance identified by pubkeyHash, txid and vout
func (utxos *UTXOIndex) FindUTXOByVin(pubkeyHash account.PubKeyHash, txid []byte, vout int) *utxo.UTXO {
	utxosOfKey := utxos.GetAllUTXOsByPubKeyHash(pubkeyHash)
	return utxosOfKey.GetUtxo(txid, vout)
}

func (utxos *UTXOIndex) UpdateUtxo(tx *transaction.Transaction) bool {
	if !tx.IsCoinbase() && !tx.IsRewardTx() && !tx.IsGasRewardTx() && !tx.IsGasChangeTx() {
		for _, txin := range tx.Vin {
			isContract, _ := account.PubKeyHash(txin.PubKey).IsContract()
			// spent contract utxo
			pubKeyHash := txin.PubKey
			if !isContract {
				// spent normal utxo
				ta := account.NewTransactionAccountByPubKey(txin.PubKey)
				_, err := account.IsValidPubKey(txin.PubKey)
				if err != nil {
					logger.WithError(err).Warn("UTXOIndex: txin.pubKey error, discard update in utxo.")
					return false
				}
				pubKeyHash = ta.GetPubKeyHash()
			}

			err := utxos.removeUTXO(pubKeyHash, txin.Txid, txin.Vout)
			if err != nil {
				logger.WithError(err).Warn("UTXOIndex: removeUTXO error, discard update in utxo.")
				return false
			}
		}
	}
	for i, txout := range tx.Vout {
		utxos.AddUTXO(txout, tx.ID, i)
	}
	return true
}

// Update removes the UTXOs spent in the transactions in newBlk from the index and adds UTXOs generated in the
// transactions to the index. The index will be saved to db as a result. If saving failed, index won't be updated.
func (utxos *UTXOIndex) UpdateUtxos(txs []*transaction.Transaction) {
	// Create a copy of the index so operations below are only temporal
	for _, tx := range txs {
		utxos.UpdateUtxo(tx)
	}
}

// UndoTxsInBlock compute the (previous) UTXOIndex resulted from undoing the transactions in given blk.
// Note that the operation does not save the index to db.
func (utxos *UTXOIndex) UndoTxsInBlock(blk *block.Block, db storage.Storage) error {

	for i := len(blk.GetTransactions()) - 1; i >= 0; i-- {
		tx := blk.GetTransactions()[i]
		err := utxos.excludeVoutsInTx(tx, db)
		if err != nil {
			return err
		}
		if tx.IsCoinbase() || tx.IsRewardTx() || tx.IsGasRewardTx() || tx.IsGasChangeTx() {
			continue
		}
		err = utxos.unspendVinsInTx(tx, db)
		if err != nil {
			return err
		}
	}
	return nil
}

// excludeVoutsInTx removes the UTXOs generated in a transaction from the UTXOIndex.
func (utxos *UTXOIndex) excludeVoutsInTx(tx *transaction.Transaction, db storage.Storage) error {
	for i, vout := range tx.Vout {
		err := utxos.removeUTXO(vout.PubKeyHash, tx.ID, i)
		if err != nil {
			return err
		}
	}
	return nil
}

func getTXOutputSpent(in transactionbase.TXInput, db storage.Storage) (transactionbase.TXOutput, int, error) {

	vout, err := transaction.GetTxOutput(in, db)

	if err != nil {
		return transactionbase.TXOutput{}, 0, ErrTXInputInvalid
	}
	return vout, in.Vout, nil
}

// unspendVinsInTx adds UTXOs back to the UTXOIndex as a result of undoing the spending of the UTXOs in a transaction.
func (utxos *UTXOIndex) unspendVinsInTx(tx *transaction.Transaction, db storage.Storage) error {
	for _, vin := range tx.Vin {
		vout, voutIndex, err := getTXOutputSpent(vin, db)
		if err != nil {
			return err
		}
		utxos.AddUTXO(vout, vin.Txid, voutIndex)
	}
	return nil
}

// AddUTXO adds an unspent TXOutput to index
func (utxos *UTXOIndex) AddUTXO(txout transactionbase.TXOutput, txid []byte, vout int) {
	originalUtxos := utxos.GetAllUTXOsByPubKeyHash(txout.PubKeyHash)

	var u *utxo.UTXO
	//if it is a smart contract deployment utxo add it to contract utxos
	if isContract, _ := txout.PubKeyHash.IsContract(); isContract {
		if originalUtxos.Size() == 0 {
			u = utxo.NewUTXO(txout, txid, vout, utxo.UtxoCreateContract)
			contractUtxos := utxos.GetAllUTXOsByPubKeyHash(contractUtxoKey)
			utxos.mutex.Lock()
			contractUtxos.PutUtxo(u)
			utxos.index[hex.EncodeToString(contractUtxoKey)] = contractUtxos
			utxos.mutex.Unlock()
		} else {
			u = utxo.NewUTXO(txout, txid, vout, utxo.UtxoInvokeContract)
		}
	} else {
		u = utxo.NewUTXO(txout, txid, vout, utxo.UtxoNormal)
	}

	utxos.mutex.Lock()
	originalUtxos.PutUtxo(u)
	utxos.index[txout.PubKeyHash.String()] = originalUtxos
	utxos.mutex.Unlock()
}

func (utxos *UTXOIndex) GetContractUtxos() []*utxo.UTXO {
	utxoTx := utxos.GetAllUTXOsByPubKeyHash(contractUtxoKey)

	var contractUtxos []*utxo.UTXO
	for _, utxo := range utxoTx.Indices {
		contractUtxos = append(contractUtxos, utxo)
	}
	return contractUtxos
}

// removeUTXO finds and removes a UTXO from UTXOIndex
func (utxos *UTXOIndex) removeUTXO(pkh account.PubKeyHash, txid []byte, vout int) error {
	originalUtxos := utxos.GetAllUTXOsByPubKeyHash(pkh)

	u := originalUtxos.GetUtxo(txid, vout)
	if u == nil {
		return ErrUTXONotFound
	}

	utxos.mutex.Lock()
	originalUtxos.RemoveUtxo(txid, vout)
	utxos.index[pkh.String()] = originalUtxos
	utxos.mutex.Unlock()

	if u.UtxoType != utxo.UtxoCreateContract {
		return nil
	}
	// remove contract utxos
	isContract, _ := pkh.IsContract()
	if isContract {
		contractUtxos := utxos.GetAllUTXOsByPubKeyHash(contractUtxoKey)

		contractUtxo := contractUtxos.GetUtxo(txid, vout)

		if contractUtxo == nil {
			return ErrUTXONotFound
		}
		utxos.mutex.Lock()
		contractUtxos.RemoveUtxo(txid, vout)
		utxos.index[hex.EncodeToString(contractUtxoKey)] = contractUtxos
		utxos.mutex.Unlock()
	}
	return nil
}

//creates a deepcopy of the receiver object
func (utxos *UTXOIndex) DeepCopy() *UTXOIndex {
	utxos.mutex.RLock()
	defer utxos.mutex.RUnlock()

	utxocopy := NewUTXOIndex(utxos.cache)
	for pkh, utxoTx := range utxos.index {
		newUtxoTx := utxoTx.DeepCopy()
		utxocopy.index[pkh] = newUtxoTx
	}
	return utxocopy
}
