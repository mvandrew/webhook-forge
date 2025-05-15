package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"webhook-forge/pkg/logger"
)

// TokenGenerator generates secure tokens
type TokenGenerator struct {
	logger logger.Logger
}

// NewTokenGenerator creates a new token generator
func NewTokenGenerator(logger logger.Logger) *TokenGenerator {
	return &TokenGenerator{
		logger: logger,
	}
}

// GenerateToken generates a random token using current time and random bytes
// This method is similar to HookService.GenerateToken to maintain consistency
func (g *TokenGenerator) GenerateToken() string {
	// Get current time as part of the token generation
	timestamp := time.Now().UnixNano()

	// Create a random component (16 bytes = 32 hex chars)
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		// If there's an error reading random, fallback to less random but still useful method
		g.logger.Error("Failed to generate random bytes for token", logger.Field{Key: "error", Value: err.Error()})
		randomBytes = []byte(fmt.Sprintf("%016x", timestamp))
	}

	// Combine timestamp and random component
	token := fmt.Sprintf("%x-%s", timestamp, hex.EncodeToString(randomBytes))

	return token
}
