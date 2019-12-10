package vm

import "C"
import (
		"unsafe"
		logger "github.com/sirupsen/logrus"
)


const (
	PRODUCER_ADDR = "producer_address"

	PRODUCER_ADD_INDEX  = "producer_add"
	PRODUCER_DEL_INDEX = "producer_del"

	ADD = "ADD"
	DEL = "DEL"
)

//export ProducerFunc
func ProducerFunc(address unsafe.Pointer, option *C.char , nodeAddr *C.char) bool {
	addr := uint64(uintptr(address))
	goNodeAddr := C.GoString(nodeAddr)
	goOption := C.GoString(option)

	logger.Infof("ProducerFunc comming,option: %v, goNodeAddr: %v", goOption, goNodeAddr)

	engine := getV8EngineByAddress(addr)
	if engine == nil {
		logger.WithFields(logger.Fields{
			"contract_address": addr,
			"nodeAddr":              goNodeAddr,
			"goOption":              goOption,
		}).Debug("SmartContract: failed to get state handler!")
		return false
	}

	switch goOption {
	case ADD:
		engine.state.GetStorageByAddress(PRODUCER_ADDR)[PRODUCER_ADD_INDEX] = goNodeAddr
	case DEL:
		engine.state.GetStorageByAddress(PRODUCER_ADDR)[PRODUCER_DEL_INDEX] = goNodeAddr
	}

	return true
}
