package yps

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Logs a failed login attempt.
func LogFailedLoginAttempt(address string) error {
	return Log(LogLevelInfo, "login-failed", "Failed login attempt", map[string]string{
		"address": address,
	})
}

// Logs a successful login attempt.
func LogSuccessfulLoginAttempt(level string) error {
	return Log(LogLevelInfo, "login-success", "Successful login attempt for "+level, map[string]string{
		"level": level,
	})
}

// Logs a new line.
func Log(logLevel LogLevel, eventType string, message string, data interface{}) error {
	return TheDb.AddLogLine(logLevel, eventType, message, data)
}

// handler

type LogsRequest struct {
	Page int `form:"page"`
}

type LogsResponse struct {
	Logs []LogLine `json:"logs"`
}

func getLogs(c *gin.Context) {
	var req SearchRequest
	if err := c.ShouldBind(&req); err != nil {
		fmt.Println("Could not get logs binding:", err.Error())
		c.JSON(400, gin.H{"error": "Logs params not found"})
		return
	}

	response, err := TheDb.GetLogs(req.Page)
	if err != nil {
		fmt.Println("Could not get logs:", err.Error())
		c.JSON(400, gin.H{"error": "Could not get logs"})
		return
	}

	c.JSON(http.StatusOK, map[string]any{
		"logs": response,
	})
}
