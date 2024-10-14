package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	MinAllowedValue = -1000000
	MaxAllowedValue = 1000000
	MaxAllowedCount = 10000000
	DefaultMaxValue = 100
)

func handleGenerate(c *gin.Context) {
	min, max, count, flo, unique, itime, err := parseQueryParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if min >= max {
		max = min + DefaultMaxValue
		if max > MaxAllowedValue {
			max = MaxAllowedValue
		}
	}

	if unique {
		rangeSize := int(max - min + 1)
		if count > rangeSize {
			count = rangeSize
		}
	}

	start := time.Now()

	var response gin.H
	var numbers []int
	var floatNumbers []float64
	var stats map[int]int

	if flo {
		floatNumbers, err = generateRandomFloats(min, max, count)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating random numbers"})
			return
		}
		response = gin.H{
			"numbers": floatNumbers,
			"min_num": min,
			"max_num": max,
			"flo":     flo,
			"unique":  unique,
		}
	} else {
		numbers, stats, err = generateRandomNumbersWithStats(int(min), int(max), count, unique)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating random numbers"})
			return
		}
		response = gin.H{
			"numbers": numbers,
			"min_num": min,
			"max_num": max,
			"flo":     flo,
			"unique":  unique,
			"stats":   stats,
		}
	}

	duration := time.Since(start)
	response["generation_time"] = duration.String()
	requestDetails := fmt.Sprintf("min=%v&max=%v&count=%v&flo=%v&unique=%v", min, max, count, flo, unique)
	requester := c.ClientIP()
	blockinfo := Blockinfo{
		Requester:      requester,
		Min:            min,
		Max:            max,
		Count:          count,
		Unique:         unique,
		GenerationTime: duration.String(),
		RequestDetails: requestDetails,
	}

	if flo {
		blockinfo.RandomFloats = floatNumbers
	} else {
		blockinfo.RandomNumbers = numbers
		blockinfo.Stats = stats
	}

	mutex.Lock()
	fmt.Printf("itime: %d\n", itime)
	if itime > 0 {
		if int(itime) > len(futureBlockInfos)-1 {
			itime = uint64(len(futureBlockInfos) - 1)
		}
		futureBlockInfos[itime] = append(futureBlockInfos[itime], blockinfo)
	} else {
		Blockchain[len(Blockchain)-1].Blockinf = append(Blockchain[len(Blockchain)-1].Blockinf, blockinfo)
	}
	mutex.Unlock()

	c.JSON(http.StatusOK, response)
}

func parseQueryParams(c *gin.Context) (float64, float64, int, bool, bool, uint64, error) {
	minStr := c.Query("min")
	maxStr := c.Query("max")
	countStr := c.Query("count")
	floStr := c.Query("flo")
	uniqueStr := c.Query("unique")
	itimeStr := c.Query("itime")

	itime, err := strconv.ParseUint(itimeStr, 10, 64)
	if err != nil {
		itime = 0
	}
	fmt.Println("Parsed itime:", itime)

	min, err := strconv.ParseFloat(minStr, 64)
	if err != nil || min < MinAllowedValue {
		return 0, 0, 0, false, false, 0, fmt.Errorf("invalid or out of range min parameter")
	}

	max, err := strconv.ParseFloat(maxStr, 64)
	if err != nil || max > MaxAllowedValue {
		return 0, 0, 0, false, false, 0, fmt.Errorf("invalid or out of range max parameter")
	}

	count, err := strconv.Atoi(countStr)
	if err != nil || count < 1 || count > MaxAllowedCount {
		return 0, 0, 0, false, false, 0, fmt.Errorf("invalid or out of range count parameter")
	}

	flo, err := strconv.ParseBool(floStr)
	if err != nil {
		flo = false
	}

	unique, err := strconv.ParseBool(uniqueStr)
	if err != nil {
		unique = false
	}

	return min, max, count, flo, unique, itime, nil
}

func publishBlock() {
	mutex.Lock()
	defer mutex.Unlock()
	fmt.Println(futureBlockInfos)
	if len(Blockchain[len(Blockchain)-1].Blockinf) == 0 && len(futureBlockInfos[0]) == 0 {
		for i := 0; i < len(futureBlockInfos)-1; i++ {
			futureBlockInfos[i] = futureBlockInfos[i+1]
		}
		futureBlockInfos[len(futureBlockInfos)-1] = []Blockinfo{}
		return
	}

	lastBlock := Blockchain[len(Blockchain)-1]
	readyBlockInfos := lastBlock.Blockinf

	readyBlockInfos = append(readyBlockInfos, futureBlockInfos[0]...)
	for i := 0; i < len(futureBlockInfos)-1; i++ {
		futureBlockInfos[i] = futureBlockInfos[i+1]
	}
	futureBlockInfos[len(futureBlockInfos)-1] = []Blockinfo{}
	fmt.Println(futureBlockInfos)
	if len(readyBlockInfos) > 0 {
		newBlock := createBlock(lastBlock, readyBlockInfos)
		Blockchain = append(Blockchain, newBlock)
		saveBlockToFile(newBlock)
		Blockchain[len(Blockchain)-1].Blockinf = []Blockinfo{} 
	}
}

func handleGetBlock(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id parameter"})
		return
	}

	filename := "block_" + strconv.Itoa(id) + ".json"
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Block not found"})
		return
	}

	var block Block
	if err := json.Unmarshal(data, &block); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading block data"})
		return
	}

	c.JSON(http.StatusOK, block)
}
