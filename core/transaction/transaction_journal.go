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
package transaction

import (
	transactionpb "github.com/dappley/go-dappley/core/transaction/pb"
	"github.com/dappley/go-dappley/core/transactionbase"
	transactionbasepb "github.com/dappley/go-dappley/core/transactionbase/pb"
	"github.com/golang/protobuf/proto"
	logger "github.com/sirupsen/logrus"
)

// TxJournal refers to transaction log data.
// It holds output array in each transactionbase.
type TxJournal struct {
	Txid []byte
	Vout []transactionbase.TXOutput
}

// Constructor
func NewTxJournal(txid []byte, vouts []transactionbase.TXOutput) *TxJournal {
	txJournal := &TxJournal{txid, vouts}
	return txJournal
}

func (txJournal *TxJournal) SerializeJournal() ([]byte, error) {
	rawBytes, err := proto.Marshal(txJournal.toProto())
	if err != nil {
		logger.WithError(err).Panic("TransactionJournal: Cannot serialize transactionJournal!")
		return nil, err
	}
	return rawBytes, nil
}

func DeserializeJournal(b []byte) (*TxJournal, error) {
	pb := &transactionpb.TransactionJournal{}
	err := proto.Unmarshal(b, pb)
	if err != nil {
		logger.WithError(err).Panic("TransactionJournal: Cannot deserialize transactionJournal!")
		return &TxJournal{}, err
	}
	txJournal := &TxJournal{}
	txJournal.fromProto(pb)
	return txJournal, nil
}

func (txJournal *TxJournal) toProto() proto.Message {
	var voutArray []*transactionbasepb.TXOutput
	for _, txout := range txJournal.Vout {
		voutArray = append(voutArray, txout.ToProto().(*transactionbasepb.TXOutput))
	}
	return &transactionpb.TransactionJournal{
		Vout: voutArray,
	}
}

func (txJournal *TxJournal) fromProto(pb proto.Message) {
	var voutArray []transactionbase.TXOutput
	txout := transactionbase.TXOutput{}
	for _, txoutpb := range pb.(*transactionpb.TransactionJournal).GetVout() {
		txout.FromProto(txoutpb)
		voutArray = append(voutArray, txout)
	}
	txJournal.Vout = voutArray
}
