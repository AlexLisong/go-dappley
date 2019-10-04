package lScState

import (
	"testing"

	"github.com/dappley/go-dappley/core/scState"
	"github.com/stretchr/testify/assert"
)

func TestScState_RevertState(t *testing.T) {

	ss := scState.NewScState()
	ls := make(map[string]string)
	ls["key1"] = "value1"
	state := ss.GetStates()
	state["addr1"] = ls

	changeLog1 := make(map[string]map[string]string)
	changeLog2 := make(map[string]map[string]string)
	changeLog3 := make(map[string]map[string]string)

	changePair1 := make(map[string]string)
	changePair2 := make(map[string]string)
	changePair3 := make(map[string]string)
	changePair4 := make(map[string]string)

	expect1 := make(map[string]map[string]string)
	expect2 := make(map[string]map[string]string)
	expect3 := make(map[string]map[string]string)

	changePair1["key1"] = "2"
	changePair2["key3"] = "3"
	changePair3["key4"] = "4"
	changePair4["key1"] = "2"
	changePair4["key4"] = "4"

	expect1["addr1"] = changePair1

	expect2["addr1"] = changePair4
	expect2["addr2"] = changePair2

	changeLog1["addr1"] = changePair1
	revertState(changeLog1, ss)
	assert.Equal(t, expect1, state)

	changeLog2["addr2"] = changePair2
	changeLog2["addr1"] = changePair3
	revertState(changeLog2, ss)
	assert.Equal(t, expect2, state)

	changeLog3["addr2"] = nil
	changeLog3["addr1"] = nil
	revertState(changeLog3, ss)
	assert.Equal(t, expect3, state)
	assert.Equal(t, 0, len(state))

}

func TestScState_FindChangedValue(t *testing.T) {
	newSS := scState.NewScState()
	oldSS := scState.NewScState()

	ls1 := make(map[string]string)
	ls2 := make(map[string]string)
	ls3 := make(map[string]string)

	ls1["key1"] = "value1"
	ls1["key2"] = "value2"
	ls1["key3"] = "value3"

	ls2["key1"] = "value1"
	ls2["key2"] = "value2"
	ls2["key3"] = "4"

	ls3["key1"] = "value1"
	ls3["key3"] = "4"

	expect1 := make(map[string]map[string]string)
	expect2 := make(map[string]map[string]string)
	expect3 := make(map[string]map[string]string)
	expect4 := make(map[string]map[string]string)
	expect5 := make(map[string]map[string]string)
	expect6 := make(map[string]map[string]string)

	expect2["address1"] = nil
	expect4["address1"] = map[string]string{
		"key2": "value2",
		"key3": "value3",
	}

	expect5["address1"] = map[string]string{
		"key2": "value2",
		"key3": "value3",
	}

	expect5["address2"] = nil

	expect6["address2"] = ls2

	change1 := findChangedValue(newSS, oldSS)
	assert.Equal(t, expect1, change1)

	newS := newSS.GetStates()
	olds := oldSS.GetStates()
	newS["address1"] = ls1
	change2 := findChangedValue(newSS, oldSS)
	assert.Equal(t, 1, len(change2))
	assert.Equal(t, expect2, change2)

	olds["address1"] = ls1
	change3 := findChangedValue(newSS, oldSS)
	assert.Equal(t, 0, len(change3))
	assert.Equal(t, expect3, change3)

	newS["address1"] = ls3
	change4 := findChangedValue(newSS, oldSS)
	assert.Equal(t, 1, len(change4))
	assert.Equal(t, expect4, change4)

	newS["address2"] = ls2
	change5 := findChangedValue(newSS, oldSS)
	assert.Equal(t, 2, len(change5))
	assert.Equal(t, expect5, change5)

	olds["address2"] = ls2
	olds["address1"] = ls3
	delete(newS, "address2")
	change6 := findChangedValue(newSS, oldSS)
	assert.Equal(t, 1, len(change6))
	assert.Equal(t, expect6, change6)

}
