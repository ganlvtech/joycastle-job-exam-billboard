package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
)

// region Billboard data structure

type BillboardItem struct {
	UID       int64 // 玩家 UID
	Score     int64 // 活动总分 0 ~ 10000
	Timestamp int64 // 得到分数的秒级时间戳
}

type Billboard struct {
	RankMap map[int64]int // uid -> index
	Data    []BillboardItem
}

func NewBillboard(data []BillboardItem) *Billboard {
	sort.Slice(data, func(i, j int) bool {
		if data[i].Score > data[j].Score { // 分数高的靠前
			return true
		}
		if data[i].Score < data[j].Score {
			return false
		}
		if data[i].Timestamp < data[j].Timestamp { // 时间早的靠前
			return true
		}
		return false
	})
	rankMap := make(map[int64]int, len(data))
	for i, item := range data {
		rankMap[item.UID] = i
	}
	return &Billboard{
		RankMap: rankMap,
		Data:    data,
	}
}

// endregion

// region Request and Response

type NearbyRanksResponse struct {
	Code int64
	Msg  string
	Data []NearbyRanksResponseListItem
}

type NearbyRanksResponseListItem struct {
	UID       int64
	Score     int64
	Timestamp int64
	Rank      int64
}

// endregion

func LoadCSV(filename string) ([]BillboardItem, error) {
	var result []BillboardItem
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("open %s error: %v", filename, err)
	}
	defer f.Close()
	reader := bufio.NewReader(f)
	_, _, _ = reader.ReadLine()
	for {
		var uid int64
		var score int64
		var timestamp int64
		_, err = fmt.Fscanf(reader, "%d,%d,%d\r\n", &uid, &score, &timestamp)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, fmt.Errorf("read UID,Score,Timestamp from %s error: %v", filename, err)
			}
		}
		result = append(result, BillboardItem{
			UID:       uid,
			Score:     score,
			Timestamp: timestamp,
		})
	}
	return result, nil
}

func main() {
	data, err := LoadCSV("data.csv")
	if err != nil {
		log.Fatalf("load csv error: %v", err)
		return
	}
	billboard := NewBillboard(data)

	h := http.NewServeMux()
	h.HandleFunc("/nearby_ranks", func(w http.ResponseWriter, r *http.Request) {
		u, err := url.Parse(r.RequestURI)
		if err != nil {
			w.WriteHeader(400)
			_, _ = fmt.Fprintf(w, "parse uri error: %v", err)
			return
		}
		query := u.Query()
		uidStr := query.Get("uid")
		uid, err := strconv.ParseInt(uidStr, 10, 64)
		if err != nil {
			w.WriteHeader(400)
			_, _ = fmt.Fprintf(w, "parse uid error: %v", err)
			return
		}

		index, ok := billboard.RankMap[uid]
		if !ok {
			w.WriteHeader(400)
			_, _ = fmt.Fprintf(w, "uid not found: %d", uid)
			return
		}
		beginIndex := index - 10
		endIndex := index + 10
		if beginIndex < 0 {
			beginIndex = 0
		}
		if endIndex >= len(billboard.Data) {
			endIndex = len(billboard.Data) - 1
		}
		respData := []NearbyRanksResponseListItem{}
		for i := beginIndex; i <= endIndex; i++ {
			item := billboard.Data[i]
			respData = append(respData, NearbyRanksResponseListItem{
				UID:       item.UID,
				Score:     item.Score,
				Timestamp: item.Timestamp,
				Rank:      int64(i + 1),
			})
		}
		w.Header().Add("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(200)
		b, err := json.Marshal(NearbyRanksResponse{
			Code: 0,
			Msg:  "OK",
			Data: respData,
		})
		if err != nil {
			return
		}
		_, _ = w.Write(b)
	})
	fmt.Println("Listening on 127.0.0.1:8000")
	err = http.ListenAndServe("127.0.0.1:8000", h)
	if err != nil {
		log.Fatal(err)
		return
	}
}
