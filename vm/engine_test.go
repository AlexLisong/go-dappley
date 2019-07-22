package vm

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/dappley/go-dappley/client"
	"github.com/dappley/go-dappley/util"

	"github.com/dappley/go-dappley/common"
	"github.com/dappley/go-dappley/core"
	"github.com/dappley/go-dappley/crypto/keystore/secp256k1"
	"github.com/stretchr/testify/assert"
)

var dummyAddr = "dummyAddr"

func TestScEngine_Execute(t *testing.T) {
	script := `'use strict';

var AddrChecker = function(){
	
};

AddrChecker.prototype = {
		check:function(addr,dummy){
    	return Blockchain.verifyAddress(addr)+dummy;
    }
};
module.exports = new AddrChecker();
`

	sc := NewV8Engine()
	sc.ImportSourceCode(script)

	sc.SetExecutionLimits(DefaultLimitsOfGas, DefaultLimitsOfTotalMemorySize)
	ret, _ := sc.Execute("check", "\"dastXXWLe5pxbRYFhcyUq8T3wb5srWkHKa\",34")
	assert.Equal(t, "35", ret)
}

func TestScEngine_Execute_SyntaxError(t *testing.T) {
	// Missing quotes around 'use strict'
	script := `use strict;

var AddrChecker = function(){
	
};

AddrChecker.prototype = {
	check:function(addr,dummy){
    	return Blockchain.verifyAddress(addr)+dummy;
	}
};
module.exports = new AddrChecker();
`

	sc := NewV8Engine()
	sc.ImportSourceCode(script)

	sc.SetExecutionLimits(DefaultLimitsOfGas, DefaultLimitsOfTotalMemorySize)
	_, err := sc.Execute("check", "\"dastXXWLe5pxbRYFhcyUq8T3wb5srWkHKa\",34")
	assert.NotNil(t, err)
}

func TestScEngine_BlockchainTransfer(t *testing.T) {
	script := `'use strict';
var MathTest = function(){};
MathTest.prototype = {
    transfer: function(to, amount, tip){
        return Blockchain.transfer(to, amount, tip);
    }
};
module.exports = new MathTest();`

	contractPubKeyHash := client.NewContractPubKeyHash()
	contractAddr := contractPubKeyHash.GenerateAddress()
	contractUTXOs := []*core.UTXO{
		{
			Txid:     []byte("1"),
			TxIndex:  0,
			TXOutput: *core.NewTxOut(common.NewAmount(0), contractAddr, "somecontract"),
		},
		{
			Txid:     []byte("1"),
			TxIndex:  1,
			TXOutput: *core.NewTxOut(common.NewAmount(15), contractAddr, ""),
		},
		{
			Txid:     []byte("2"),
			TxIndex:  0,
			TXOutput: *core.NewTxOut(common.NewAmount(3), contractAddr, ""),
		},
	}

	sc := NewV8Engine()
	sc.ImportSourceCode(script)
	sc.ImportContractAddr(contractAddr)
	sc.ImportSourceTXID([]byte("thatTX"))
	sc.ImportUTXOs(contractUTXOs)

	sc.SetExecutionLimits(DefaultLimitsOfGas, DefaultLimitsOfTotalMemorySize)
	result, _ := sc.Execute("transfer", "'16PencPNnF8CiSx2EBGEd1axhf7vuHCouj','10','2'")

	assert.Equal(t, "0", result)
	if assert.Equal(t, 1, len(sc.generatedTXs)) {
		if assert.Equal(t, 2, len(sc.generatedTXs[0].Vin)) {
			assert.Equal(t, []byte("1"), sc.generatedTXs[0].Vin[1].Txid)
			assert.Equal(t, 1, sc.generatedTXs[0].Vin[1].Vout)
			assert.Equal(t, []byte("thatTX"), sc.generatedTXs[0].Vin[1].Signature)
			assert.Equal(t, []byte(contractPubKeyHash), sc.generatedTXs[0].Vin[1].PubKey)
		}
		if assert.Equal(t, 2, len(sc.generatedTXs[0].Vout)) {
			// payout
			assert.Equal(t, common.NewAmount(10), sc.generatedTXs[0].Vout[0].Value)
			// change
			assert.Equal(t, common.NewAmount(15-10-2), sc.generatedTXs[0].Vout[1].Value)

			assert.Equal(t, client.NewAddress("16PencPNnF8CiSx2EBGEd1axhf7vuHCouj"), sc.generatedTXs[0].Vout[0].PubKeyHash.GenerateAddress())
			assert.Equal(t, contractPubKeyHash, sc.generatedTXs[0].Vout[1].PubKeyHash)
		}
	}
}

func TestScEngine_StorageGet(t *testing.T) {
	script := `'use strict';

var StorageTest = function(){
	
};

StorageTest.prototype = {
	set:function(key,value){
    	return LocalStorage.set(key,value);
    },
	get:function(key){
    	return LocalStorage.get(key);
    }
};
module.exports = new StorageTest();
`

	ss := core.NewScState()
	ss.GetStorageByAddress(dummyAddr)["key"] = "7"
	sc := NewV8Engine()
	sc.ImportSourceCode(script)
	sc.ImportContractAddr(client.NewAddress(dummyAddr))
	sc.ImportLocalStorage(ss)

	sc.SetExecutionLimits(DefaultLimitsOfGas, DefaultLimitsOfTotalMemorySize)
	ret, _ := sc.Execute("get", "\"key\"")
	assert.Equal(t, "7", ret)
}

func TestScEngine_StorageSet(t *testing.T) {
	script := `'use strict';

var StorageTest = function(){
	
};

StorageTest.prototype = {
	set:function(key,value){
    	return LocalStorage.set(key, value);
    },
	get:function(key){
    	return LocalStorage.get(key);
    },
	setColor: function(key, color){
		var car = {type:"Fiat", model:"500", color:"white"};
		car.color = color;
		return LocalStorage.set(key, car);
	},
	getColor: function(key){
		return LocalStorage.get(key).color;
	}
};
module.exports = new StorageTest();
`
	ss := core.NewScState()
	sc := NewV8Engine()
	sc.ImportSourceCode(script)
	sc.ImportLocalStorage(ss)
	sc.ImportContractAddr(client.NewAddress(dummyAddr))
	sc.SetExecutionLimits(DefaultLimitsOfGas, DefaultLimitsOfTotalMemorySize)

	ret, _ := sc.Execute("set", "\"key\",6")
	assert.Equal(t, "0", ret)
	ret2, _ := sc.Execute("get", "\"key\"")
	assert.Equal(t, "6", ret2)
	ret3, _ := sc.Execute("set", "\"key\",\"abcd\"")
	assert.Equal(t, "0", ret3)
	ret4, _ := sc.Execute("get", "\"key\"")
	assert.Equal(t, "abcd", ret4)
	ret5, _ := sc.Execute("setColor", "\"key\",\"BLACK\"")
	assert.Equal(t, "0", ret5)
	ret6, _ := sc.Execute("getColor", "\"key\"")
	assert.Equal(t, "BLACK", ret6)
}

func TestScEngine_StorageDel(t *testing.T) {
	script := `'use strict';

var StorageTest = function(){
	
};

StorageTest.prototype = {
	set:function(key,value){
		_log.error("Test case in Storage del ", "set")
    	return LocalStorage.set(key,value);
    },
	get:function(key){
    	return LocalStorage.get(key);
    },
	del:function(key){
    	return LocalStorage.del(key);
    }
};
module.exports = new StorageTest();
`
	ss := core.NewScState()
	sc := NewV8Engine()
	sc.ImportSourceCode(script)
	sc.ImportLocalStorage(ss)
	sc.ImportContractAddr(client.NewAddress(dummyAddr))
	sc.SetExecutionLimits(DefaultLimitsOfGas, DefaultLimitsOfTotalMemorySize)
	ret, _ := sc.Execute("set", "\"key\",6")
	assert.Equal(t, "0", ret)
	ret2, _ := sc.Execute("del", "\"key\"")
	assert.Equal(t, "0", ret2)
	ret3, _ := sc.Execute("get", "\"key\"")
	assert.Equal(t, "null", ret3)
}

func TestScEngine_Reward(t *testing.T) {
	script :=
		`'use strict';

var RewardTest = function(){
	
};

RewardTest.prototype = {
	reward:function(addr,amount){
    	return _native_reward.record(addr,amount);
    }
};
module.exports = new RewardTest();
`
	ss := make(map[string]string)
	sc := NewV8Engine()
	sc.ImportSourceCode(script)
	sc.ImportRewardStorage(ss)

	sc.SetExecutionLimits(DefaultLimitsOfGas, DefaultLimitsOfTotalMemorySize)
	ret, _ := sc.Execute("reward", "\"myAddr\",\"8\"")
	assert.Equal(t, "0", ret)
	ret2, _ := sc.Execute("reward", "\"myAddr\",\"9\"")
	assert.Equal(t, "0", ret2)
	assert.Equal(t, "17", ss["myAddr"])
}

func TestScEngine_TransactionTest(t *testing.T) {
	script :=
		`'use strict';

var TransactionTest = function(){
};

TransactionTest.prototype = {
	dump:function(dummy) {
		_log.error("dump")
		_log.error("tx id:", _tx.id)
		_log.error("prevUtxo length:", _prevUtxos.length)
		_log.error("tx vin length:", _tx.vin.length)
		let index = 0
		for (let vin of _tx.vin) {
				_log.error("Vin index:", index, " id:", vin.txid, " vout:", vin.vout, 
				    " signature:", vin.signature, " pubkey:", vin.pubkey)
				let prevUtxo = _prevUtxos[index]
				_log.error("PrevUtxo id:", prevUtxo.txid, " txIndex:", prevUtxo.txIndex, 
				    " value:", prevUtxo.value, " pubkeyhash:", prevUtxo.pubkeyhash, " address:", prevUtxo.address)
	    	}
		_log.error("tx vout length:", _tx.vin.length)
		index = 0
		for (let vout of _tx.vout) {
			_log.error("index:", index, " amount:", vout.amount, " pubkeyhash:", vout.pubkeyhash)
		}
	}
};
module.exports = new TransactionTest();
`
	ss := core.NewScState()
	sc := NewV8Engine()
	sc.ImportSourceCode(script)
	sc.ImportLocalStorage(ss)
	tx := core.MockTransaction()
	sc.ImportTransaction(tx)
	sc.ImportPrevUtxos(core.MockUtxos(tx.Vin))
	sc.SetExecutionLimits(DefaultLimitsOfGas, DefaultLimitsOfTotalMemorySize)
	sc.Execute("dump", "\"dummy\"")
}

func TestStepRecord(t *testing.T) {
	script, _ := ioutil.ReadFile("jslib/step_recorder.js")

	reward := make(map[string]string)
	ss := core.NewScState()
	sc := NewV8Engine()
	sc.ImportSourceCode(string(script))
	sc.ImportLocalStorage(ss)
	sc.ImportContractAddr(client.NewAddress(dummyAddr))
	sc.ImportRewardStorage(reward)

	sc.SetExecutionLimits(DefaultLimitsOfGas, DefaultLimitsOfTotalMemorySize)
	ret, _ := sc.Execute("record", "\"dastXXWLe5pxbRYFhcyUq8T3wb5srWkHKa\", 20")
	assert.Equal(t, "0", ret)
	assert.Equal(t, "20", ss.GetStorageByAddress(client.NewAddress(dummyAddr).String())["dastXXWLe5pxbRYFhcyUq8T3wb5srWkHKa"])
	assert.Equal(t, "20", reward["dastXXWLe5pxbRYFhcyUq8T3wb5srWkHKa"])
	ret2, _ := sc.Execute("record", "\"dastXXWLe5pxbRYFhcyUq8T3wb5srWkHKa\", 15")
	assert.Equal(t, "0", ret2)
	assert.Equal(t, "35", ss.GetStorageByAddress(client.NewAddress(dummyAddr).String())["dastXXWLe5pxbRYFhcyUq8T3wb5srWkHKa"])
	assert.Equal(t, "35", reward["dastXXWLe5pxbRYFhcyUq8T3wb5srWkHKa"])
	ret3, _ := sc.Execute("record", "\"fastXXWLe5pxbRYFhcyUq8T3wb5srWkHKa\", 10")
	assert.Equal(t, "0", ret3)
	assert.Equal(t, "10", ss.GetStorageByAddress(client.NewAddress(dummyAddr).String())["fastXXWLe5pxbRYFhcyUq8T3wb5srWkHKa"])
	assert.Equal(t, "10", reward["fastXXWLe5pxbRYFhcyUq8T3wb5srWkHKa"])
	assert.Equal(t, "35", ss.GetStorageByAddress(client.NewAddress(dummyAddr).String())["dastXXWLe5pxbRYFhcyUq8T3wb5srWkHKa"])
	assert.Equal(t, "35", reward["dastXXWLe5pxbRYFhcyUq8T3wb5srWkHKa"])
}

func TestCrypto_VerifySignature(t *testing.T) {
	script, _ := ioutil.ReadFile("test/test_crypto.js")

	sc := NewV8Engine()
	sc.ImportSourceCode(string(script))

	kp := client.NewKeyPair()
	msg := "hello world dappley"
	privData, _ := secp256k1.FromECDSAPrivateKey(&kp.PrivateKey)
	data := sha256.Sum256([]byte(msg))
	signature, _ := secp256k1.Sign(data[:], privData)

	sc.SetExecutionLimits(DefaultLimitsOfGas, DefaultLimitsOfTotalMemorySize)
	ret, _ := sc.Execute("verifySig",
		fmt.Sprintf("\"%s\", \"%s\", \"%s\"",
			msg,
			hex.EncodeToString(kp.PublicKey),
			hex.EncodeToString(signature),
		),
	)
	assert.Equal(
		t,
		"true",
		ret,
	)
}

func TestCrypto_VerifyPublicKey(t *testing.T) {
	script, _ := ioutil.ReadFile("test/test_crypto.js")

	sc := NewV8Engine()
	sc.ImportSourceCode(string(script))

	kp := client.NewKeyPair()
	fmt.Println(kp.PublicKey)
	pkh, err := client.NewUserPubKeyHash(kp.PublicKey)
	assert.Nil(t, err)
	addr := pkh.GenerateAddress()
	fmt.Println(addr)

	sc.SetExecutionLimits(DefaultLimitsOfGas, DefaultLimitsOfTotalMemorySize)
	ret, _ := sc.Execute("verifyPk",
		fmt.Sprintf("\"%s\", \"%s\"",
			addr,
			hex.EncodeToString(kp.PublicKey),
		),
	)
	assert.Equal(
		t,
		"true",
		ret,
	)
	ret2, _ := sc.Execute("verifyPk",
		fmt.Sprintf("\"%s\", \"%s\"",
			"IncorrectAddress",
			hex.EncodeToString(kp.PublicKey),
		),
	)
	assert.Equal(
		t,
		"false",
		ret2,
	)
}

func TestMath(t *testing.T) {
	script, _ := ioutil.ReadFile("test/test_math.js")

	sc := NewV8Engine()
	sc.ImportSourceCode(string(script))
	sc.ImportSeed(10)

	sc.SetExecutionLimits(DefaultLimitsOfGas, DefaultLimitsOfTotalMemorySize)
	res, _ := sc.Execute("random", "20")
	assert.Equal(t, "14", res)
}

func TestBlkHeight(t *testing.T) {
	script, _ := ioutil.ReadFile("test/test_blockchain.js")

	sc := NewV8Engine()
	sc.ImportSourceCode(string(script))
	sc.ImportCurrBlockHeight(22334)

	sc.SetExecutionLimits(DefaultLimitsOfGas, DefaultLimitsOfTotalMemorySize)
	ret, _ := sc.Execute("getBlkHeight", "")
	assert.Equal(t, "22334", ret)
}

func TestRecordEvent(t *testing.T) {
	script, _ := ioutil.ReadFile("test/test_event.js")

	ss := core.NewScState()
	sc := NewV8Engine()
	sc.ImportLocalStorage(ss)
	sc.ImportSourceCode(string(script))

	sc.SetExecutionLimits(DefaultLimitsOfGas, DefaultLimitsOfTotalMemorySize)
	ret, _ := sc.Execute("trigger", "\"topic\",\"data\"")
	assert.Equal(t, "0", ret)
	assert.Equal(t, "topic", ss.GetEvents()[0].GetTopic())
	assert.Equal(t, "data", ss.GetEvents()[0].GetData())
}

func TestTrimWhiteSpaces(t *testing.T) {
	script, _ := ioutil.ReadFile("test/test_blockchain.js")
	scriptStr := string(script)
	str := strings.Replace(scriptStr, " ", "", -1)
	str = strings.Replace(str, "\\n", "", -1)
	fmt.Println(str)
}

func TestGetNodeAddress(t *testing.T) {
	script, _ := ioutil.ReadFile("test/test_blockchain.js")

	sc := NewV8Engine()
	sc.ImportSourceCode(string(script))
	sc.ImportNodeAddress(client.NewAddress("testAddr"))

	sc.SetExecutionLimits(DefaultLimitsOfGas, DefaultLimitsOfTotalMemorySize)
	ret, _ := sc.Execute("getNodeAddress", "")
	assert.Equal(t, "testAddr", ret)
}

func TestNewAddress(t *testing.T) {
	kp := client.NewKeyPair()
	privData, _ := secp256k1.FromECDSAPrivateKey(&kp.PrivateKey)
	pk := hex.EncodeToString(privData)
	publicKey := hex.EncodeToString(kp.PublicKey)
	pkh, _ := client.NewUserPubKeyHash(kp.PublicKey)
	addr := pkh.GenerateAddress()
	fmt.Println("privatekey:", pk)
	fmt.Println("publickey:", publicKey)
	fmt.Println("addr:", addr)
}

func TestAddGasCount(t *testing.T) {
	vout := core.NewContractTXOutput(client.NewAddress("cd9N6MRsYxU1ToSZjLnqFhTb66PZcePnAD"), "{\"function\":\"add\",\"args\":[\"1\",\"3\"]}")
	tx := core.Transaction{
		Vout: []core.TXOutput{*vout},
	}
	ctx := tx.ToContractTx()
	script, _ := ioutil.ReadFile("test/test_add.js")

	sc := NewV8Engine()
	sc.ImportSourceCode(string(script))

	sc.SetExecutionLimits(DefaultLimitsOfGas, DefaultLimitsOfTotalMemorySize)

	function, args := util.DecodeScInput(vout.Contract)
	assert.NotEmpty(t, function)

	totalArgs := util.PrepareArgs(args)
	_, err := sc.Execute("add", totalArgs)
	assert.Nil(t, err)

	gasCount := sc.ExecutionInstructions()
	// record base gas
	baseGas, _ := ctx.GasCountOfTxBase()
	gasCount += baseGas.Uint64()

	// min gas of each tx
	minGasTx := core.MinGasCountPerTransaction.Uint64()
	// dataLen
	dataGas := uint64(35)
	// instruction gas
	instructionGas := uint64(25)
	assert.Equal(t, minGasTx+dataGas+instructionGas, gasCount)
}

func TestStepRecordGasCount(t *testing.T) {
	vout := core.NewContractTXOutput(client.NewAddress("cd9N6MRsYxU1ToSZjLnqFhTb66PZcePnAD"),
		"{\"function\":\"record\",\"args\":[\"dYgmFyXLg5jSfbysWoZF7Zimnx95xg77Qo\",\"2000\"]}")
	tx := core.Transaction{
		Vout: []core.TXOutput{*vout},
	}
	ctx := tx.ToContractTx()
	script, _ := ioutil.ReadFile("test/test_step_recorder.js")

	ss := core.NewScState()
	sc := NewV8Engine()
	sc.ImportLocalStorage(ss)
	sc.ImportSourceCode(string(script))

	sc.SetExecutionLimits(DefaultLimitsOfGas, DefaultLimitsOfTotalMemorySize)

	function, args := util.DecodeScInput(vout.Contract)
	assert.NotEmpty(t, function)

	totalArgs := util.PrepareArgs(args)
	_, err := sc.Execute("record", totalArgs)
	assert.Nil(t, err)

	gasCount := sc.ExecutionInstructions()
	// record base gas
	baseGas, _ := ctx.GasCountOfTxBase()
	gasCount += baseGas.Uint64()

	// min gas of each tx
	minGasTx := core.MinGasCountPerTransaction.Uint64()
	// dataLen
	dataGas := uint64(74)
	// instruction gas
	instructionGas := uint64(61)
	assert.Equal(t, minGasTx+dataGas+instructionGas, gasCount)
}