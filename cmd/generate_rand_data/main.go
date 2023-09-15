package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
)

func main() {
	f, err := os.OpenFile("data.csv", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer f.Close()
	writer := bufio.NewWriter(f)
	defer writer.Flush()
	_, _ = writer.WriteString("UID,Score,Timestamp\r\n")
	for uid := 1; uid <= 1e6; uid++ {
		score := rand.Intn(10001)
		timestamp := time.Date(2023, 9, 15, 0, 0, 0, 0, time.Local).Unix() + rand.Int63n(30*86400)
		_, _ = fmt.Fprintf(writer, "%d,%d,%d\r\n", uid, score, timestamp)
	}
}
