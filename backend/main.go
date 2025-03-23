

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// BookRequest represents the request body for book generation.
type BookRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Chapters    int    `json:"chapters"`
	ApiKey      string `json:"api_key,omitempty"` // API key is optional in the request.
}

// QwenMessage represents a message in the Qwen API request.
type QwenMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// QwenAPIRequest represents the request body for the Qwen API.
type QwenAPIRequest struct {
	Model        string `json:"model"`
	Input        struct {
		Messages []QwenMessage `json:"messages"`
	} `json:"input"`
	ResultFormat string `json:"result_format"`
}

// QwenResponse represents the response from the Qwen API.
type QwenResponse struct {
	Output struct {
		FinishReason string `json:"finish_reason"`
		Text         string `json:"text"`
	} `json:"output"`
}

func main() {
	// Load environment variables from .env file.
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: Error loading .env file:", err)
	}

	router := gin.Default()

	// Configure CORS
	router.Use(cors.Default())

	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello There!")
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	router.POST("/generate-book", generateBook)

	// Get port from environment variable or use default.
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Fatal(router.Run(":" + port))
}

func generateBook(c *gin.Context) {
	// Parse the request body using ShouldBindJSON.
	var req BookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Validate the request.
	if req.Title == "" || req.Description == "" || req.Chapters <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Title, description, and chapters are required",
		})
		return
	}

	// Get the API key from the request or the environment variable.
	apiKey := req.ApiKey
	if apiKey == "" {
		apiKey = os.Getenv("QWEN_API_KEY")
		if apiKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "API key is required either in request or environment variable",
			})
			return
		}
	}

	// Create the system prompt.
	systemPrompt := fmt.Sprintf(`You are a professional book writer.
Generate a complete book with the following details:
- Title: %s
- Description: %s
- Number of chapters: %d

The book should have a coherent narrative that follows the description.
Each chapter should have a title and substantial content.
Format the book with proper Markdown, including headings for chapters.
Create a compelling opening and satisfying conclusion.`, req.Title, req.Description, req.Chapters)

	// Create the Qwen API request payload.
	var qwenReq QwenAPIRequest
	qwenReq.Model = "qwen-plus"
	qwenReq.ResultFormat = "message"
	qwenReq.Input.Messages = []QwenMessage{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Please generate a complete book titled '%s' with %d chapters based on this description: %s", req.Title, req.Chapters, req.Description),
		},
	}

	// Call the Qwen API.
	bookContent, err := callQwenAPI(qwenReq, apiKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate book: " + err.Error(),
		})
		return
	}

	// Return the generated book.
	c.JSON(http.StatusOK, gin.H{
		"book": bookContent,
	})
}

func callQwenAPI(req QwenAPIRequest, apiKey string) (string, error) {
	// Marshal the request payload into JSON.
	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create the HTTP request with the proper Qwen API endpoint.
	url := "https://dashscope-intl.aliyuncs.com/api/v1/services/aigc/text-generation/generation"
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Create a context with a 5-minute timeout and attach it to the request.
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()
	httpReq = httpReq.WithContext(ctx)

	// Set the required headers.
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 300 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse the JSON response.
	var qwenResp QwenResponse
	if err := json.Unmarshal(body, &qwenResp); err != nil {
		return "", fmt.Errorf("failed to parse JSON response: %w, response body: %s", err, string(body))
	}

	if qwenResp.Output.Text == "" {
		return "", fmt.Errorf("API returned an empty or invalid response: %+v", qwenResp)
	}

	return qwenResp.Output.Text, nil
}

