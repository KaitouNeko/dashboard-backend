package config

import (
	"ai-workshop/pkg/util"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/joho/godotenv/autoload" // package that loads env
)

type Config struct {
	// llm api keys
	GeminiAPIKey string
	OpenAiAPIKey string

	// vector db config
	MilvusHost     string
	MilvusPort     string
	MilvusRESTPort string
	MilvusUsername string
	MilvusPassword string

	// postgres db
	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDBName   string

	// for transmitting goroutine errors
	errorChan           chan error
	cancelFunc          context.CancelFunc
	MilvusServiceHealty bool
}

func NewConfig() *Config {
	config := &Config{
		errorChan: make(chan error),
	}

	err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config. Error: %s", err)
	}

	// start a goroutine to check milvus health status
	ctx, cancel := context.WithCancel(context.Background())
	config.cancelFunc = cancel

	go config.monitorMilvusService(ctx)

	// start goroutine to cehck milvus erors
	go config.handleErrors()

	return config
}

/**
* Sets up all initial configuration such as LLM keys and vector DB connections.
**/
func (c *Config) LoadConfig() error {
	// --- Config Setup ---

	// -- llm keys --
	c.GeminiAPIKey = util.GetEnvString("GEMINI_API_KEY", "")
	c.OpenAiAPIKey = util.GetEnvString("OPENAI_API_KEY", "")

	fmt.Printf("env: %s\n", c.OpenAiAPIKey)
	fmt.Printf("env: %s\n", c.GeminiAPIKey)

	if c.GeminiAPIKey == "" {
		log.Fatal("GEMINI_API_KEY 環境變量未設置")
	}
	if c.OpenAiAPIKey == "" {
		log.Fatal("OPENAI_API_KEY 環境變量未設置")
	} else {
		log.Printf(c.OpenAiAPIKey + "OPENAI_API_KEY 環境變量已設置")
	}

	// -- vector db --
	c.MilvusHost = util.GetEnvString("MILVUS_HOST", "localhost")
	c.MilvusPort = util.GetEnvString("MILVUS_PORT", "19530")
	c.MilvusRESTPort = util.GetEnvString("MILVUS_REST_PORT", "9091")
	c.MilvusUsername = util.GetEnvString("MILVUS_USERNAME", "root")
	c.MilvusPassword = util.GetEnvString("MILVUS_PASSWORD", "milvus")

	log.Printf("Successfully loaded Gemini API Key")
	log.Printf("Running in local development mode")
	log.Printf("Milvus configuration: host=%s, port=%s, rest_port=%s",
		c.MilvusHost,
		c.MilvusPort,
		c.MilvusRESTPort)

	// -- Postgres DB --
	c.PostgresHost = util.GetEnvString("POSTGRES_HOST", "localhost")
	c.PostgresHost = util.GetEnvString("POSTGRES_HOST", "127.0.0.1")
	c.PostgresPort = util.GetEnvString("POSTGRES_PORT", "5555")
	c.PostgresUser = util.GetEnvString("POSTGRES_USER", "user")
	c.PostgresPassword = util.GetEnvString("POSTGRES_PASSWORD", "password")
	c.PostgresDBName = util.GetEnvString("POSTGRES_DB", "ai_poc_db")

	return nil
}

/**
* Checks if milvus service is up and healthy.
**/
func (c *Config) monitorMilvusService(ctx context.Context) {
	log.Printf("initiated monitoring milvus service health...")

	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	// first check
	c.checkMilvusHealth()

	for {
		select {
		case <-ticker.C:
			c.checkMilvusHealth() // check again every 10 seconds
		case <-ctx.Done():
			log.Printf("Cancelling health check, exiting...")
			// stop infinite loop
			return
		}
	}
}

/**
* checks mulvis health.
**/
func (c *Config) checkMilvusHealth() {
	resp, err := http.Get(fmt.Sprintf("http://%s:%s/api/v1/health", c.MilvusHost, c.MilvusRESTPort))

	if err != nil {
		c.MilvusServiceHealty = false
		// transmit error through channel
		c.errorChan <- fmt.Errorf("Cannot connect to Milvus API: %v", err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.MilvusServiceHealty = false
		c.errorChan <- fmt.Errorf("Milvus service is unhealthy: HTTP %d", resp.StatusCode)
		return
	}

	if !c.MilvusServiceHealty {
		log.Printf("Milvus service is now healthy")
	}

	c.MilvusServiceHealty = true
}

/**
* concurrently reads all errors snt from checkMilvusService
**/
func (c *Config) handleErrors() {
	for err := range c.errorChan {
		fmt.Errorf("Error detected with mulvis service during health check: %s\n", err)
	}
}

/**
* Clean up goroutine and channels.
**/
func (c *Config) CleanUp() {
	if c.cancelFunc != nil {
		c.cancelFunc() // stop monitoring goroutine
	}

	time.Sleep(time.Millisecond * 100)
	close(c.errorChan)

	log.Println("Cleanup up config resources.")
}
