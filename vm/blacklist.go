package vm

import "C"
import (
	"github.com/dappley/go-dappley/core/account"
	logger "github.com/sirupsen/logrus"
	"strings"
	"unsafe"
)

const (
	PRODUCER_ADDRESS_COLLECTION_STR = "producer_address_collection_str"
	PRODUCER_ADDRESS_COLLECTION_MAP  = "producer_address_collection_map"

	BLACKLIST = "blacklist_map"
	EXIST =  "exist"

	BLACKLIST_ADD = "ADD"
	BLACKLIST_DEL = "DEL"
)

//export BlacklistFunc
func BlacklistFunc(address unsafe.Pointer, option *C.char , blackAddr *C.char) bool {
	addr := uint64(uintptr(address))
	goBlackAddr := C.GoString(blackAddr)
	goOption := C.GoString(option)
	logger.Infof("BlacklistFunc comming,option: %v, goNodeAddr: %v", goOption, goBlackAddr)

	engine := getV8EngineByAddress(addr)
	if engine == nil {
		logger.WithFields(logger.Fields{
			"contract_address": addr,
			"blackAddr":              goBlackAddr,
			"goOption":              goOption,
		}).Debug("SmartContract: failed to get state handler!")
		return false
	}

	walletAddress := account.GenerateAddress(engine.tx.Vin[0].PubKey)
	producerStr := engine.state.GetStorageByAddress(PRODUCER_ADDRESS_COLLECTION_MAP)[PRODUCER_ADDRESS_COLLECTION_STR]
	producerArray := strings.Split(producerStr,";")
	logger.Infof("producerArray is %v", producerArray)

	bExist := true
	for _, v := range producerArray{
		if v == walletAddress.String(){
			bExist = true
			break
		}
	}

	if bExist{
		switch goOption {
		case BLACKLIST_ADD:
			engine.state.GetStorageByAddress(BLACKLIST)[goBlackAddr] = EXIST
		case BLACKLIST_DEL:
			delete(engine.state.GetStorageByAddress(BLACKLIST), goBlackAddr)
		}
	}else {
		logger.WithFields(logger.Fields{
			"contract_address": addr,
			"blackAddr":              goBlackAddr,
			"goOption":              goOption,
			"walletAddress": walletAddress.String(),
		}).Error("wallet address is invalid!")
		return false
	}
	return true
}
