package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	metricspb "github.com/dappley/go-dappley/metrics/pb"
)

var sampleStatValueFunc = func() metricspb.StatValue { return &metricspb.Stat_TransactionPoolSize{} }

func TestDataStore_String(t *testing.T) {
	t.Parallel()
	ds := NewDataStore(1, time.Second)
	assert.Equal(t, "{\"metrics\":{}}", ds.String())

	// register new metric
	err := ds.RegisterNewMetric("test", sampleStatValueFunc)
	assert.Nil(t, err)
	assert.Equal(t, "{\"metrics\":{\"test\":{\"stats\":[]}}}", ds.String())
}

func TestDataStoreCapacityConstraint(t *testing.T) {
	t.Parallel()
	ds := NewDataStore(1, time.Second)
	err := ds.RegisterNewMetric("test", sampleStatValueFunc)
	assert.Nil(t, err)

	ds.StartUpdate()
	time.Sleep(3 * time.Second)
	ds.StopUpdate()
	time.Sleep(time.Second)
	// ensure capacity constraint is not violated
	assert.Equal(t, 1, ds.getNumStats("test"))
}

func TestDataStore_RegisterNewMetric(t *testing.T) {
	t.Parallel()
	ds := NewDataStore(1, time.Second)

	err := ds.RegisterNewMetric("test", sampleStatValueFunc)
	assert.Nil(t, err)

	err = ds.RegisterNewMetric("test", sampleStatValueFunc)
	assert.NotNil(t, err)
	assert.Equal(t, "unable to register duplicate metric", err.Error())
}

func TestDataStore_Update(t *testing.T) {
	t.Parallel()
	ds := NewDataStore(5, time.Second)

	err := ds.RegisterNewMetric("test", sampleStatValueFunc)
	assert.Nil(t, err)

	ds.StartUpdate()
	time.Sleep(2 * time.Second)
	ds.StopUpdate()

	// in case we need to wait for last collected stat
	time.Sleep(time.Second)

	// ensure some stats were collected
	numStats := ds.getNumStats("test")
	assert.True(t, numStats > 0)

	// test stop update
	time.Sleep(2 * time.Second)
	assert.Equal(t, numStats, ds.getNumStats("test"))

	// test restart
	ds.StartUpdate()
	time.Sleep(2 * time.Second)
	assert.True(t, ds.getNumStats("test") > numStats)
	ds.StopUpdate()
	time.Sleep(time.Second)
	numStats = ds.getNumStats("test")
	time.Sleep(2 * time.Second)
	assert.Equal(t, numStats, ds.getNumStats("test"))
}
