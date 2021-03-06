package lutxo

import (
	"encoding/hex"

	"github.com/dappley/go-dappley/core/transaction"

	"github.com/dappley/go-dappley/core/account"
	"github.com/dappley/go-dappley/core/utxo"
	logger "github.com/sirupsen/logrus"
)

//FindVinUtxosInUtxoPool Find the transaction in a utxo pool. Returns true only if all Vins are found in the utxo pool
func FindVinUtxosInUtxoPool(utxoPool UTXOIndex, tx transaction.Transaction) ([]*utxo.UTXO, error) {
	var res []*utxo.UTXO
	for _, vin := range tx.Vin {
		// some vin.PubKey is contract address's PubKeyHash
		isContract, _ := account.PubKeyHash(vin.PubKey).IsContract()
		pubKeyHash := vin.PubKey
		if !isContract {
			if ok, _ := account.IsValidPubKey(vin.PubKey); !ok {
				return nil, transaction.ErrNewUserPubKeyHash
			}
			ta := account.NewTransactionAccountByPubKey(vin.PubKey)
			pubKeyHash = ta.GetPubKeyHash()
		}
		utxo := utxoPool.FindUTXOByVin([]byte(pubKeyHash), vin.Txid, vin.Vout)
		if utxo == nil {
			logger.WithFields(logger.Fields{
				"txid":      hex.EncodeToString(tx.ID),
				"vin_id":    hex.EncodeToString(vin.Txid),
				"vin_index": vin.Vout,
			}).Warn("Transaction: Can not find vin")
			return nil, transaction.ErrTXInputNotFound
		}
		res = append(res, utxo)
	}
	return res, nil
}
