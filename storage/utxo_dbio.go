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

package storage

import (
	"github.com/dappley/go-dappley/core/account"
	"github.com/dappley/go-dappley/core/transactionbase"
	"github.com/dappley/go-dappley/core/utxo"
	lru "github.com/hashicorp/golang-lru"
)

const UtxoCacheLRUCacheLimit = 1024

// UTXOCache holds temporary UTXOTx data
type UTXODBIO struct {
	// key: address, value: UTXOTx
	cache        *lru.Cache
	db           Storage
	txOutPutDBIO *TXOutPutDBIO
}

func NewUTXODBIO(db Storage) *UTXODBIO {
	utxoDBIO := &UTXODBIO{
		cache:        nil,
		db:           db,
		txOutPutDBIO: NewTXOutPutDBIO(db),
	}
	utxoDBIO.cache, _ = lru.New(UtxoCacheLRUCacheLimit)
	return utxoDBIO
}

// Return value from cache
func (utxoDBIO *UTXODBIO) GetTxOutput(vin transactionbase.TXInput) (transactionbase.TXOutput, error) {
	return utxoDBIO.txOutPutDBIO.GetTxOutput(vin)
}

// Return value from cache
func (utxoDBIO *UTXODBIO) Get(pubKeyHash account.PubKeyHash) *utxo.UTXOTx {
	mapData, ok := utxoDBIO.cache.Get(string(pubKeyHash))
	if ok {
		return mapData.(*utxo.UTXOTx)
	}

	rawBytes, err := utxoDBIO.db.Get(pubKeyHash)

	var utxoTx utxo.UTXOTx
	if err == nil {
		utxoTx = utxo.DeserializeUTXOTx(rawBytes)
		utxoDBIO.cache.Add(string(pubKeyHash), &utxoTx)
	} else {
		utxoTx = utxo.NewUTXOTx()
	}
	return &utxoTx
}

// Add new data into cache
func (utxoDBIO *UTXODBIO) Put(pubKeyHash account.PubKeyHash, value *utxo.UTXOTx) error {
	if pubKeyHash == nil {
		return account.ErrEmptyPublicKeyHash
	}
	savedUtxoTx := value.DeepCopy()
	utxoDBIO.cache.Add(string(pubKeyHash), &savedUtxoTx)
	return utxoDBIO.db.Put(pubKeyHash, value.Serialize())
}

func (utxoDBIO *UTXODBIO) Delete(pubKeyHash account.PubKeyHash) error {
	if pubKeyHash == nil {
		return account.ErrEmptyPublicKeyHash
	}
	return utxoDBIO.db.Del(pubKeyHash)
}
