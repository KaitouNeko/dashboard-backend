package llm

import "errors"

var (
	// ErrAPIKeyNotConfigured 表示 API Key 未配置
	ErrAPIKeyNotConfigured = errors.New("API key not configured")
	// ErrNoValidResponse 表示沒有有效的回應
	ErrNoValidResponse = errors.New("no valid response generated")
)
