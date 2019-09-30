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

package ltransactionpool

import (
	"bytes"
	"encoding/hex"
	"sort"
	"sync"

	"github.com/dappley/go-dappley/core/transaction"
	transactionpb "github.com/dappley/go-dappley/core/transaction/pb"
	"github.com/dappley/go-dappley/core/transactionpool"
	transactionPoolpb "github.com/dappley/go-dappley/core/transactionpool/pb"
	"github.com/dappley/go-dappley/logic/ltransaction"
	"github.com/dappley/go-dappley/logic/lutxo"

	"github.com/asaskevich/EventBus"
	"github.com/dappley/go-dappley/common/pubsub"

	"github.com/dappley/go-dappley/network/networkmodel"
	"github.com/dappley/go-dappley/storage"
	"github.com/golang-collections/collections/stack"
	"github.com/golang/protobuf/proto"
	logger "github.com/sirupsen/logrus"
)

const (
	NewTransactionTopic   = "NewTransaction"
	EvictTransactionTopic = "EvictTransaction"

	TxPoolDbKey = "txpool"

	BroadcastTx       = "BroadcastTx"
	BroadcastBatchTxs = "BraodcastBatchTxs"
)

var (
	txPoolSubscribedTopics = []string{
		BroadcastTx,
		BroadcastBatchTxs,
	}
)

type TransactionPoolLogic struct {
	txpool     *transactionpool.TransactionPool
	pendingTxs []*transaction.Transaction
	sizeLimit  uint32
	EventBus   EventBus.Bus
	mutex      sync.RWMutex
	netService NetService
}

func NewTransactionPoolLogic(netService NetService, limit uint32) *TransactionPoolLogic {
	tp := transactionpool.NewTransactionPool()
	txPool := &TransactionPoolLogic{
		txpool:     tp,
		pendingTxs: make([]*transaction.Transaction, 0),
		sizeLimit:  limit,
		EventBus:   EventBus.New(),
		mutex:      sync.RWMutex{},
		netService: netService,
	}
	txPool.ListenToNetService()
	return txPool
}

func (txPool *TransactionPoolLogic) GetTipOrder() []string { return txPool.txpool.TipOrder }

func (txPool *TransactionPoolLogic) ListenToNetService() {
	if txPool.netService == nil {
		return
	}

	txPool.netService.Listen(txPool)
}

func (txPool *TransactionPoolLogic) GetSubscribedTopics() []string {
	return txPoolSubscribedTopics
}

func (txPool *TransactionPoolLogic) GetTopicHandler(topic string) pubsub.TopicHandler {

	switch topic {
	case BroadcastTx:
		return txPool.BroadcastTxHandler
	case BroadcastBatchTxs:
		return txPool.BroadcastBatchTxsHandler
	}
	return nil
}

func (txPool *TransactionPoolLogic) DeepCopy() *TransactionPoolLogic {
	tp := transactionpool.NewTransactionPool()
	txPoolCopy := TransactionPoolLogic{
		txpool:    tp,
		sizeLimit: txPool.sizeLimit,
		EventBus:  EventBus.New(),
		mutex:     sync.RWMutex{},
	}

	copy(txPoolCopy.txpool.TipOrder, txPool.txpool.TipOrder)

	for key, tx := range txPool.txpool.Txs {
		newTx := tx.Value.DeepCopy()
		newTxNode := transaction.NewTransactionNode(&newTx)

		for childKey, childTx := range tx.Children {
			newTxNode.Children[childKey] = childTx
		}
		txPoolCopy.txpool.Txs[key] = newTxNode
	}

	return &txPoolCopy
}

func (txPool *TransactionPoolLogic) SetSizeLimit(sizeLimit uint32) {
	txPool.sizeLimit = sizeLimit
}

func (txPool *TransactionPoolLogic) GetSizeLimit() uint32 {
	return txPool.sizeLimit
}

func (txPool *TransactionPoolLogic) GetTransactions() []*transaction.Transaction {
	txPool.mutex.RLock()
	defer txPool.mutex.RUnlock()
	return txPool.getSortedTransactions()
}

func (txPool *TransactionPoolLogic) GetNumOfTxInPool() int {
	txPool.mutex.RLock()
	defer txPool.mutex.RUnlock()

	return len(txPool.txpool.Txs)
}

func (txPool *TransactionPoolLogic) ResetPendingTransactions() {
	txPool.mutex.Lock()
	defer txPool.mutex.Unlock()

	txPool.pendingTxs = make([]*transaction.Transaction, 0)
}

func (txPool *TransactionPoolLogic) GetAllTransactions() []*transaction.Transaction {
	txPool.mutex.RLock()
	defer txPool.mutex.RUnlock()

	txs := []*transaction.Transaction{}
	for _, tx := range txPool.pendingTxs {
		txs = append(txs, tx)
	}

	for _, tx := range txPool.getSortedTransactions() {
		txs = append(txs, tx)
	}

	return txs
}

//PopTransactionWithMostTips pops the transactions with the most tips
func (txPool *TransactionPoolLogic) PopTransactionWithMostTips(utxoIndex *lutxo.UTXOIndex) *transaction.TransactionNode {
	txPool.mutex.Lock()
	defer txPool.mutex.Unlock()

	txNode := txPool.getMaxTipTransaction()
	tempUtxoIndex := utxoIndex.DeepCopy()
	if txNode == nil {
		return txNode
	}
	//remove the transaction from tip order
	txPool.txpool.TipOrder = txPool.txpool.TipOrder[1:]

	if err := ltransaction.VerifyTransaction(tempUtxoIndex, txNode.Value, 0); err == nil {
		txPool.insertChildrenIntoSortedWaitlist(txNode)
		txPool.removeTransaction(txNode)
	} else {
		logger.WithError(err).Warn("Transaction Pool: Pop max tip transaction failed!")
		txPool.removeTransactionNodeAndChildren(txNode.Value)
		return nil
	}

	txPool.pendingTxs = append(txPool.pendingTxs, txNode.Value)
	return txNode
}

//Rollback adds a popped transaction back to the transaction pool. The existing transactions in txpool may be dependent on the input transactionbase. However, the input transaction should never be dependent on any transaction in the current pool
func (txPool *TransactionPoolLogic) Rollback(tx transaction.Transaction) {
	txPool.mutex.Lock()
	defer txPool.mutex.Unlock()

	rollbackTxNode := transaction.NewTransactionNode(&tx)
	txPool.updateChildren(rollbackTxNode)
	newTipOrder := []string{}

	for _, txid := range txPool.txpool.TipOrder {
		if _, exist := rollbackTxNode.Children[txid]; !exist {
			newTipOrder = append(newTipOrder, txid)
		}
	}

	txPool.txpool.TipOrder = newTipOrder

	txPool.addTransaction(rollbackTxNode)
	txPool.insertIntoTipOrder(rollbackTxNode)

}

//updateChildren traverses through all transactions in transaction pool and find the input node's children
func (txPool *TransactionPoolLogic) updateChildren(node *transaction.TransactionNode) {
	for txid, txNode := range txPool.txpool.Txs {
	loop:
		for _, vin := range txNode.Value.Vin {
			if bytes.Compare(vin.Txid, node.Value.ID) == 0 {
				node.Children[txid] = txNode.Value
				break loop
			}
		}
	}
}

//Push pushes a new transaction into the pool
func (txPool *TransactionPoolLogic) Push(tx transaction.Transaction) {
	txPool.mutex.Lock()
	defer txPool.mutex.Unlock()
	if txPool.sizeLimit == 0 {
		logger.Warn("TransactionPool: transaction is not pushed to pool because sizeLimit is set to 0.")
		return
	}

	txNode := transaction.NewTransactionNode(&tx)

	if txPool.txpool.CurrSize != 0 && txPool.txpool.CurrSize+uint32(txNode.Size) >= txPool.sizeLimit {
		logger.WithFields(logger.Fields{
			"sizeLimit": txPool.sizeLimit,
		}).Warn("TransactionPool: is full.")

		return
	}

	txPool.addTransactionAndSort(txNode)

}

//CleanUpMinedTxs updates the transaction pool when a new block is added to the blockchain.
//It removes the packed transactions from the txpool while keeping their children
func (txPool *TransactionPoolLogic) CleanUpMinedTxs(minedTxs []*transaction.Transaction) {
	txPool.mutex.Lock()
	defer txPool.mutex.Unlock()

	for _, tx := range minedTxs {

		txNode, ok := txPool.txpool.Txs[hex.EncodeToString(tx.ID)]
		if !ok {
			continue
		}
		txPool.insertChildrenIntoSortedWaitlist(txNode)
		txPool.removeTransaction(txNode)
		txPool.removeFromTipOrder(tx.ID)
	}
}

func (txPool *TransactionPoolLogic) removeFromTipOrder(txID []byte) {
	key := hex.EncodeToString(txID)

	for index, value := range txPool.txpool.TipOrder {
		if value == key {
			txPool.txpool.TipOrder = append(txPool.txpool.TipOrder[:index], txPool.txpool.TipOrder[index+1:]...)
			return
		}
	}

}

func (txPool *TransactionPoolLogic) cleanUpTxSort() {
	newTxOrder := []string{}
	for _, txid := range txPool.txpool.TipOrder {
		if _, ok := txPool.txpool.Txs[txid]; ok {
			newTxOrder = append(newTxOrder, txid)
		}
	}
	txPool.txpool.TipOrder = newTxOrder
}

func (txPool *TransactionPoolLogic) getSortedTransactions() []*transaction.Transaction {

	nodes := make(map[string]*transaction.TransactionNode)
	scDeploymentTxExists := make(map[string]bool)

	for key, node := range txPool.txpool.Txs {
		nodes[key] = node
		ctx := node.Value.ToContractTx()
		if ctx != nil && !ctx.IsExecutionContract() {
			scDeploymentTxExists[ctx.GetContractPubKeyHash().GenerateAddress().String()] = true
		}
	}

	var sortedTxs []*transaction.Transaction
	for len(nodes) > 0 {
		for key, node := range nodes {
			if !checkDependTxInMap(node.Value, nodes) {
				ctx := node.Value.ToContractTx()
				if ctx != nil {
					ctxPkhStr := ctx.GetContractPubKeyHash().GenerateAddress().String()
					if ctx.IsExecutionContract() {
						if !scDeploymentTxExists[ctxPkhStr] {
							sortedTxs = append(sortedTxs, node.Value)
							delete(nodes, key)
						}
					} else {
						sortedTxs = append(sortedTxs, node.Value)
						delete(nodes, key)
						scDeploymentTxExists[ctxPkhStr] = false
					}
				} else {
					sortedTxs = append(sortedTxs, node.Value)
					delete(nodes, key)
				}
			}
		}
	}

	return sortedTxs
}

func checkDependTxInMap(tx *transaction.Transaction, existTxs map[string]*transaction.TransactionNode) bool {
	for _, vin := range tx.Vin {
		if _, exist := existTxs[hex.EncodeToString(vin.Txid)]; exist {
			return true
		}
	}
	return false
}

func (txPool *TransactionPoolLogic) GetTransactionById(txid []byte) *transaction.Transaction {
	txPool.mutex.RLock()
	defer txPool.mutex.RUnlock()
	txNode, ok := txPool.txpool.Txs[hex.EncodeToString(txid)]
	if !ok {
		return nil
	}
	return txNode.Value
}

func (txPool *TransactionPoolLogic) getDependentTxs(txNode *transaction.TransactionNode) map[string]*transaction.TransactionNode {

	toRemoveTxs := make(map[string]*transaction.TransactionNode)
	toCheckTxs := []*transaction.TransactionNode{txNode}

	for len(toCheckTxs) > 0 {
		currentTxNode := toCheckTxs[0]
		toCheckTxs = toCheckTxs[1:]
		for key, _ := range currentTxNode.Children {
			toCheckTxs = append(toCheckTxs, txPool.txpool.Txs[key])
		}
		toRemoveTxs[hex.EncodeToString(currentTxNode.Value.ID)] = currentTxNode
	}

	return toRemoveTxs
}

// The param toRemoveTxs must be calculated by function getDependentTxs
func (txPool *TransactionPoolLogic) removeSelectedTransactions(toRemoveTxs map[string]*transaction.TransactionNode) {
	for _, txNode := range toRemoveTxs {
		txPool.removeTransactionNodeAndChildren(txNode.Value)
	}
}

//removeTransactionNodeAndChildren removes the txNode from tx pool and all its children.
//Note: this function does not remove the node from tipOrder!
func (txPool *TransactionPoolLogic) removeTransactionNodeAndChildren(tx *transaction.Transaction) {

	txStack := stack.New()
	txStack.Push(hex.EncodeToString(tx.ID))
	for txStack.Len() > 0 {
		txid := txStack.Pop().(string)
		currTxNode, ok := txPool.txpool.Txs[txid]
		if !ok {
			continue
		}
		for _, child := range currTxNode.Children {
			txStack.Push(hex.EncodeToString(child.ID))
		}
		txPool.removeTransaction(currTxNode)
	}
}

//removeTransactionNodeAndChildren removes the txNode from tx pool.
//Note: this function does not remove the node from tipOrder!
func (txPool *TransactionPoolLogic) removeTransaction(txNode *transaction.TransactionNode) {
	txPool.disconnectFromParent(txNode.Value)
	txPool.EventBus.Publish(EvictTransactionTopic, txNode.Value)
	txPool.txpool.CurrSize -= uint32(txNode.Size)
	MetricsTransactionPoolSize.Dec(1)
	delete(txPool.txpool.Txs, hex.EncodeToString(txNode.Value.ID))
}

//disconnectFromParent removes itself from its parent's node's children field
func (txPool *TransactionPoolLogic) disconnectFromParent(tx *transaction.Transaction) {
	for _, vin := range tx.Vin {
		if parentTx, exist := txPool.txpool.Txs[hex.EncodeToString(vin.Txid)]; exist {
			delete(parentTx.Children, hex.EncodeToString(tx.ID))
		}
	}
}

func (txPool *TransactionPoolLogic) removeMinTipTx() {
	minTipTx := txPool.getMinTipTransaction()
	if minTipTx == nil {
		return
	}
	txPool.removeTransactionNodeAndChildren(minTipTx.Value)
	txPool.txpool.TipOrder = txPool.txpool.TipOrder[:len(txPool.txpool.TipOrder)-1]
}

func (txPool *TransactionPoolLogic) addTransactionAndSort(txNode *transaction.TransactionNode) {
	isDependentOnParent := false
	for _, vin := range txNode.Value.Vin {
		parentTx, exist := txPool.txpool.Txs[hex.EncodeToString(vin.Txid)]
		if exist {
			parentTx.Children[hex.EncodeToString(txNode.Value.ID)] = txNode.Value
			isDependentOnParent = true
		}
	}

	txPool.addTransaction(txNode)

	txPool.EventBus.Publish(NewTransactionTopic, txNode.Value)

	//if it depends on another tx in txpool, the transaction will be not be included in the sorted list
	if isDependentOnParent {
		return
	}

	txPool.insertIntoTipOrder(txNode)
}

func (txPool *TransactionPoolLogic) addTransaction(txNode *transaction.TransactionNode) {
	txPool.txpool.Txs[hex.EncodeToString(txNode.Value.ID)] = txNode
	txPool.txpool.CurrSize += uint32(txNode.Size)
	MetricsTransactionPoolSize.Inc(1)
}

func (txPool *TransactionPoolLogic) insertChildrenIntoSortedWaitlist(txNode *transaction.TransactionNode) {
	for _, child := range txNode.Children {
		parentTxidsInTxPool := txPool.GetParentTxidsInTxPool(child)
		if len(parentTxidsInTxPool) == 1 {
			txPool.insertIntoTipOrder(txPool.txpool.Txs[hex.EncodeToString(child.ID)])
		}
	}
}

func (txPool *TransactionPoolLogic) GetParentTxidsInTxPool(tx *transaction.Transaction) []string {
	txids := []string{}
	for _, vin := range tx.Vin {
		txidStr := hex.EncodeToString(vin.Txid)
		if _, exist := txPool.txpool.Txs[txidStr]; exist {
			txids = append(txids, txidStr)
		}
	}
	return txids
}

//insertIntoTipOrder insert a transaction into txSort based on tip.
//If the transaction is a child of another transaction, the transaction will NOT be inserted
func (txPool *TransactionPoolLogic) insertIntoTipOrder(txNode *transaction.TransactionNode) {
	index := sort.Search(len(txPool.txpool.TipOrder), func(i int) bool {
		if txPool.txpool.Txs[txPool.txpool.TipOrder[i]] == nil {
			logger.WithFields(logger.Fields{
				"txid":             txPool.txpool.TipOrder[i],
				"len_of_tip_order": len(txPool.txpool.TipOrder),
				"len_of_txs":       len(txPool.txpool.Txs),
			}).Warn("TransactionPool: the transaction in tip order does not exist in txs!")
			return false
		}
		if txPool.txpool.Txs[txPool.txpool.TipOrder[i]].Value == nil {
			logger.WithFields(logger.Fields{
				"txid":             txPool.txpool.TipOrder[i],
				"len_of_tip_order": len(txPool.txpool.TipOrder),
				"len_of_txs":       len(txPool.txpool.Txs),
			}).Warn("TransactionPool: the transaction in tip order does not exist in txs!")
		}
		return txPool.txpool.Txs[txPool.txpool.TipOrder[i]].GetTipsPerByte().Cmp(txNode.GetTipsPerByte()) == -1
	})

	txPool.txpool.TipOrder = append(txPool.txpool.TipOrder, "")
	copy(txPool.txpool.TipOrder[index+1:], txPool.txpool.TipOrder[index:])
	txPool.txpool.TipOrder[index] = hex.EncodeToString(txNode.Value.ID)
}

func deserializeTxPool(d []byte) *TransactionPoolLogic {

	txPoolProto := &transactionPoolpb.TransactionPool{}
	err := proto.Unmarshal(d, txPoolProto)
	if err != nil {
		println(err)
		logger.WithError(err).Panic("TxPool: failed to deserialize TxPool transactions.")
	}
	txpoollogic := NewTransactionPoolLogic(nil, 1)
	txPool := transactionpool.NewTransactionPool()
	txPool.FromProto(txPoolProto)
	txpoollogic.txpool = txPool
	MetricsTransactionPoolSize.Clear()
	MetricsTransactionPoolSize.Inc(int64(len(txpoollogic.txpool.Txs)))
	return txpoollogic
}

func LoadTxPoolFromDatabase(db storage.Storage, netService NetService, txPoolSize uint32) *TransactionPoolLogic {
	rawBytes, err := db.Get([]byte(TxPoolDbKey))
	if err != nil && err.Error() == storage.ErrKeyInvalid.Error() || len(rawBytes) == 0 {
		return NewTransactionPoolLogic(netService, txPoolSize)
	}
	txPool := deserializeTxPool(rawBytes)
	txPool.sizeLimit = txPoolSize
	txPool.netService = netService
	return txPool
}

func (txPool *TransactionPoolLogic) serialize() []byte {

	rawBytes, err := proto.Marshal(txPool.txpool.ToProto())
	if err != nil {
		logger.WithError(err).Panic("TxPool: failed to serialize TxPool transactions.")
	}
	return rawBytes
}

func (txPool *TransactionPoolLogic) SaveToDatabase(db storage.Storage) error {
	txPool.mutex.Lock()
	defer txPool.mutex.Unlock()
	return db.Put([]byte(TxPoolDbKey), txPool.serialize())
}

//getMinTipTransaction gets the transaction.TransactionNode with minimum tip
func (txPool *TransactionPoolLogic) getMaxTipTransaction() *transaction.TransactionNode {
	txid := txPool.getMaxTipTxid()
	if txid == "" {
		return nil
	}
	for txPool.txpool.Txs[txid] == nil {
		logger.WithFields(logger.Fields{
			"txid": txid,
		}).Warn("TransactionPool: max tip transaction is not found in pool")
		txPool.txpool.TipOrder = txPool.txpool.TipOrder[1:]
		txid = txPool.getMaxTipTxid()
		if txid == "" {
			return nil
		}
	}
	return txPool.txpool.Txs[txid]
}

//getMinTipTransaction gets the transaction.TransactionNode with minimum tip
func (txPool *TransactionPoolLogic) getMinTipTransaction() *transaction.TransactionNode {
	txid := txPool.getMinTipTxid()
	if txid == "" {
		return nil
	}
	return txPool.txpool.Txs[txid]
}

//getMinTipTxid gets the txid of the transaction with minimum tip
func (txPool *TransactionPoolLogic) getMaxTipTxid() string {
	if len(txPool.txpool.TipOrder) == 0 {
		logger.Warn("TransactionPool: nothing in the tip order")
		return ""
	}
	return txPool.txpool.TipOrder[0]
}

//getMinTipTxid gets the txid of the transaction with minimum tip
func (txPool *TransactionPoolLogic) getMinTipTxid() string {
	if len(txPool.txpool.TipOrder) == 0 {
		return ""
	}
	return txPool.txpool.TipOrder[len(txPool.txpool.TipOrder)-1]
}

func (txPool *TransactionPoolLogic) BroadcastTx(tx *transaction.Transaction) {
	txPool.netService.BroadcastNormalPriorityCommand(BroadcastTx, tx.ToProto())
}

func (txPool *TransactionPoolLogic) BroadcastTxHandler(input interface{}) {

	var command *networkmodel.DappRcvdCmdContext
	command = input.(*networkmodel.DappRcvdCmdContext)

	//TODO: Check if the blockchain state is ready
	txpb := &transactionpb.Transaction{}

	if err := proto.Unmarshal(command.GetData(), txpb); err != nil {
		logger.Warn(err)
	}

	tx := &transaction.Transaction{}
	tx.FromProto(txpb)
	//TODO: Check if the transaction is generated from running a smart contract
	//utxoIndex := lutxo.NewUTXOIndex(n.GetBlockchain().GetUtxoCache())
	//if tx.IsFromContract(utxoIndex) {
	//	return
	//}

	tx.CreateTime = -1
	txPool.Push(*tx)

	if command.IsBroadcast() {
		//relay the original command
		txPool.netService.Relay(command.GetCommand(), networkmodel.PeerInfo{}, networkmodel.NormalPriorityCommand)
	}
}

func (txPool *TransactionPoolLogic) BroadcastBatchTxs(txs []transaction.Transaction) {

	if len(txs) == 0 {
		return
	}

	transactions := transaction.NewTransactions(txs)

	txPool.netService.BroadcastNormalPriorityCommand(BroadcastBatchTxs, transactions.ToProto())
}

func (txPool *TransactionPoolLogic) BroadcastBatchTxsHandler(input interface{}) {

	var command *networkmodel.DappRcvdCmdContext
	command = input.(*networkmodel.DappRcvdCmdContext)

	//TODO: Check if the blockchain state is ready
	txspb := &transactionpb.Transactions{}

	if err := proto.Unmarshal(command.GetData(), txspb); err != nil {
		logger.Warn(err)
	}

	txs := &transaction.Transactions{}

	//load the tx with proto
	txs.FromProto(txspb)

	for _, tx := range txs.GetTransactions() {
		//TODO: Check if the transaction is generated from running a smart contract
		//utxoIndex := lutxo.NewUTXOIndex(n.GetBlockchain().GetUtxoCache())
		//if tx.IsFromContract(utxoIndex) {
		//	return
		//}
		tx.CreateTime = -1
		txPool.Push(tx)
	}

	if command.IsBroadcast() {
		//relay the original command
		txPool.netService.Relay(command.GetCommand(), networkmodel.PeerInfo{}, networkmodel.NormalPriorityCommand)
	}

}
