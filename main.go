package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// 定义与 JSON 数据匹配的结构体
type Audit struct {
	FreezeAuthorityDisabled bool    `json:"freezeAuthorityDisabled"`
	LpBurnedPercentage      float64 `json:"lpBurnedPercentage"`
	MintAuthorityDisabled   bool    `json:"mintAuthorityDisabled"`
	TopHoldersPercentage    float64 `json:"topHoldersPercentage"`
}

type BaseAsset struct {
	CircSupply   float64 `json:"circSupply"`
	Decimals     int     `json:"decimals"`
	Dev          string  `json:"dev"`
	DevMintCount int     `json:"devMintCount"`
	Fdv          float64 `json:"fdv"`
	HolderCount  int     `json:"holderCount"`
	Icon         string  `json:"icon"`
	ID           string  `json:"id"`
	Launchpad    string  `json:"launchpad"`
	Mcap         float64 `json:"mcap"`
	Name         string  `json:"name"`
	NativePrice  float64 `json:"nativePrice"`
	PoolAmount   float64 `json:"poolAmount"`
	Symbol       string  `json:"symbol"`
	TokenProgram string  `json:"tokenProgram"`
	TotalSupply  float64 `json:"totalSupply"`
	UsdPrice     float64 `json:"usdPrice"`
}

type QuoteAsset struct {
	Decimals   int     `json:"decimals"`
	ID         string  `json:"id"`
	PoolAmount float64 `json:"poolAmount"`
	Symbol     string  `json:"symbol"`
}

type Stats struct {
	BuyVolume   float64 `json:"buyVolume"`
	NumBuyers   int     `json:"numBuyers"`
	NumBuys     int     `json:"numBuys"`
	NumSellers  int     `json:"numSellers"`
	NumSells    int     `json:"numSells"`
	NumTraders  int     `json:"numTraders"`
	PriceChange float64 `json:"priceChange"`
	SellVolume  float64 `json:"sellVolume"`
}

type Data struct {
	Audit      Audit      `json:"audit"`
	BaseAsset  BaseAsset  `json:"baseAsset"`
	Chain      string     `json:"chain"`
	CreatedAt  string     `json:"createdAt"`
	Dex        string     `json:"dex"`
	ID         string     `json:"id"`
	Liquidity  float64    `json:"liquidity"`
	QuoteAsset QuoteAsset `json:"quoteAsset"`
	Stats1h    Stats      `json:"stats1h"`
	Stats24h   Stats      `json:"stats24h"`
	Stats5m    Stats      `json:"stats5m"`
	Stats6h    Stats      `json:"stats6h"`
	Type       string     `json:"type"`
	UpdatedAt  string     `json:"updatedAt"`
}

func main() {
	// 定时任务：每 10 分钟执行一次
	ticker := time.NewTicker(3 * time.Minute)
	defer ticker.Stop()

	// 在 goroutine 中执行定时任务
	go func() {
		for {
			select {
			case <-ticker.C:
				runTask()
			}
		}
	}()

	// 阻止主 goroutine 退出
	select {}
}

// runTask 执行你原来的任务逻辑
func runTask() {
	url := "https://api.ape.pro/api/v1/gems"
	requestBody := `{"new":{"notPumpfunToken":false},"aboutToGraduate":{},"graduated":{}}`

	// 发起 HTTP 请求
	body, err := makeRequest(url, requestBody)
	if err != nil {
		log.Printf("Request failed: %v", err)
		return
	}

	// 解析 JSON 响应
	graduatedData, err := extractGraduatedData(body)
	if err != nil {
		log.Printf("Failed to extract graduated data: %v", err)
		return
	}

	// 打开 SQLite 数据库
	db, err := sql.Open("sqlite3", "/home/xonedev/project/tweet_sent.db")
	if err != nil {
		log.Printf("Failed to open database: %v", err)
		return
	}
	defer db.Close()

	// 遍历池数据并插入数据库
	if err := processPools(db, graduatedData["pools"]); err != nil {
		log.Printf("Failed to process pools: %v", err)
	}
}

// makeRequest 发起 HTTP 请求并返回响应体
func makeRequest(url, requestBody string) ([]byte, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	setHeaders(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

// setHeaders 设置 HTTP 请求头
func setHeaders(req *http.Request) {
	req.Header.Set("accept", "*/*")
	req.Header.Set("accept-language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("cookie", "cf_clearance=9aYSJi1nYB46ZLW37KvvVevlkLmDvlzce5XGwPZE9VM-1736822574-1.2.1.1-3h4a52VprdVeakKe3tZQY6sHimRoxD20JYaCtqSYOmsqiq_q542pOnPFe2RHuFB1CZoCYhV7dhnnO8E5jeDAoEHczmPtUtRuoYKaJ31w490UxscOHm.DeFHC0CJDt9s5t58ul4AyfhwRkSa4Lgm0GbkrocwgeVP4xf1kvpWt2_KYb5VvZIrZnr4fMMPSk6eQKKvfpkfOcYuC219XtH87cLRtS9Y5CeEvXPID5Fw0JsQ7doGC51zaayHpO1fbmH50caON0Fo9BqAMhALQQc0yW56LoF.w_wM781RnmtOrqrdSlVB.MtyPzvsmM2V4CqMC1hjLlcrjQoUoKiFiQF0aZw")
	req.Header.Set("dnt", "1")
	req.Header.Set("origin", "https://ape.pro")
	req.Header.Set("priority", "u=1, i")
	req.Header.Set("referer", "https://ape.pro/")
	req.Header.Set("sec-ch-ua", `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`)
	req.Header.Set("sec-ch-ua-mobile", "?1")
	req.Header.Set("sec-ch-ua-platform", `"Android"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-site")
	req.Header.Set("user-agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Mobile Safari/537.36")
	req.Header.Set("x-ape-client-id", "529914d")
	req.Header.Set("x-ape-client-token", "v3FGgD6Ensof6kCno2dz3fDTVqylFayiz1M75v2v1B5+D+b2+pd8TRU3+8W8xV/v91n6wi99+NfHtypfV9nK9thAjYolLDoey8w++rv/HOthiXNTsJ8Zhzc=")
}

// extractGraduatedData 提取"graduated"字段中的数据
func extractGraduatedData(body []byte) (map[string]interface{}, error) {
	var responseData map[string]interface{}
	if err := json.Unmarshal(body, &responseData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	graduatedData, ok := responseData["graduated"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to find 'graduated' key in the response or invalid format")
	}

	return graduatedData, nil
}

// processPools 处理并插入池数据
func processPools(db *sql.DB, poolsData interface{}) error {
	pools, ok := poolsData.([]interface{})
	if !ok {
		return fmt.Errorf("failed to find 'pools' key or invalid format")
	}

	for pool := range pools {
		poolBytes, err := json.Marshal(pool)
		if err != nil {
			return fmt.Errorf("failed to convert pool to JSON bytes: %w", err)
		}

		var data Data
		if err := json.Unmarshal(poolBytes, &data); err != nil {
			return fmt.Errorf("failed to unmarshal pool into Data struct: %w", err)
		}

		// 获取当前时间戳
		currentTime := time.Now()

		// 使用 INSERT OR IGNORE 来避免重复插入
		insertSQL := `INSERT OR IGNORE INTO tokens (id, token_data, timestamp, status) VALUES (?, ?, ?, ?)`
		_, err = db.Exec(insertSQL, data.BaseAsset.ID, poolBytes, currentTime, "unprocessed")
		if err != nil {
			return fmt.Errorf("failed to insert token: %w", err)
		}

	}
	return nil
}
