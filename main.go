package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func Blgen() {
	for {
		time.Sleep(1 * time.Minute)
		publishBlock()
		fmt.Println("Block created")
	}
}

func main() {
	r := gin.Default()

	if err := loadBlockchain(); err != nil {
		fmt.Println("Failed to load blockchain:", err)
		return
	}

	go Blgen()
	r.GET("/randomnumber", handleGenerate)
	r.GET("/getblock", handleGetBlock)
	r.Run(":8081")
}
