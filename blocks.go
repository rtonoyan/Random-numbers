package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"sync"
	"time"
)

type Blockinfo struct {
	Requester      string      `json:"requester"`
	Min            float64     `json:"min"`
	Max            float64     `json:"max"`
	Count          int         `json:"count"`
	Unique         bool        `json:"unique"`
	GenerationTime string      `json:"generation_time"`
	RandomNumbers  []int       `json:"random_numbers,omitempty"`
	RandomFloats   []float64   `json:"random_floats,omitempty"`
	Stats          map[int]int `json:"stats,omitempty"`
	RequestDetails string      `json:"request_details"`
}

type Block struct {
	ID           int         `json:"id"`
	Timestamp    string      `json:"timestamp"`
	Blockinf     []Blockinfo `json:"block_info"`
	Hash         string      `json:"hash"`
	PreviousHash string      `json:"previous_hash"`
}

var (
	Blockchain       []Block
	futureBlockInfos [10][]Blockinfo
	mutex            sync.Mutex
)

func calculateHash(block Block) string {
	blockInfoJson, err := json.Marshal(block.Blockinf)
	if err != nil {
		fmt.Println("Error serializing Blockinf:", err)
		return ""
	}
	record := strconv.Itoa(block.ID) + block.Timestamp + block.PreviousHash + string(blockInfoJson)
	h := sha256.New()
	h.Write([]byte(record))
	return hex.EncodeToString(h.Sum(nil))
}

func createBlock(oldBlock Block, blockinfos []Blockinfo) Block {
	newBlock := Block{
		ID:           oldBlock.ID + 1,
		Timestamp:    time.Now().String(),
		PreviousHash: oldBlock.Hash,
		Blockinf:     blockinfos,
	}
	newBlock.Hash = calculateHash(newBlock)
	return newBlock
}

func saveBlockToFile(block Block) error {
	file, err := json.MarshalIndent(block, "", " ")
	if err != nil {
		return err
	}
	filename := "block_" + strconv.Itoa(block.ID) + ".json"
	return ioutil.WriteFile(filename, file, 0644)
}

func loadBlockchain() error {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		return err
	}
	for _, file := range files {
		if len(file.Name()) > 6 && file.Name()[:6] == "block_" && file.Name()[len(file.Name())-5:] == ".json" {
			data, err := os.ReadFile(file.Name())
			if err != nil {
				return err
			}
			var block Block
			if err := json.Unmarshal(data, &block); err != nil {
				return err
			}
			Blockchain = append(Blockchain, block)
		}
	}
	if len(Blockchain) == 0 {
		genesisBlock := Block{
			ID:           0,
			Timestamp:    time.Now().String(),
			Hash:         "0",
			PreviousHash: "",
		}
		Blockchain = append(Blockchain, genesisBlock)
	}
	return nil
}
