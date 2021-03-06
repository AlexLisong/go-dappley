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

package core

import (
	"github.com/dappley/go-dappley/core/account"
	"github.com/dappley/go-dappley/core/scState"
	"github.com/dappley/go-dappley/core/transaction"
	"github.com/dappley/go-dappley/core/utxo"
)

type ScEngineManager interface {
	CreateEngine() ScEngine
	RunScheduledEvents(contractUtxo []*utxo.UTXO, scStorage *scState.ScState, blkHeight uint64, seed int64)
}

type ScEngine interface {
	DestroyEngine()
	ImportSourceCode(source string)
	ImportLocalStorage(state *scState.ScState)
	ImportContractAddr(contractAddr account.Address)
	ImportSourceTXID(txid []byte)
	ImportUTXOs(utxos []*utxo.UTXO)
	ImportRewardStorage(rewards map[string]string)
	ImportTransaction(tx *transaction.Transaction)
	ImportContractCreateUTXO(utxo *utxo.UTXO)
	ImportPrevUtxos(utxos []*utxo.UTXO)
	ImportCurrBlockHeight(currBlkHeight uint64)
	ImportSeed(seed int64)
	ImportNodeAddress(addr account.Address)
	GetGeneratedTXs() []*transaction.Transaction
	Execute(function, args string) (string, error)
	SetExecutionLimits(uint64, uint64) error
	ExecutionInstructions() uint64
	CheckContactSyntax(source string) error
}
