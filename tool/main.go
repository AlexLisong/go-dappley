package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/dappley/go-dappley/common"
	"github.com/dappley/go-dappley/config"
	"github.com/dappley/go-dappley/config/pb"
	"github.com/dappley/go-dappley/consensus"
	"github.com/dappley/go-dappley/core"
	"github.com/dappley/go-dappley/storage"
)

const (
	genesisAddr           = "121yKAXeG4cw6uaGCBYjWk9yTWmMkhcoDD"
	genesisFilePath       = "conf/genesis.conf"
	defaultPassword       = "password"
	defaultTimeBetweenBlk = 3
)

type fileInfo struct {
	height        int
	differentFrom int
	db            *storage.LevelDB
}

type Key struct {
	Key     string `json:"key"`
	Address string `json:"address"`
}

type Keys struct {
	Keys []Key `json:"keys"`
}

func main() {
	// var filePath string

	numberBuffer := flag.Int("number", 1, "an int")

	flag.Parse()
	genesisConf := &configpb.DynastyConfig{}
	config.LoadConfig(genesisFilePath, genesisConf)
	fmt.Println("load genesis file")
	number := *numberBuffer
	files := make([]fileInfo, number)
	maxProducers := (int)(genesisConf.GetMaxProducers())
	dynasty := consensus.NewDynastyWithConfigProducers(genesisConf.GetProducers(), maxProducers)
	keys := loadPrivateKey()
	for i := 0; i < number; i++ {
		reader := bufio.NewReader(os.Stdin)
		//enter filename
		fmt.Printf("Enter file name for blockchain%d: \n", i+1)
		text, _ := reader.ReadString('\n')
		text = strings.TrimSuffix(text, "\n")
		db := storage.OpenDatabase(text)
		defer db.Close()
		files[i].db = db
		//enter blockchain height
		fmt.Printf("Enter max height for blockchain%d: \n", i+1)
		height, _ := reader.ReadString('\n')
		height = strings.TrimSuffix(height, "\n")
		iheight, _ := strconv.Atoi(height)
		files[i].height = iheight
		//enter height of blockchain have different with other block (0 means no different)
		fmt.Printf("Enter a different starting height for blockchain%d(0 for no different): \n", i+1)
		different, _ := reader.ReadString('\n')
		different = strings.TrimSuffix(different, "\n")
		idifferent, _ := strconv.Atoi(different)
		if iheight <= idifferent || idifferent < 1 {
			files[i].differentFrom = iheight
		} else {
			files[i].differentFrom = idifferent
		}

	}

	generateNewBlockChain(files, dynasty, keys)
}

func generateNewBlockChain(files []fileInfo, d *consensus.Dynasty, keys Keys) {
	bcs := make([]*core.Blockchain, len(files))
	addr := core.NewAddress(genesisAddr)
	for i := 0; i < len(files); i++ {
		bc := core.CreateBlockchain(addr, files[i].db, nil, 20)
		bcs[i] = bc
	}
	var time int64
	time = 1532392928
	max, index := getMaxHeightOfDifferentStart(files)
	for i := 0; i < max; i++ {
		time = time + defaultTimeBetweenBlk
		b := generateBlock(bcs[index], time, d, keys)

		for idx := 0; idx < len(files); idx++ {
			if files[idx].differentFrom >= i {
				bcs[idx].AddBlockToTail(b)
			}
		}
	}

	for i := 0; i < len(files); i++ {
		makeBlockChainToSize(bcs[i], files[i].height, time, d, keys)
		fmt.Println(bcs[i].GetMaxHeight())
	}

}

func getMaxHeightOfDifferentStart(files []fileInfo) (int, int) {
	max := 0
	index := 0
	for i := 0; i < len(files); i++ {
		if max < files[i].differentFrom {
			max = files[i].differentFrom
			index = i
		}
	}
	return max, index
}

func makeBlockChainToSize(bc *core.Blockchain, size int, time int64, d *consensus.Dynasty, keys Keys) {

	for bc.GetMaxHeight() < uint64(size) {
		time = time + defaultTimeBetweenBlk
		b := generateBlock(bc, time, d, keys)
		bc.AddBlockToTail(b)
	}
}

func generateBlock(bc *core.Blockchain, time int64, d *consensus.Dynasty, keys Keys) *core.Block {
	producer := d.ProducerAtATime(time)
	key := keys.getPrivateKeyByAddress(producer)
	tailBlk, _ := bc.GetTailBlock()
	cbtx := core.NewCoinbaseTX(producer, "", bc.GetMaxHeight()+1, common.NewAmount(0))
	b := core.NewBlockWithTimestamp([]*core.Transaction{&cbtx}, tailBlk, time)
	hash := b.CalculateHashWithNonce(0)
	b.SetHash(hash)
	b.SetNonce(0)
	b.SignBlock(key, hash)
	return b
}

func loadPrivateKey() Keys {
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

func (k Keys) getPrivateKeyByAddress(address string) string {
	for _, key := range k.Keys {
		if key.Address == address {
			return key.Key
		}
	}
	return ""
}
