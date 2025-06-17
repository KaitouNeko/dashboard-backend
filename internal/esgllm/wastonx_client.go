package esgllm

import (
	"ai-workshop/pkg/util"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type WatsonxClient struct {
}

type MessageContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type Message struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type RequestBody struct {
	Messages         []Message   `json:"messages"`
	ProjectID        string      `json:"project_id"`
	ModelID          string      `json:"model_id"`
	FrequencyPenalty int         `json:"frequency_penalty"`
	MaxTokens        int         `json:"max_tokens"`
	PresencePenalty  int         `json:"presence_penalty"`
	Temperature      float64     `json:"temperature"`
	TopP             float64     `json:"top_p"`
	Seed             interface{} `json:"seed"`
	Stop             []string    `json:"stop"`
}

// WatsonxProvider
type WatsonxProvider struct {
	apiKey string
	token  string
}

func NewWatsonxProvider(apiKey string) *WatsonxProvider {

	return &WatsonxProvider{
		apiKey: apiKey,
	}
}

func (p *WatsonxProvider) GetWatsonxToken() error {
	apiKey := util.GetEnvString("WATSONX_API_KEY", "")

	form := url.Values{}
	form.Add("apikey", apiKey)
	form.Add("grant_type", "urn:ibm:params:oauth:grant-type:apikey")

	req, err := http.NewRequest("POST", "https://iam.cloud.ibm.com/identity/token", bytes.NewBufferString(form.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// request with API key
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status error response: %s", string(body))
	}

	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	fmt.Println("Status Code:", resp.StatusCode)
	fmt.Println("Body:", string(body))

	p.token = result.AccessToken

	return nil
}

func (p *WatsonxProvider) GenerateContent(ctx context.Context, prompt string) (string, error) {
	p.GetWatsonxToken()

	url := "https://jp-tok.ml.cloud.ibm.com/ml/v1/text/chat?version=2023-05-29"
	// accessToken := "eyJraWQiOiIyMDE5MDcyNCIsImFsZyI6IlJTMjU2In0.eyJpYW1faWQiOiJJQk1pZC02OTYwMDBONVVHIiwiaWQiOiJJQk1pZC02OTYwMDBONVVHIiwicmVhbG1pZCI6IklCTWlkIiwianRpIjoiM2IzN2M2MjUtOGNjNy00ZGZkLTkxZWEtNzEyODI3N2I0ZDM4IiwiaWRlbnRpZmllciI6IjY5NjAwME41VUciLCJnaXZlbl9uYW1lIjoiY2lvdSIsImZhbWlseV9uYW1lIjoia2lraSIsIm5hbWUiOiJjaW91IGtpa2kiLCJlbWFpbCI6Imtpa2kuY2lvdUBjbG91ZC1pbnRlcmFjdGl2ZS5jb20iLCJzdWIiOiJraWtpLmNpb3VAY2xvdWQtaW50ZXJhY3RpdmUuY29tIiwiYXV0aG4iOnsic3ViIjoia2lraS5jaW91QGNsb3VkLWludGVyYWN0aXZlLmNvbSIsImlhbV9pZCI6IklCTWlkLTY5NjAwME41VUciLCJuYW1lIjoiY2lvdSBraWtpIiwiZ2l2ZW5fbmFtZSI6ImNpb3UiLCJmYW1pbHlfbmFtZSI6Imtpa2kiLCJlbWFpbCI6Imtpa2kuY2lvdUBjbG91ZC1pbnRlcmFjdGl2ZS5jb20ifSwiYWNjb3VudCI6eyJ2YWxpZCI6dHJ1ZSwiYnNzIjoiNTQwOGM1NTM0Zjk3NDFmNDllZWY4MTNjNDdhNDhiNWYiLCJpbXNfdXNlcl9pZCI6IjEzODA2NjgxIiwiZnJvemVuIjp0cnVlLCJpc19lbnRlcnByaXNlX2FjY291bnQiOmZhbHNlLCJlbnRlcnByaXNlX2lkIjoiZWU1NzVjNTc3ODc2NGQ0MDkxNTVhYTM1NzgwZWM4ZDEiLCJpbXMiOiIyODE0NTI3In0sIm1mYSI6eyJpbXMiOnRydWV9LCJpYXQiOjE3NTAwOTYwMTgsImV4cCI6MTc1MDA5OTYxOCwiaXNzIjoiaHR0cHM6Ly9pYW0uY2xvdWQuaWJtLmNvbS9pZGVudGl0eSIsImdyYW50X3R5cGUiOiJ1cm46aWJtOnBhcmFtczpvYXV0aDpncmFudC10eXBlOmFwaWtleSIsInNjb3BlIjoiaWJtIG9wZW5pZCIsImNsaWVudF9pZCI6ImRlZmF1bHQiLCJhY3IiOjEsImFtciI6WyJwd2QiXX0.kGsi-L-A4-7YvVRJ5LNXi6C5il3OJdzD83dNjSWVAJQ_zMgr8Mmqps_wVDaR1ZJ-h4_udkX30h9Yjgg1Pr1A8TdmhJO29kx0kbIO93FjUj6cmrk29SC84bBeVzdE44qCgE0_rBjbNVCOxA3Ziuch-hs_miUGfD69nx9VXq6Uz2UrCDPhXVgcmJei5bHDuPo32pGhAq0UQ5qv-a6VStB57EBbTvL7l12mN2BprBRtqFKlMoiyKyzVV0ncLvW8RKUVrDMn9MqHp3Hq4LpSwnX9gPn97hqBGVb9SMrq8PGSXpRp5x9uk89fSRe8YF8hsf1gIO5q1mQGPOrLWNIRNzL6ZQ"

	headers := map[string]string{
		"Accept":        "application/json",
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + p.token,
	}
	systemPrompt := "你是一個擅長處理ESG指標的助手，請根據用戶的問題提供準確和有用的回答。"

	requestBody := RequestBody{
		Messages: []Message{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role: "user",
				Content: []MessageContent{
					{
						Type: "text",
						Text: prompt,
					},
				},
			},
		},
		// ProjectID:        "6e12cc37-c62b-4724-815c-d61499e89036",
		ProjectID:        util.GetEnvString("WATSONX_PROJECT_ID", ""),
		ModelID:          "meta-llama/llama-3-3-70b-instruct",
		FrequencyPenalty: 0,
		MaxTokens:        2000,
		PresencePenalty:  0,
		Temperature:      0,
		TopP:             1,
		Seed:             nil,
		Stop:             []string{},
	}

	// Serialize JSON body
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %v", err)
	}

	// Create request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	// Add headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("status error response: %d - %s", resp.StatusCode, string(body))
	}

	// Read response
	body, err := io.ReadAll(resp.Body)
	strBody := string(body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	return strBody, nil
}

// Close 關閉客戶端
func (p *WatsonxProvider) Close() {
	// OpenAI 客戶端不需要關閉操作
}

// CreateEmbedding 創建單個文本的嵌入向量
func (s *WatsonxProvider) CreateEmbedding(text string) ([]float32, error) {
	return nil, nil
}

// CreateBatchEmbeddings 批量創建文本的嵌入向量
func (s *WatsonxProvider) CreateBatchEmbeddings(texts []string) ([][]float32, error) {
	return nil, nil
}
