# AI Workshop

這是一個使用 Golang 實現的 AI 聊天應用程序，整合了 Gemini API 和 Milvus 向量數據庫。

project/
├── cmd/ # 應用程序入口
│ └── main.go
├── internal/ # 應用程序特定代碼
│ ├── config/ # 應用程序配置
│ ├── rag/ # RAG 特定邏輯
│ ├── chat/ # 聊天功能
│ └── documents/ # 文檔管理業務邏輯
├── pkg/ # 可被其他應用重用的包
│ ├── ai/ # AI 服務通用封裝
│ │ ├── openai/ # OpenAI API 封裝
│ │ └── gemini/ # Gemini API 封裝
│ ├── embeddings/ # 嵌入向量處理
│ ├── milvus/ # Milvus 客戶端
│ └── utils/ # 通用工具函數
└── static/ # 靜態資源

## 功能特點

- 使用 Gemini API 進行自然語言處理
- 使用 Milvus 進行向量存儲和檢索
- 提供 RESTful API 接口
- 簡單的 Web 界面

## 環境要求

- Go 1.21 或更高版本
- Docker 和 Docker Compose
- Gemini API 密鑰

## 配置

1. 設置 Gemini API 密鑰：

```bash
export GEMINI_API_KEY=你的API密鑰
export OPENAI_API_KEY=你的API密鑰
```

2. 使用 Docker Compose 啟動服務：

```bash
docker compose up -d
```

## 本地開發

1. 啟動依賴服務：

```bash
docker compose up -d milvus
```

2. 運行應用程序：

```bash
go run main.go
```

## API 端點

- `POST /api/chat` - 基本聊天功能
- `POST /api/rag` - RAG 問答功能

## 開發說明

- 使用 Gin 框架處理 HTTP 請求
- 使用 Gemini API 進行文本生成
- 使用 Milvus 進行向量存儲和檢索

系統架構與組件關係
embeddings 資料夾：

- 定義了EmbeddingType類型，用於標識不同的嵌入模型類型（OpenAI、Gemini等）
- 提供嵌入向量生成服務，將文本轉換為向量
- 支援多種模型：OpenAI Ada-002、OpenAI Embedding-3-Small/Large、Gemini
- 問題：缺少EmbeddingType的類型定義，已修復

documents 資料夾：

- 負責文檔的管理，包括存儲和檢索
- 使用Milvus作為向量數據庫，存儲文本和對應的嵌入向量
- 依賴embeddings服務來生成嵌入向量
- 根據不同的嵌入模型類型使用不同的集合名稱
- 問題：需要支持新增的嵌入模型類型，已更新

llm 資料夾：

- 定義了LLMType類型和LLMProvider接口
- 提供Factory工廠類，可以根據需要創建不同的LLM服務（OpenAI或Gemini）
- 實現了兩種LLM客戶端：OpenAIClient和GeminiClient
- 每個客戶端提供GenerateContent方法用於生成文本回應

rag 資料夾：

- RAG (Retrieval-Augmented Generation) 是整個系統的核心
- 依賴documents服務來檢索相關文檔
- 依賴embeddings服務生成查詢的嵌入向量
- 依賴llm服務生成最終回應
- 將前端請求的嵌入模型類型參數傳遞給下層組件

### milvus conflict

docker stop milvus-standalone
docker rm -f milvus-standalone
docker compose up --build
