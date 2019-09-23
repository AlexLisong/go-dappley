package ltransaction

import (
	"testing"

	"github.com/dappley/go-dappley/common"
	"github.com/dappley/go-dappley/core/transaction"
	"github.com/dappley/go-dappley/core/transactionbase"
	"github.com/dappley/go-dappley/storage"
	"github.com/dappley/go-dappley/util"
	"github.com/stretchr/testify/assert"
)

var tx1 = transaction.Transaction{
	ID:       util.GenerateRandomAoB(1),
	Vin:      transactionbase.GenerateFakeTxInputs(),
	Vout:     transactionbase.GenerateFakeTxOutputs(),
	Tip:      common.NewAmount(5),
	GasLimit: common.NewAmount(0),
	GasPrice: common.NewAmount(0),
}

func TestJournalPutAndGet(t *testing.T) {
	db := storage.NewRamStorage()
	vin := transactionbase.TXInput{tx1.ID, 1, nil, nil}
	err := PutTxJournal(tx1, db)
	assert.Nil(t, err)
	vout, err := GetTxOutput(vin, db)
	// Expect transaction logs have been successfully saved
	assert.Nil(t, err)
	assert.Equal(t, vout.Account.GetPubKeyHash(), tx1.Vout[1].Account.GetPubKeyHash())
}
