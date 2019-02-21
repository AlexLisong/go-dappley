package tool

import (
	"encoding/json"
	"fmt"
	"github.com/dappley/go-dappley/client"
	"github.com/dappley/go-dappley/logic"
	logger "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"

	"github.com/dappley/go-dappley/common"
	"github.com/dappley/go-dappley/consensus"
	"github.com/dappley/go-dappley/core"
	"github.com/dappley/go-dappley/storage"
)

const (
	genesisAddr           = "121yKAXeG4cw6uaGCBYjWk9yTWmMkhcoDD"
	genesisFilePath       = "conf/genesis.conf"
	defaultPassword       = "password"
	defaultTimeBetweenBlk = 5
)

var (
	password             = "testpassword"
	maxWallet            = 4
	currBalance          = make(map[string]uint64)
	numOfTx				 = 100
	time 				 = int64(1532392928)
)

type FileInfo struct {
	Height        int
	DifferentFrom int
	Db            *storage.LevelDB
}

type Key struct {
	Key     string `json:"key"`
	Address string `json:"address"`
}

type Keys struct {
	Keys []Key `json:"keys"`
}

type GeneralConfigs struct{
	NumOfNormalTx int
}

func GenerateNewBlockChain(files []FileInfo, d *consensus.Dynasty, keys Keys, config GeneralConfigs) {
	bcs := make([]*core.Blockchain, len(files))
	addr := core.NewAddress(genesisAddr)
	numOfTx = config.NumOfNormalTx
	for i := range files{
		bc := core.CreateBlockchain(addr, files[i].Db, nil, 2000, nil)
		bcs[i] = bc
	}

	for i, p := range d.GetProducers(){
		logger.WithFields(logger.Fields{
			"producer" : p,
		}).Info("Producer:",i)
	}

	wm, err := logic.GetWalletManager(client.GetWalletFilePath())
	if err != nil {
		logger.Panic("Cannot get wallet manager.")
	}
	addrs := CreateWallet(wm)
	producer := core.NewAddress(d.ProducerAtATime(time))
	key := keys.getPrivateKeyByAddress(producer)
	logic.SetMinerKeyPair(key)

	//max, index := GetMaxHeightOfDifferentStart(files)
	//fund every miner
	parentBlks := make([]*core.Block, len(files))
	utxoIndexes := make([]*core.UTXOIndex, len(files))
	for i := range files {
		parentBlks[i],_ = bcs[i].GetTailBlock()
		utxoIndexes[i] = core.NewUTXOIndex()
		for j := 0; j < len(d.GetProducers()); j++ {
			b := generateBlock(utxoIndexes[i], parentBlks[i], bcs[i], d, keys, []*core.Transaction{})
			bcs[i].AddBlockToDb(b)
			parentBlks[i] = b
		}
	}

	//fund from miner
	fundingBlock := generateFundingBlock(utxoIndexes[0], parentBlks[0], bcs[0], d, keys, addrs[0], key, wm)
	for idx := range files {
		bcs[idx].AddBlockToDb(fundingBlock)
	}
	parentBlks[0] = fundingBlock

	for i, file := range files {
		makeBlockChainToSize(utxoIndexes[i],parentBlks[i], bcs[i], file.Height, d, keys, addrs, wm)
	}

}

func GetMaxHeightOfDifferentStart(files []FileInfo) (int, int) {
	max := 0
	index := 0
	for i, file := range files {
		if max < file.DifferentFrom {
			max = file.DifferentFrom
			index = i
		}
	}
	return max, index
}

func makeBlockChainToSize(utxoIndex *core.UTXOIndex, parentBlk *core.Block, bc *core.Blockchain, size int, d *consensus.Dynasty, keys Keys, addrs []core.Address, wm *client.WalletManager) {

	tailBlk := parentBlk
	for bc.GetMaxHeight() < uint64(size) {
		txs := generateTransactions(utxoIndex, addrs, wm)
		b := generateBlock(utxoIndex, tailBlk, bc, d, keys, txs)
		bc.AddBlockToDb(b)
		tailBlk = b
	}
}

func generateBlock(utxoIndex *core.UTXOIndex, parentBlk *core.Block, bc *core.Blockchain, d *consensus.Dynasty, keys Keys, txs []*core.Transaction) *core.Block {
	producer := core.NewAddress(d.ProducerAtATime(time))
	key := keys.getPrivateKeyByAddress(producer)
	cbtx := core.NewCoinbaseTX(producer, "", bc.GetMaxHeight()+1, common.NewAmount(0))
	txs = append(txs, &cbtx)
	utxoIndex.UpdateUtxo(&cbtx)
	b := core.NewBlockWithTimestamp(txs, parentBlk, time)
	hash := b.CalculateHashWithNonce(0)
	b.SetHash(hash)
	b.SetNonce(0)
	b.SignBlock(key, hash)
	time = time + defaultTimeBetweenBlk
	logger.WithFields(logger.Fields{
		"producer" : producer.String(),
		"timestamp": time,
		"blkHeight" : b.GetHeight(),
	}).Info("Tool:Generating Block...")
	return b
}

func generateFundingBlock(utxoIndex *core.UTXOIndex, parentBlk *core.Block, bc *core.Blockchain, d *consensus.Dynasty, keys Keys, fundAddr core.Address, minerPrivKey string, wm *client.WalletManager) *core.Block{
	logger.Info("generateFundingBlock")
	initFund := uint64(100000)
	initFundAmount := common.NewAmount(initFund)
	minerKeyPair := core.GetKeyPairByString(minerPrivKey)
	pkh,_ := core.NewUserPubKeyHash(minerKeyPair.PublicKey)
	fmt.Println("value is :",utxoIndex.GetAllUTXOsByPubKeyHash(pkh)[0].Value)
	tx := newTransaction(minerKeyPair.GenerateAddress(false), fundAddr, minerKeyPair, utxoIndex, pkh, initFundAmount)
	utxoIndex.UpdateUtxo(tx)
	currBalance[fundAddr.String()] = initFund
	return generateBlock(utxoIndex, parentBlk, bc, d, keys, []*core.Transaction{tx})
}

func generateTransactions(utxoIndex *core.UTXOIndex, addrs []core.Address, wm *client.WalletManager) []*core.Transaction{
	pkhmap := getPubKeyHashes(addrs, wm)
	txs := []*core.Transaction{}
	for i:=0;i< numOfTx;i++{
		tx:=generateTransaction(addrs, wm, utxoIndex, pkhmap)
		utxoIndex.UpdateUtxo(tx)
		txs = append(txs, tx)
	}
	return txs
}

func getPubKeyHashes(addrs []core.Address, wm *client.WalletManager) map[core.Address]core.PubKeyHash{
	res := make(map[core.Address]core.PubKeyHash)
	for _, addr := range addrs {
		wallet := wm.GetWalletByAddress(addr)
		pubKeyHash, _ := core.NewUserPubKeyHash(wallet.GetKeyPair().PublicKey)
		res[addr] = pubKeyHash
	}
	return res
}

func generateTransaction(addrs []core.Address, wm *client.WalletManager, utxoIndex *core.UTXOIndex, pkhmap map[core.Address]core.PubKeyHash) *core.Transaction{
	sender, receiver := getSenderAndReceiver(addrs)
	amount := common.NewAmount(1)
	senderWallet := wm.GetWalletByAddress(sender)
	if senderWallet == nil || len(senderWallet.Addresses) == 0 {
		logger.Panic("Can not find sender wallet")
	}
	tx := newTransaction(sender, receiver, senderWallet.GetKeyPair(), utxoIndex, pkhmap[sender], amount)
	currBalance[sender.String()] -= 1
	currBalance[receiver.String()] += 1

	return tx
}

func newTransaction(sender, receiver core.Address, senderKeyPair *core.KeyPair, utxoIndex *core.UTXOIndex, senderPkh core.PubKeyHash, amount *common.Amount) *core.Transaction{
	utxos, _ := utxoIndex.GetUTXOsByAmount([]byte(senderPkh), amount)

	tx, err := core.NewUTXOTransaction(utxos, sender, receiver, amount, senderKeyPair, common.NewAmount(0), "")

	if err!= nil {
		logger.WithError(err).Panic("Create transaction failed!")
	}

	return &tx
}

func getSenderAndReceiver(addrs []core.Address) (sender,receiver core.Address){
	for i, addr := range addrs{
		if currBalance[addr.String()] > 1000 {
			sender = addr
			if i == len(addrs)-1 {
				receiver = addrs[0]
			}else{
				receiver = addrs[i+1]
			}
			return
		}
	}
	logger.Panic("getSenderAndReceiver failed")
	return
}

func CreateRandomTransactions([]core.Address) []*core.Transaction{
	return nil
}

func CreateWallet(wm *client.WalletManager) []core.Address {

	addresses := wm.GetAddresses()
	numOfWallets := len(addresses)
	for i := numOfWallets; i < maxWallet; i++ {
		_, err := logic.CreateWalletWithpassphrase(password)
		if err != nil {
			logger.WithError(err).Panic("Cannot create new wallet.")
		}
	}

	addresses = wm.GetAddresses()
	logger.WithFields(logger.Fields{
		"addresses": addresses,
	}).Info("Wallets are created")
	return addresses
}

func LoadPrivateKey() Keys {
	jsonFile, err := os.Open("conf/key.json")
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var keys Keys

	json.Unmarshal(byteValue, &keys)

	return keys
}

func (k Keys) getPrivateKeyByAddress(address core.Address) string {
	for _, key := range k.Keys {
		if key.Address == address.Address {
			return key.Key
		}
	}
	return ""
}
