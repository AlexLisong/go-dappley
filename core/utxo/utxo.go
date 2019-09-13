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
	"github.com/dappley/go-dappley/core/account"
	"github.com/dappley/go-dappley/core/transactionbase"
	utxopb "github.com/dappley/go-dappley/core/utxo/pb"
	"github.com/golang/protobuf/proto"
)

type UtxoType int

const (
	UtxoNormal UtxoType = iota
	UtxoCreateContract
	UtxoInvokeContract
)

// UTXO contains the meta info of an unspent TXOutput.
type UTXO struct {
	transactionbase.TXOutput
	Txid     []byte
	TxIndex  int
	UtxoType UtxoType
}

// NewUTXO returns an UTXO instance constructed from a TXOutput.
func NewUTXO(txout transactionbase.TXOutput, txid []byte, vout int, utxoType UtxoType) *UTXO {
	return &UTXO{txout, txid, vout, utxoType}
}

func (utxo *UTXO) ToProto() proto.Message {
	return &utxopb.Utxo{
		Amount:        utxo.Value.Bytes(),
		PublicKeyHash: []byte(utxo.Account.GetPubKeyHash()),
		Txid:          utxo.Txid,
		TxIndex:       uint32(utxo.TxIndex),
		UtxoType:      uint32(utxo.UtxoType),
		Contract:      utxo.Contract,
	}
}

func (utxo *UTXO) FromProto(pb proto.Message) {
	utxopb := pb.(*utxopb.Utxo)
	utxo.Value = common.NewAmountFromBytes(utxopb.Amount)
	utxo.Account = account.NewContractAccountByPubKeyHash(utxopb.PublicKeyHash)
	utxo.Txid = utxopb.Txid
	utxo.TxIndex = int(utxopb.TxIndex)
	utxo.UtxoType = UtxoType(utxopb.UtxoType)
	utxo.Contract = utxopb.Contract
}
