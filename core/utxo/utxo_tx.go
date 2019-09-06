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

package utxo

import (
	"github.com/dappley/go-dappley/common"
	utxopb "github.com/dappley/go-dappley/core/utxo/pb"
	"github.com/golang/protobuf/proto"
	"github.com/raviqqe/hamt"
	logger "github.com/sirupsen/logrus"
	"hash/fnv"
	"strconv"
)

// UTXOTx holds txid_vout and UTXO pairs
//type UTXOTx hamt.Map

type UTXOTx struct {
	Indices map[string]*UTXO
}

type StringEntry string

func (key *StringEntry) Hash() uint32 {
	h := fnv.New32a()
	h.Write([]byte(string(*key)))
	return h.Sum32()
}

func (key *StringEntry) Equal(other hamt.Entry) bool {
	otherStr, ok := other.(*StringEntry)
	if !ok {
		return false
	}

	return string(*key) == string(*otherStr)
}

func NewUTXOTx() UTXOTx {
	return UTXOTx{Indices: map[string]*UTXO{}}
}

// Construct with UTXO data
func NewUTXOTxWithData(utxo *UTXO) UTXOTx {
	key := string(utxo.Txid) + "_" + strconv.Itoa(utxo.TxIndex)
	return UTXOTx{Indices: map[string]*UTXO{key: utxo}}
}

// Construct with map size
func NewUTXOTxWithSize(size int) UTXOTx {
	return UTXOTx{Indices: make(map[string]*UTXO, size)}
}

func DeserializeUTXOTx(d []byte) UTXOTx {
	utxoTx := NewUTXOTx()

	utxoList := &utxopb.UtxoList{}
	err := proto.Unmarshal(d, utxoList)
	if err != nil {
		logger.WithFields(logger.Fields{"error": err}).Error("UtxoTx: parse UtxoTx failed.")
		return utxoTx
	}

	for _, utxoPb := range utxoList.Utxos {
		var utxo = &UTXO{}
		utxo.FromProto(utxoPb)
		utxoTx.PutUtxo(utxo)
	}

	return utxoTx
}

func (utxoTx UTXOTx) Serialize() []byte {
	utxoList := &utxopb.UtxoList{}

	for _, utxo := range utxoTx.Indices {
		utxoList.Utxos = append(utxoList.Utxos, utxo.ToProto().(*utxopb.Utxo))
	}
	bytes, err := proto.Marshal(utxoList)
	if err != nil {
		logger.WithFields(logger.Fields{"error": err}).Error("UtxoTx: serialize UtxoTx failed.")
		return nil
	}
	return bytes
}

// Returns utxo info by transaction id and vout index
func (utxoTx UTXOTx) GetUtxo(txid []byte, vout int) *UTXO {
	key := string(txid) + "_" + strconv.Itoa(vout)
	utxo, ok := utxoTx.Indices[key]
	if !ok {
		return nil
	}
	return utxo
}

// Add new utxo to map
func (utxoTx UTXOTx) PutUtxo(utxo *UTXO) {
	key := string(utxo.Txid) + "_" + strconv.Itoa(utxo.TxIndex)
	utxoTx.Indices[key] = utxo
}

// Delete invalid element in map
func (utxoTx UTXOTx) RemoveUtxo(txid []byte, vout int) {
	key := string(txid) + "_" + strconv.Itoa(vout)
	delete(utxoTx.Indices, key)
}

func (utxoTx UTXOTx) Size() int {
	l := len(utxoTx.Indices)
	return l
}

func (utxoTx UTXOTx) GetAllUtxos() []*UTXO {
	var utxos []*UTXO
	for _, utxo := range utxoTx.Indices {
		utxos = append(utxos, utxo)
	}
	return utxos
}

func (utxoTx UTXOTx) PrepareUtxos(amount *common.Amount) ([]*UTXO, bool) {
	sum := common.NewAmount(0)

	if utxoTx.Size() < 1 {
		return nil, false
	}

	var utxos []*UTXO
	for _, utxo := range utxoTx.Indices {
		if utxo.UtxoType == UtxoCreateContract {
			continue
		}

		sum = sum.Add(utxo.Value)
		utxos = append(utxos, utxo)
		if sum.Cmp(amount) >= 0 {
			return utxos, true
		}
	}
	return nil, false
}

func (utxoTx UTXOTx) DeepCopy() UTXOTx {
	newUtxoTx := NewUTXOTxWithSize(utxoTx.Size())
	for key, utxo := range utxoTx.Indices {
		newUtxoTx.Indices[key] = utxo
	}
	return newUtxoTx
}
