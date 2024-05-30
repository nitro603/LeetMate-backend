package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/youtube/v3"
)

var (
	youTubeAPIKey = "YOUR_YOUTUBE_API_KEY"
	chatGPTAPIKey = "YOUR_CHATGPT_API_KEY"
)

func main() {
	r := gin.Default()

	r.GET("/youtube/search", youtubeSearchHandler)
	r.POST("/chatgpt/query", chatGPTQueryHandler)

	r.Run(":8080")
}

func youtubeSearchHandler(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter is required"})
		return
	}

	client := &http.Client{
		Transport: &transport.APIKey{Key: youTubeAPIKey},
	}

	service, err := youtube.New(client)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	call := service.Search.List([]string).Q(query).MaxResults(5)
	response, err := call.Do()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func chatGPTQueryHandler(c *gin.Context) {
	var request struct {
		Query string `json:"query" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+chatGPTAPIKey).
		SetBody(map[string]interface{}{
			"model": "gpt-4",
			"messages": []map[string]string{
				{"role": "user", "content": request.Query},
			},
		}).
		Post("https://api.openai.com/v1/chat/completions")

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(resp.StatusCode(), resp.String())
}
