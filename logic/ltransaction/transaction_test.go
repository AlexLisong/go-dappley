package ltransaction

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/binary"
	"errors"
	"testing"

	"github.com/dappley/go-dappley/common"
	"github.com/dappley/go-dappley/core"
	"github.com/dappley/go-dappley/core/account"
	"github.com/dappley/go-dappley/core/scState"
	"github.com/dappley/go-dappley/core/transaction"
	"github.com/dappley/go-dappley/core/transactionbase"
	"github.com/dappley/go-dappley/core/utxo"
	"github.com/dappley/go-dappley/crypto/keystore/secp256k1"
	"github.com/dappley/go-dappley/logic/lutxo"
	"github.com/dappley/go-dappley/storage"
	"github.com/dappley/go-dappley/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var tx1 = transaction.Transaction{
	ID:       util.GenerateRandomAoB(1),
	Vin:      transactionbase.GenerateFakeTxInputs(),
	Vout:     transactionbase.GenerateFakeTxOutputs(),
	Tip:      common.NewAmount(5),
	GasLimit: common.NewAmount(0),
	GasPrice: common.NewAmount(0),
}

func TestSign(t *testing.T) {
	// Fake a key pair
	privKey, _ := ecdsa.GenerateKey(secp256k1.S256(), bytes.NewReader([]byte("fakefakefakefakefakefakefakefakefakefake")))
	ecdsaPubKey, _ := secp256k1.FromECDSAPublicKey(&privKey.PublicKey)
	pubKey := append(privKey.PublicKey.X.Bytes(), privKey.PublicKey.Y.Bytes()...)
	ta := account.NewTransactionAccountByPubKey(pubKey)

	// Previous transactions containing UTXO of the Address
	prevTXs := []*utxo.UTXO{
		{transactionbase.TXOutput{common.NewAmount(13), ta.GetPubKeyHash(), ""}, []byte("01"), 0, utxo.UtxoNormal},
		{transactionbase.TXOutput{common.NewAmount(13), ta.GetPubKeyHash(), ""}, []byte("02"), 0, utxo.UtxoNormal},
		{transactionbase.TXOutput{common.NewAmount(13), ta.GetPubKeyHash(), ""}, []byte("03"), 0, utxo.UtxoNormal},
	}

	// New transaction to be signed (paid from the fake account)
	txin := []transactionbase.TXInput{
		{[]byte{1}, 0, nil, pubKey},
		{[]byte{3}, 0, nil, pubKey},
		{[]byte{3}, 2, nil, pubKey},
	}
	txout := []transactionbase.TXOutput{
		{common.NewAmount(19), ta.GetPubKeyHash(), ""},
	}
	tx := transaction.Transaction{nil, txin, txout, common.NewAmount(0), common.NewAmount(0), common.NewAmount(0), 0}

	// ltransaction.Sign the transaction
	err := tx.Sign(*privKey, prevTXs)
	if assert.Nil(t, err) {
		// Assert that the signatures were created by the fake key pair
		for i, vin := range tx.Vin {

			if assert.NotNil(t, vin.Signature) {
				txCopy := tx.TrimmedCopy(false)
				txCopy.Vin[i].Signature = nil
				txCopy.Vin[i].PubKey = []byte(ta.GetPubKeyHash())

				verified, err := secp256k1.Verify(txCopy.Hash(), vin.Signature, ecdsaPubKey)
				assert.Nil(t, err)
				assert.True(t, verified)
			}
		}
	}
}

func TestVerifyCoinbaseTransaction(t *testing.T) {
	var prevTXs = map[string]transaction.Transaction{}

	var tx1 = transaction.Transaction{
		ID:   util.GenerateRandomAoB(1),
		Vin:  transactionbase.GenerateFakeTxInputs(),
		Vout: transactionbase.GenerateFakeTxOutputs(),
		Tip:  common.NewAmount(2),
	}

	var tx2 = transaction.Transaction{
		ID:   util.GenerateRandomAoB(1),
		Vin:  transactionbase.GenerateFakeTxInputs(),
		Vout: transactionbase.GenerateFakeTxOutputs(),
		Tip:  common.NewAmount(5),
	}
	var tx3 = transaction.Transaction{
		ID:   util.GenerateRandomAoB(1),
		Vin:  transactionbase.GenerateFakeTxInputs(),
		Vout: transactionbase.GenerateFakeTxOutputs(),
		Tip:  common.NewAmount(10),
	}
	var tx4 = transaction.Transaction{
		ID:   util.GenerateRandomAoB(1),
		Vin:  transactionbase.GenerateFakeTxInputs(),
		Vout: transactionbase.GenerateFakeTxOutputs(),
		Tip:  common.NewAmount(20),
	}
	prevTXs[string(tx1.ID)] = tx2
	prevTXs[string(tx2.ID)] = tx3
	prevTXs[string(tx3.ID)] = tx4

	// test verifying coinbase transactions
	var t5 = transaction.NewCoinbaseTX(account.NewAddress("13ZRUc4Ho3oK3Cw56PhE5rmaum9VBeAn5F"), "", 5, common.NewAmount(0))
	bh1 := make([]byte, 8)
	binary.BigEndian.PutUint64(bh1, 5)
	txin1 := transactionbase.TXInput{nil, -1, bh1, []byte("Reward to test")}
	txout1 := transactionbase.NewTXOutput(common.NewAmount(10000000), account.NewContractAccountByAddress(account.NewAddress("13ZRUc4Ho3oK3Cw56PhE5rmaum9VBeAn5F")))
	var t6 = transaction.Transaction{nil, []transactionbase.TXInput{txin1}, []transactionbase.TXOutput{*txout1}, common.NewAmount(0), common.NewAmount(0), common.NewAmount(0), 0}

	// test valid coinbase transaction
	err5 := VerifyTransaction(&lutxo.UTXOIndex{}, &t5, 5)
	assert.Nil(t, err5)
	err6 := VerifyTransaction(&lutxo.UTXOIndex{}, &t6, 5)
	assert.Nil(t, err6)

	// test coinbase transaction with incorrect blockHeight
	err5 = VerifyTransaction(&lutxo.UTXOIndex{}, &t5, 10)
	assert.NotNil(t, err5)

	// test coinbase transaction with incorrect Subsidy
	bh2 := make([]byte, 8)
	binary.BigEndian.PutUint64(bh2, 5)
	txin2 := transactionbase.TXInput{nil, -1, bh2, []byte(nil)}
	txout2 := transactionbase.NewTXOutput(common.NewAmount(9), account.NewContractAccountByAddress(account.NewAddress("13ZRUc4Ho3oK3Cw56PhE5rmaum9VBeAn5F")))
	var t7 = transaction.Transaction{nil, []transactionbase.TXInput{txin2}, []transactionbase.TXOutput{*txout2}, common.NewAmount(0), common.NewAmount(0), common.NewAmount(0), 0}
	err7 := VerifyTransaction(&lutxo.UTXOIndex{}, &t7, 5)
	assert.NotNil(t, err7)

}

func TestVerifyNoCoinbaseTransaction(t *testing.T) {
	// Fake a key pair
	privKey, _ := ecdsa.GenerateKey(secp256k1.S256(), bytes.NewReader([]byte("fakefakefakefakefakefakefakefakefakefake")))
	privKeyByte, _ := secp256k1.FromECDSAPrivateKey(privKey)
	pubKey := append(privKey.PublicKey.X.Bytes(), privKey.PublicKey.Y.Bytes()...)
	ta := account.NewTransactionAccountByPubKey(pubKey)

	// Fake a wrong key pair
	wrongPrivKey, _ := ecdsa.GenerateKey(secp256k1.S256(), bytes.NewReader([]byte("FAKEfakefakefakefakefakefakefakefakefake")))
	wrongPrivKeyByte, _ := secp256k1.FromECDSAPrivateKey(wrongPrivKey)
	wrongPubKey := append(wrongPrivKey.PublicKey.X.Bytes(), wrongPrivKey.PublicKey.Y.Bytes()...)
	utxoIndex := lutxo.NewUTXOIndex(utxo.NewUTXOCache(storage.NewRamStorage()))
	utxoTx := utxo.NewUTXOTx()

	utxoTx.PutUtxo(&utxo.UTXO{transactionbase.TXOutput{common.NewAmount(4), ta.GetPubKeyHash(), ""}, []byte{1}, 0, utxo.UtxoNormal})
	utxoTx.PutUtxo(&utxo.UTXO{transactionbase.TXOutput{common.NewAmount(3), ta.GetPubKeyHash(), ""}, []byte{2}, 1, utxo.UtxoNormal})

	utxoIndex.SetIndex(map[string]*utxo.UTXOTx{
		ta.GetPubKeyHash().String(): &utxoTx,
	})

	// Prepare a transaction to be verified
	txin := []transactionbase.TXInput{{[]byte{1}, 0, nil, pubKey}}
	txin1 := append(txin, transactionbase.TXInput{[]byte{2}, 1, nil, pubKey})      // Normal test
	txin2 := append(txin, transactionbase.TXInput{[]byte{2}, 1, nil, wrongPubKey}) // previous not found with wrong pubkey
	txin3 := append(txin, transactionbase.TXInput{[]byte{3}, 1, nil, pubKey})      // previous not found with wrong Txid
	txin4 := append(txin, transactionbase.TXInput{[]byte{2}, 2, nil, pubKey})      // previous not found with wrong TxIndex
	txout := []transactionbase.TXOutput{{common.NewAmount(7), ta.GetPubKeyHash(), ""}}
	txout2 := []transactionbase.TXOutput{{common.NewAmount(8), ta.GetPubKeyHash(), ""}} //Vout amount > Vin amount

	tests := []struct {
		name     string
		tx       transaction.Transaction
		signWith []byte
		ok       error
	}{
		{"normal", transaction.Transaction{nil, txin1, txout, common.NewAmount(0), common.NewAmount(0), common.NewAmount(0), 0}, privKeyByte, nil},
		{"previous tx not found with wrong pubkey", transaction.Transaction{nil, txin2, txout, common.NewAmount(0), common.NewAmount(0), common.NewAmount(0), 0}, privKeyByte, errors.New("Transaction: prevUtxos not found")},
		{"previous tx not found with wrong Txid", transaction.Transaction{nil, txin3, txout, common.NewAmount(0), common.NewAmount(0), common.NewAmount(0), 0}, privKeyByte, errors.New("Transaction: prevUtxos not found")},
		{"previous tx not found with wrong TxIndex", transaction.Transaction{nil, txin4, txout, common.NewAmount(0), common.NewAmount(0), common.NewAmount(0), 0}, privKeyByte, errors.New("Transaction: prevUtxos not found")},
		{"Amount invalid", transaction.Transaction{nil, txin1, txout2, common.NewAmount(0), common.NewAmount(0), common.NewAmount(0), 0}, privKeyByte, errors.New("Transaction: ID is invalid")},
		{"ltransaction.Sign invalid", transaction.Transaction{nil, txin1, txout, common.NewAmount(0), common.NewAmount(0), common.NewAmount(0), 0}, wrongPrivKeyByte, errors.New("Transaction: ID is invalid")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tx.ID = tt.tx.Hash()
			// Generate signatures for all tx inputs
			for i := range tt.tx.Vin {
				txCopy := tt.tx.TrimmedCopy(false)
				txCopy.Vin[i].Signature = nil
				txCopy.Vin[i].PubKey = []byte(ta.GetPubKeyHash())
				signature, _ := secp256k1.Sign(txCopy.Hash(), tt.signWith)
				tt.tx.Vin[i].Signature = signature
			}

			// Verify the signatures
			err := VerifyTransaction(utxoIndex, &tt.tx, 0)
			assert.Equal(t, tt.ok, err)
		})
	}
}

func TestInvalidExecutionTx(t *testing.T) {
	var prikey1 = "bb23d2ff19f5b16955e8a24dca34dd520980fe3bddca2b3e1b56663f0ec1aa71"
	var pubkey1 = account.GenerateKeyPairByPrivateKey(prikey1).GetPublicKey()
	var ta1 = account.NewTransactionAccountByPubKey(pubkey1)
	var deploymentTx = transaction.Transaction{
		ID: nil,
		Vin: []transactionbase.TXInput{
			{tx1.ID, 1, nil, pubkey1},
		},
		Vout: []transactionbase.TXOutput{
			{common.NewAmount(5), ta1.GetPubKeyHash(), "dapp_schedule"},
		},
		Tip: common.NewAmount(1),
	}
	deploymentTx.ID = deploymentTx.Hash()
	contractPubkeyHash := deploymentTx.Vout[0].PubKeyHash

	utxoIndex := lutxo.NewUTXOIndex(utxo.NewUTXOCache(storage.NewRamStorage()))
	utxoTx := utxo.NewUTXOTx()

	utxoTx.PutUtxo(&utxo.UTXO{deploymentTx.Vout[0], deploymentTx.ID, 0, utxo.UtxoNormal})
	utxoIndex.SetIndex(map[string]*utxo.UTXOTx{
		ta1.GetPubKeyHash().String(): &utxoTx,
	})

	var executionTx = transaction.Transaction{
		ID: nil,
		Vin: []transactionbase.TXInput{
			{deploymentTx.ID, 0, nil, pubkey1},
		},
		Vout: []transactionbase.TXOutput{
			{common.NewAmount(3), contractPubkeyHash, "execution"},
		},
		Tip: common.NewAmount(2),
	}
	executionTx.ID = executionTx.Hash()
	executionTx.Sign(account.GenerateKeyPairByPrivateKey(prikey1).GetPrivateKey(), utxoIndex.GetAllUTXOsByPubKeyHash(ta1.GetPubKeyHash()).GetAllUtxos())

	err1 := VerifyTransaction(lutxo.NewUTXOIndex(utxo.NewUTXOCache(storage.NewRamStorage())), &executionTx, 0)
	err2 := VerifyTransaction(utxoIndex, &executionTx, 0)
	assert.NotNil(t, err1)
	assert.Nil(t, err2)
}

func TestTransaction_Execute(t *testing.T) {

	tests := []struct {
		name              string
		scAddr            string
		toAddr            string
		expectContractRun bool
	}{
		{
			name:              "CallAContract",
			scAddr:            "cWDSCWqwYRM6jNiN83PuRGvtcDuPpzBcfb",
			toAddr:            "cWDSCWqwYRM6jNiN83PuRGvtcDuPpzBcfb",
			expectContractRun: true,
		},
		{
			name:              "CallAWrongContractAddr",
			scAddr:            "cWDSCWqwYRM6jNiN83PuRGvtcDuPpzBcfb",
			toAddr:            "cavQdWxvUQU1HhBg1d7zJFwhf31SUaQwop",
			expectContractRun: false,
		},
		{
			name:              "NoPreviousContract",
			scAddr:            "",
			toAddr:            "cavQdWxvUQU1HhBg1d7zJFwhf31SUaQwop",
			expectContractRun: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := new(core.MockScEngine)
			contract := "helloworld!"
			toAccount := account.NewContractAccountByAddress(account.NewAddress(tt.toAddr))
			scAccount := account.NewContractAccountByAddress(account.NewAddress(tt.scAddr))

			toPKH := toAccount.GetPubKeyHash()
			scPKH := scAccount.GetPubKeyHash()

			scUtxo := utxo.UTXO{
				TxIndex: 0,
				Txid:    nil,
				TXOutput: transactionbase.TXOutput{
					PubKeyHash: scPKH,
					Contract:   contract,
				},
			}
			tx := transaction.Transaction{
				Vout:     []transactionbase.TXOutput{{nil, toPKH, "{\"function\":\"record\",\"args\":[\"dEhFf5mWTSe67mbemZdK3WiJh8FcCayJqm\",\"4\"]}"}},
				GasLimit: common.NewAmount(0),
				GasPrice: common.NewAmount(0),
			}
			ctx := tx.ToContractTx()

			index := lutxo.NewUTXOIndex(utxo.NewUTXOCache(storage.NewRamStorage()))
			if tt.scAddr != "" {
				index.AddUTXO(scUtxo.TXOutput, nil, 0)
			}

			if tt.expectContractRun {
				sc.On("ImportSourceCode", contract)
				sc.On("ImportLocalStorage", mock.Anything)
				sc.On("ImportContractAddr", mock.Anything)
				sc.On("ImportUTXOs", mock.Anything)
				sc.On("ImportSourceTXID", mock.Anything)
				sc.On("ImportRewardStorage", mock.Anything)
				sc.On("ImportTransaction", mock.Anything)
				sc.On("ImportContractCreateUTXO", mock.Anything)
				sc.On("ImportPrevUtxos", mock.Anything)
				sc.On("GetGeneratedTXs").Return([]*transaction.Transaction{})
				sc.On("ImportCurrBlockHeight", mock.Anything)
				sc.On("ImportSeed", mock.Anything)
				sc.On("Execute", mock.Anything, mock.Anything).Return("")
			}
			parentBlk := core.GenerateMockBlock()
			preUTXO, err := lutxo.FindVinUtxosInUtxoPool(*index, ctx.Transaction)

			if err != nil {
				println(err.Error())
			}
			isContractDeployed := IsContractDeployed(index, ctx)
			Execute(ctx, preUTXO, isContractDeployed, *index, scState.NewScState(), nil, sc, 0, parentBlk)
			sc.AssertExpectations(t)
		})
	}
}

//test IsCoinBase function
func TestIsCoinBase(t *testing.T) {
	var tx1 = transaction.Transaction{
		ID:   util.GenerateRandomAoB(1),
		Vin:  transactionbase.GenerateFakeTxInputs(),
		Vout: transactionbase.GenerateFakeTxOutputs(),
		Tip:  common.NewAmount(2),
	}

	assert.False(t, tx1.IsCoinbase())

	t2 := transaction.NewCoinbaseTX(account.NewAddress("13ZRUc4Ho3oK3Cw56PhE5rmaum9VBeAn5F"), "", 0, common.NewAmount(0))

	assert.True(t, t2.IsCoinbase())

}

func TestTransaction_IsRewardTx(t *testing.T) {
	tests := []struct {
		name        string
		tx          transaction.Transaction
		expectedRes bool
	}{
		{"normal", transaction.NewRewardTx(1, map[string]string{"dXnq2R6SzRNUt7ZANAqyZc2P9ziF6vYekB": "9"}), true},
		{"no rewards", transaction.NewRewardTx(1, nil), true},
		{"coinbase", transaction.NewCoinbaseTX(account.NewAddress("dXnq2R6SzRNUt7ZANAqyZc2P9ziF6vYekB"), "", 5, common.NewAmount(0)), false},
		{"normal tx", *core.MockTransaction(), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedRes, tt.tx.IsRewardTx())
		})
	}
}
