package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"codev42/routes"
	"codev42/services/agent"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGeneratePlanHTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, err := routes.SetupRouter()
	assert.NoError(t, err)

	requestBody := map[string]string{"prompt": "make code for calculator"}
	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest(http.MethodPost, "/generate-plan", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response agent.DevPlan
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, []string{}, response.Annotations)
}

func TestImplementPlanHTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, err := routes.SetupRouter()
	assert.NoError(t, err)

	requestBody := agent.DevPlan{
		Annotations: []string{"func1", "func2"},
	}
	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest(http.MethodPost, "/implement-plan", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []agent.DevResult
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)
	// for i := 0; i < len(response); i++ {
	// 	assert.Equal(t, string, response[i].Description)
	// 	assert.Equal(t, string, response[i].Code)
	// }
}
