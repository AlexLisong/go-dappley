package transactionpool

import (
	"github.com/dappley/go-dappley/core/transaction"
	"github.com/gogo/protobuf/proto"

	transactionpb "github.com/dappley/go-dappley/core/transaction/pb"
	transactionPoolpb "github.com/dappley/go-dappley/core/transactionpool/pb"
)

type TransactionPool struct {
	Txs      map[string]*transaction.TransactionNode
	TipOrder []string
	CurrSize uint32
}

func NewTransactionPool() *TransactionPool {
	txPool := &TransactionPool{
		Txs:      make(map[string]*transaction.TransactionNode),
		TipOrder: make([]string, 0),
		CurrSize: 0,
	}
	return txPool
}

func (txPool *TransactionPool) ToProto() proto.Message {
	txs := make(map[string]*transactionpb.TransactionNode)
	for key, val := range txPool.Txs {
		txs[key] = val.ToProto().(*transactionpb.TransactionNode)
	}
	return &transactionPoolpb.TransactionPool{
		Txs:      txs,
		TipOrder: txPool.TipOrder,
		CurrSize: txPool.CurrSize,
	}
}

func (txPool *TransactionPool) FromProto(pb proto.Message) {
	for key, val := range pb.(*transactionPoolpb.TransactionPool).Txs {
		txNode := transaction.NewTransactionNode(nil)
		txNode.FromProto(val)
		txPool.Txs[key] = txNode
	}
	txPool.TipOrder = pb.(*transactionPoolpb.TransactionPool).TipOrder
	txPool.CurrSize = pb.(*transactionPoolpb.TransactionPool).CurrSize

}
