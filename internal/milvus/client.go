package milvus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	baseURL string
	client  *http.Client
}

type ClientConfig struct {
	BaseURL string
	Timeout time.Duration
}

func NewClient(config *ClientConfig) *Client {
	if config == nil {
		config = &ClientConfig{
			BaseURL: "http://localhost:19530",
			Timeout: 10 * time.Second,
		}
	}

	return &Client{
		baseURL: config.BaseURL,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

func (c *Client) Close() error {
	return nil
}

// CreateCollection 創建集合
func (c *Client) CreateCollection(name string, dimension int) error {
	url := fmt.Sprintf("%s/v1/vector/collections/create", c.baseURL)

	payload := map[string]interface{}{
		"collectionName": name,
		"dimension":      dimension,
		"fields": []map[string]interface{}{
			{
				"name":       "id",
				"data_type":  "INT64",
				"is_primary": true,
				"autoID":     true,
			},
			{
				"name":      "vector",
				"data_type": "FLOAT_VECTOR",
				"dim":       dimension,
			},
			{
				"name":       "text",
				"data_type":  "VARCHAR",
				"max_length": 65535,
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化請求失敗: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("創建請求失敗: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer root:Milvus")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("發送請求失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("創建集合失敗: HTTP %d, 回應: %s", resp.StatusCode, string(body))
	}

	return nil
}

// InsertVectors 插入向量數據
func (c *Client) InsertVectors(collectionName string, vectors []map[string]interface{}) error {
	url := fmt.Sprintf("%s/v1/vector/insert", c.baseURL)

	// 準備數據
	payload := map[string]interface{}{
		"collectionName": collectionName,
		"data":           vectors,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化請求失敗: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("創建請求失敗: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer root:Milvus")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("發送請求失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("插入向量失敗: HTTP %d, 回應: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ListVectors 列出向量數據
func (c *Client) ListVectors(collectionName string) ([]map[string]interface{}, error) {
	// 使用 query 方式查詢所有數據
	queryURL := fmt.Sprintf("%s/v1/vector/query", c.baseURL)

	queryPayload := map[string]interface{}{
		"collectionName": collectionName,
		"output_fields":  []string{"id", "vector", "text"},
		"filter":         "id > 0",
		"limit":          100,
	}

	queryJSON, err := json.Marshal(queryPayload)
	if err != nil {
		return nil, fmt.Errorf("序列化查詢請求失敗: %v", err)
	}

	fmt.Printf("查詢請求數據: %s\n", string(queryJSON))

	queryReq, err := http.NewRequest("POST", queryURL, bytes.NewBuffer(queryJSON))
	if err != nil {
		return nil, fmt.Errorf("創建查詢請求失敗: %v", err)
	}

	queryReq.Header.Set("Content-Type", "application/json")
	queryReq.Header.Set("Accept", "application/json")
	queryReq.Header.Set("Authorization", "Bearer root:Milvus")

	queryResp, err := c.client.Do(queryReq)
	if err != nil {
		return nil, fmt.Errorf("發送查詢請求失敗: %v", err)
	}
	defer queryResp.Body.Close()

	queryBody, err := io.ReadAll(queryResp.Body)
	if err != nil {
		return nil, fmt.Errorf("讀取查詢回應失敗: %v", err)
	}

	fmt.Printf("查詢回應: %s\n", string(queryBody))

	if queryResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("查詢失敗: HTTP %d, 回應: %s", queryResp.StatusCode, string(queryBody))
	}

	var result struct {
		Code int                      `json:"code"`
		Data []map[string]interface{} `json:"data"`
	}

	if err := json.Unmarshal(queryBody, &result); err != nil {
		return nil, fmt.Errorf("解析查詢回應失敗: %v", err)
	}

	return result.Data, nil
}

// DeleteVectors 刪除向量
func (c *Client) DeleteVectors(collectionName string, ids []string) error {
	url := fmt.Sprintf("%s/v1/vector/delete", c.baseURL)

	// 準備請求體
	payload := map[string]interface{}{
		"dbName":         "default",
		"collectionName": collectionName,
		"id":             ids,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化請求失敗: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("創建請求失敗: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer root:Milvus")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("發送請求失敗: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("讀取回應失敗: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("刪除向量失敗: HTTP %d, 回應: %s", resp.StatusCode, string(body))
	}

	// 檢查回應碼是否為200
	var result struct {
		Code int         `json:"code"`
		Data interface{} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("解析回應失敗: %v", err)
	}

	if result.Code != 200 {
		return fmt.Errorf("刪除向量失敗: 回應碼 %d", result.Code)
	}

	return nil
}

// ListCollections 列出所有集合
func (c *Client) ListCollections() ([]string, error) {
	url := fmt.Sprintf("%s/v1/vector/collections", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("創建請求失敗: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer root:Milvus")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("發送請求失敗: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("讀取回應失敗: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("列出集合失敗: HTTP %d, 回應: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Code int      `json:"code"`
		Data []string `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析回應失敗: %v", err)
	}

	if result.Code != 200 {
		return nil, fmt.Errorf("列出集合失敗: 回應碼 %d", result.Code)
	}

	return result.Data, nil
}

// DeleteCollection 刪除集合
func (c *Client) DeleteCollection(collectionName string) error {
	url := fmt.Sprintf("%s/v1/vector/collections/drop", c.baseURL)

	payload := map[string]interface{}{
		"collectionName": collectionName,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化請求失敗: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("創建請求失敗: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer root:Milvus")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("發送請求失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("刪除集合失敗: HTTP %d, 回應: %s", resp.StatusCode, string(body))
	}

	return nil
}

// SearchVectors 向量搜尋
func (c *Client) SearchVectors(collectionName string, vectorToSearch []float32, topK int) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/v1/vector/search", c.baseURL)

	// 準備請求體
	payload := map[string]interface{}{
		"collectionName": collectionName,
		"vector":         vectorToSearch,
		"output_fields":  []string{"id", "text"},
		"topk":           topK,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("序列化請求失敗: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("創建請求失敗: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer root:Milvus")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("發送請求失敗: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("讀取回應失敗: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("向量搜尋失敗: HTTP %d, 回應: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Code int                      `json:"code"`
		Data []map[string]interface{} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析回應失敗: %v", err)
	}

	if result.Code != 200 {
		return nil, fmt.Errorf("向量搜尋失敗: 回應碼 %d", result.Code)
	}

	return result.Data, nil
}

func (c *Client) Search(collectionName string, vector []float32, topK int) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/v1/vector/search", c.baseURL)

	payload := map[string]interface{}{
		"collectionName": collectionName,
		"vector":         vector,
		"outputFields":   []string{"id", "text"},
		"limit":          topK,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("序列化請求失敗: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("創建請求失敗: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer root:Milvus")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("發送請求失敗: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("讀取回應失敗: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("搜索失敗: HTTP %d, 回應: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Code int                      `json:"code"`
		Data []map[string]interface{} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析回應失敗: %v", err)
	}

	if result.Code != 200 {
		return nil, fmt.Errorf("搜索失敗: 回應碼 %d", result.Code)
	}

	return result.Data, nil
}

// CollectionExists 檢查指定的集合是否存在
func (c *Client) CollectionExists(collectionName string) (bool, error) {
	collections, err := c.ListCollections()
	if err != nil {
		return false, fmt.Errorf("獲取集合列表失敗: %v", err)
	}

	for _, name := range collections {
		if name == collectionName {
			return true, nil
		}
	}

	return false, nil
}
