package routes

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sehwan505/codev42/configs"
	"github.com/sehwan505/codev42/services/agent"
)

func SetupRouter() (*gin.Engine, error) {
	r := gin.Default()
	config, err := configs.GetConfig()
	if err != nil {
		return nil, err
	}

	r.POST("/generate-plan", func(c *gin.Context) {
		var request struct {
			Prompt string `json:"prompt" binding:"required"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		masterAgent := agent.NewMasterAgent(config.OpenAiKey)

		devPlan, err := masterAgent.Call(request.Prompt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to generate plan: %v", err)})
			return
		}
		c.JSON(http.StatusOK, devPlan)
	})

	r.POST("/implement-plan", func(c *gin.Context) {
		var request agent.DevPlan

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		workerAgent := agent.NewWorkerAgent(config.OpenAiKey)
		results, err := workerAgent.ImplementPlan(&request)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to implement plan: %v", err)})
			return
		}
		c.JSON(http.StatusOK, results)
	})

	return r, nil
}
