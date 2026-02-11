package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type CheckRequest struct {
	URL string `json:"url" binding:"required"`
}

// execCommandHelper executes the CLI command and returns the output
func (s *Server) execCommandHelper(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, s.cfg.Server.ExecPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	s.logger.Info("Executing command",
		zap.String("path", s.cfg.Server.ExecPath),
		zap.Strings("args", args),
	)

	err := cmd.Run()
	if err != nil {
		s.logger.Error("Command execution failed",
			zap.Error(err),
			zap.String("stderr", stderr.String()),
		)
		if stderr.Len() > 0 {
			return "", fmt.Errorf("command failed: %s, stderr: %s", err, stderr.String())
		}
		return "", err
	}
	return strings.TrimSpace(stdout.String()), nil
}

func (s *Server) pingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}

func (s *Server) timeHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": time.Now().UnixMilli()})
}

func (s *Server) versionHandler(c *gin.Context) {
	output, err := s.execCommandHelper(context.Background(), "version")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.String(http.StatusOK, output)
}

func (s *Server) homeHandler(c *gin.Context) {
	output, err := s.execCommandHelper(context.Background(), "home")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.String(http.StatusOK, output)
}

func (s *Server) supportHandler(c *gin.Context) {
	output, err := s.execCommandHelper(context.Background(), "support")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.String(http.StatusOK, output)
}

func (s *Server) helpHandler(c *gin.Context) {
	output, err := s.execCommandHelper(context.Background(), "help")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.String(http.StatusOK, output)
}

func (s *Server) checkHandler(c *gin.Context) {
	var req CheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Prepare command
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.Server.Timeout.Duration())
	defer cancel()

	cmd := exec.CommandContext(ctx, s.cfg.Server.ExecPath, req.URL)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	s.logger.Info("Executing command",
		zap.String("path", s.cfg.Server.ExecPath),
		zap.String("url", req.URL),
	)

	err := cmd.Run()
	if err != nil {
		s.logger.Error("Command execution failed",
			zap.Error(err),
			zap.String("stderr", stderr.String()),
		)

		// If context deadline exceeded
		if ctx.Err() == context.DeadlineExceeded {
			c.JSON(http.StatusGatewayTimeout, gin.H{"error": "command timed out"})
			return
		}

		// Try to parse stdout even if error, as CLI might return error JSON with exit code != 0?
		// Usually if CLI handles error gracefully it might return 0, but if it crashes or returns non-zero, we check stdout.
		if stdout.Len() == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "command failed", "details": stderr.String()})
			return
		}
	}

	// Parse stdout as JSON and return it directly
	var result json.RawMessage
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		s.logger.Error("Failed to parse CLI output", zap.String("output", stdout.String()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid output from cli", "output": stdout.String()})
		return
	}

	c.JSON(http.StatusOK, result)
}
