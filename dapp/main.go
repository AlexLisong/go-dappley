// Copyright (C) 2018 go-dappley authors
//
// This file is part of the go-dappley library.
//
// the go-dappley library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-dappley library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-dappley library.  If not, see <http://www.gnu.org/licenses/>.
//

package main

import (
	"flag"
	"github.com/dappley/go-dappley/common/log"
	"github.com/dappley/go-dappley/config"
	"github.com/dappley/go-dappley/config/pb"
	"github.com/dappley/go-dappley/consensus"
	"github.com/dappley/go-dappley/core/account"
	"github.com/dappley/go-dappley/core/block"
	"github.com/dappley/go-dappley/core/blockchain"
	"github.com/dappley/go-dappley/core/blockproducerinfo"
	"github.com/dappley/go-dappley/core/scState"
	"github.com/dappley/go-dappley/logic"
	"github.com/dappley/go-dappley/logic/blockproducer"
	"github.com/dappley/go-dappley/logic/downloadmanager"
	"github.com/dappley/go-dappley/logic/lblockchain"
	"github.com/dappley/go-dappley/logic/transactionpool"
	"github.com/dappley/go-dappley/metrics/logMetrics"
	"github.com/dappley/go-dappley/network"
	"github.com/dappley/go-dappley/rpc"
	"github.com/dappley/go-dappley/storage"
	"github.com/dappley/go-dappley/vm"
	"github.com/fsnotify/fsnotify"
	logger "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
	_ "net/http/pprof"
	"time"
)

const (
	genesisAddr     = "121yKAXeG4cw6uaGCBYjWk9yTWmMkhcoDD"
	configFilePath  = "conf/default.conf"
	genesisFilePath = "conf/genesis.conf"
	defaultPassword = "password"
	size1kB         = 1024
)

func main() {
	viper.AddConfigPath(".")
	viper.SetConfigFile("conf/dappley.yaml")
	if err := viper.ReadInConfig(); err != nil {
		logger.Errorf("Cannot load dappley configurations from file!  error： %v", err.Error())
		return
	}

	log.BuildLogAndInit()
	var filePath string
	flag.StringVar(&filePath, "f", configFilePath, "Configuration File Path. Default to conf/default.conf")

	var genesisPath string
	flag.StringVar(&genesisPath, "g", genesisFilePath, "Genesis Configuration File Path. Default to conf/genesis.conf")
	flag.Parse()

	logger.Infof("Genesis conf file is %v,node conf file is %v", genesisPath, filePath)
	//load genesis file information
	genesisConf := &configpb.DynastyConfig{}
	config.LoadConfig(genesisPath, genesisConf)

	if genesisConf == nil {
		logger.Error("Cannot load genesis configurations from file! Exiting...")
		return
	}

	//load config file information
	conf := &configpb.Config{}
	config.LoadConfig(filePath, conf)
	if conf == nil {
		logger.Error("Cannot load configurations from file! Exiting...")
		return
	}

	//setup
	db := storage.OpenDatabase(conf.GetNodeConfig().GetDbPath())
	defer db.Close()
	node, err := initNode(conf, db)
	if err != nil {
		return
	} else {
		defer node.Stop()
	}

	//create blockchain
	conss, _ := initConsensus(genesisConf, conf)
	txPoolLimit := conf.GetNodeConfig().GetTxPoolLimit() * size1kB
	nodeAddr := conf.GetNodeConfig().GetNodeAddress()
	blkSizeLimit := conf.GetNodeConfig().GetBlkSizeLimit() * size1kB
	scManager := vm.NewV8EngineManager(account.NewAddress(nodeAddr))
	txPool := transactionpool.NewTransactionPool(node, txPoolLimit)
	//utxo.NewPool()
	bc, err := lblockchain.GetBlockchain(db, conss, txPool, scManager, int(blkSizeLimit))

	var LIBBlk *block.Block = nil
	if err != nil {
		bc, err = logic.CreateBlockchain(account.NewAddress(genesisAddr), db, conss, txPool, scManager, int(blkSizeLimit))
		if err != nil {
			logger.Panic(err)
		}
	}else {
		LIBBlk, _ = bc.GetLIB()
	}
	bc.SetState(blockchain.BlockchainInit)

	bm := lblockchain.NewBlockchainManager(bc, blockchain.NewBlockPool(LIBBlk), node, conss)
	if err != nil {
		logger.WithError(err).Error("Failed to initialize the node! Exiting...")
		return
	}

	downloadManager := downloadmanager.NewDownloadManager(node, bm, len(conss.GetProducers()))
	downloadManager.Start()
	bm.SetDownloadRequestCh(downloadManager.GetDownloadRequestCh())

	bm.Getblockchain().SetState(blockchain.BlockchainReady)

	//start mining
	logic.SetLockAccount() //lock the account
	logic.SetMinerKeyPair(conf.GetConsensusConfig().GetPrivateKey())

	//start rpc server
	nodeConf := conf.GetNodeConfig()
	server := rpc.NewGrpcServerWithMetrics(node, bm, defaultPassword, conss, &rpc.MetricsServiceConfig{
		PollingInterval: nodeConf.GetMetricsPollingInterval(), TimeSeriesInterval: nodeConf.GetMetricsInterval()})

	server.Start(conf.GetNodeConfig().GetRpcPort())
	defer server.Stop()

	producer := blockproducerinfo.NewBlockProducerInfo(conf.GetConsensusConfig().GetMinerAddress())
	blockProducer := blockproducer.NewBlockProducer(bm, conss, producer)
	blockProducer.Start()
	defer blockProducer.Stop()

	bm.RequestDownloadBlockchain()

	if viper.GetBool("scheduleevents.enable") {
		lblockchain.SetEnableRunScheduleEvents()
	}
	if viper.GetBool("metrics.open") {
		logMetrics.LogMetricsInfo(bm.Getblockchain())
	}
	if viper.GetBool("pprof.open") {
		go func() {
			http.ListenAndServe(":60001", nil)
		}()
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		if viper.GetBool("producerchange") {
			logger.Infof("producer change true")
			viper.SetConfigType("yaml")
			viper.Set("producerchange",false)
			err := viper.WriteConfig()
			if err != nil{
				logger.Errorf("WriteConfig err: %v", err.Error())
			}
			conf := &configpb.Config{}
			config.LoadConfig(filePath, conf)
			if conf == nil {
				logger.Error("Cannot load configurations from file! Exiting...")
				return
			}

			genesisConf := &configpb.DynastyConfig{}
			config.LoadConfig(genesisPath, genesisConf)

			blockProducer.Stop()

			logic.SetMinerKeyPair(conf.GetConsensusConfig().GetPrivateKey())

			producer := blockproducerinfo.NewBlockProducerInfo(conf.GetConsensusConfig().GetMinerAddress())
			setConsensus(conss,genesisConf,conf,producer)
			blockProducer = blockproducer.NewBlockProducer(bm, conss, producer)
			blockProducer.Start()
		}
	})

	ProducerForBlock(bc,conss)

	select {}
}


func initConsensus(conf *configpb.DynastyConfig, generalConf *configpb.Config) (*consensus.DPOS, *consensus.Dynasty) {
	//set up consensus
	conss := consensus.NewDPOS(blockproducerinfo.NewBlockProducerInfo(generalConf.GetConsensusConfig().GetMinerAddress()))
	dynasty := consensus.NewDynastyWithConfigProducers(conf.GetProducers(), (int)(conf.GetMaxProducers()))
	conss.SetDynasty(dynasty)
	conss.SetKey(generalConf.GetConsensusConfig().GetPrivateKey())
	logger.WithFields(logger.Fields{
		"miner_address": generalConf.GetConsensusConfig().GetMinerAddress(),
	}).Info("Consensus is configured.")
	return conss, dynasty
}

func setConsensus(conss *consensus.DPOS,conf *configpb.DynastyConfig, generalConf *configpb.Config,producer *blockproducerinfo.BlockProducerInfo){
	dynasty := consensus.NewDynastyWithConfigProducers(conf.GetProducers(), (int)(conf.GetMaxProducers()))
	conss.SetDynasty(dynasty)
	conss.SetKey(generalConf.GetConsensusConfig().GetPrivateKey())
	conss.SetProducer(producer)

	logger.WithFields(logger.Fields{
		"miner_address": generalConf.GetConsensusConfig().GetMinerAddress(),
	}).Info("Consensus is set.")
}

func initNode(conf *configpb.Config, db storage.Storage) (*network.Node, error) {

	nodeConfig := conf.GetNodeConfig()
	seeds := nodeConfig.GetSeed()
	port := nodeConfig.GetPort()
	keyPath := nodeConfig.GetKeyPath()

	node := network.NewNode(db, seeds)
	err := node.Start(int(port), keyPath)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	return node, nil
}

func ProducerForBlock(bc *lblockchain.Blockchain, conss *consensus.DPOS){
	go func() {
		tick := time.NewTicker(time.Second*time.Duration(60))
		defer tick.Stop()

		for{
			select {
			case <-tick.C:
				sc := scState.LoadScStateFromDatabase(bc.GetDb())
				producerMap := sc.GetStorageByAddress(vm.PRODUCER_ADDR)
				logger.Infof("producer map: %v", producerMap)
				if producerMap == nil || len(producerMap) == 1{
					logger.Infof("producer map is empty")
					continue
				}

				addProducer := producerMap[vm.PRODUCER_ADD_INDEX]
				if addProducer != ""{
					err := conss.GetDynasty().AddProducer(addProducer)
					if err != nil{
						logger.Infof("add producer err: %v", err.Error())
					}
				}

				delproducer := producerMap[vm.PRODUCER_DEL_INDEX]
				if delproducer != ""{
					err := conss.GetDynasty().DeleteProducer(delproducer)
					if err != nil{
						logger.Infof("delete producer err: %v", err.Error())
					}
				}

				sc.DeleteAddress(vm.PRODUCER_ADDR)
				sc.SaveToDatabase(bc.GetDb())
			}
		}
	}()
}