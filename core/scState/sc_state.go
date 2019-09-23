package scState

import (
	"sync"

	"github.com/dappley/go-dappley/common/hash"

	scstatepb "github.com/dappley/go-dappley/core/scState/pb"
	"github.com/golang/protobuf/proto"
	logger "github.com/sirupsen/logrus"
)

type ChangeLog struct {
	Log map[string]map[string]string
}

type ScState struct {
	states map[string]map[string]string
	events []*Event
	mutex  *sync.RWMutex
}

const (
	scStateLogKey = "scLog"
	scStateMapKey = "scState"
)

func NewChangeLog() *ChangeLog {
	return &ChangeLog{make(map[string]map[string]string)}
}

func NewScState() *ScState {
	return &ScState{make(map[string]map[string]string), make([]*Event, 0), &sync.RWMutex{}}
}

func (ss *ScState) GetEvents() []*Event                     { return ss.events }
func (ss *ScState) GetStates() map[string]map[string]string { return ss.states }

func (ss *ScState) RecordEvent(event *Event) {
	ss.events = append(ss.events, event)
}

func DeserializeScState(d []byte) *ScState {
	scStateProto := &scstatepb.ScState{}
	err := proto.Unmarshal(d, scStateProto)
	if err != nil {
		logger.WithError(err).Panic("ScState: failed to deserialize UTXO states.")
	}
	ss := NewScState()
	ss.FromProto(scStateProto)
	return ss
}

func (ss *ScState) Serialize() []byte {
	rawBytes, err := proto.Marshal(ss.ToProto())
	if err != nil {
		logger.WithError(err).Panic("ScState: failed to serialize UTXO states.")
	}
	return rawBytes
}

//Get gets an item in scStorage
func (ss *ScState) Get(pubKeyHash, key string) string {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()
	if len(ss.states[pubKeyHash]) == 0 {
		return ""
	}
	return ss.states[pubKeyHash][key]
}

//Set sets an item in scStorage
func (ss *ScState) Set(pubKeyHash, key, value string) int {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	if len(ss.states[pubKeyHash]) == 0 {
		ls := make(map[string]string)
		ss.states[pubKeyHash] = ls
	}
	ss.states[pubKeyHash][key] = value
	return 0
}

//Del deletes an item in scStorage
func (ss *ScState) Del(pubKeyHash, key string) int {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()
	if len(ss.states[pubKeyHash]) == 0 {
		return 1
	}
	if ss.states[pubKeyHash][key] == "" {
		return 1
	}

	delete(ss.states[pubKeyHash], key)
	return 0
}

//GetStorageByAddress gets a storage map by address
func (ss *ScState) GetStorageByAddress(address string) map[string]string {
	if len(ss.states[address]) == 0 {
		//initializes the map with dummy data
		ss.states[address] = map[string]string{"init": "i"}
	}
	return ss.states[address]
}

func GetScStateKey(blkHash hash.Hash) []byte {
	return []byte(scStateMapKey + blkHash.String())
}

func (ss *ScState) ToProto() proto.Message {
	scState := make(map[string]*scstatepb.State)

	for key, val := range ss.states {
		scState[key] = &scstatepb.State{State: val}
	}
	return &scstatepb.ScState{States: scState}
}

func (ss *ScState) FromProto(pb proto.Message) {
	for key, val := range pb.(*scstatepb.ScState).States {
		ss.states[key] = val.State
	}
}

func (cl *ChangeLog) ToProto() proto.Message {
	changelog := make(map[string]*scstatepb.Log)

	for key, val := range cl.Log {
		changelog[key] = &scstatepb.Log{Log: val}
	}
	return &scstatepb.ChangeLog{Log: changelog}
}

func (cl *ChangeLog) FromProto(pb proto.Message) {
	for key, val := range pb.(*scstatepb.ChangeLog).Log {
		cl.Log[key] = val.Log
	}
}

func DeserializeChangeLog(d []byte) *ChangeLog {
	scStateProto := &scstatepb.ChangeLog{}
	err := proto.Unmarshal(d, scStateProto)
	if err != nil {
		logger.WithError(err).Panic("ScState: failed to deserialize chaneglog.")
	}
	cl := NewChangeLog()
	cl.FromProto(scStateProto)
	return cl
}

func (cl *ChangeLog) SerializeChangeLog() []byte {
	rawBytes, err := proto.Marshal(cl.ToProto())
	if err != nil {
		logger.WithError(err).Panic("ScState: failed to serialize changelog.")
	}
	return rawBytes
}

func (scState *ScState) DeepCopy() *ScState {
	newScState := &ScState{make(map[string]map[string]string), make([]*Event, 0), &sync.RWMutex{}}

	for address, addressState := range scState.states {
		newAddressState := make(map[string]string)
		for key, value := range addressState {
			newAddressState[key] = value
		}

		newScState.states[address] = addressState
	}

	for _, event := range scState.events {
		newScState.events = append(newScState.events, event)
	}

	return newScState
}
