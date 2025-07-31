package service

import (
	"backend/internal/config"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenType represents the type of JWT token
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// JWTClaims defines the structure of the JWT claims.
type JWTClaims struct {
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

// TokenPair represents a pair of access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// JwtService handles JWT generation and validation.
type JwtService interface {
	// GenerateTokenPair creates both access and refresh tokens for a user
	GenerateTokenPair(userID uuid.UUID, username string) (*TokenPair, error)
	// GenerateAccessToken creates a new access token for a given user
	GenerateAccessToken(userID uuid.UUID, username string) (string, error)
	// GenerateRefreshToken creates a new refresh token for a given user
	GenerateRefreshToken(userID uuid.UUID, username string) (string, error)
	// ValidateToken validates a JWT string and returns the claims if valid
	ValidateToken(tokenString string) (*JWTClaims, error)
	// GetTokenRemainingTTL calculates the remaining time until token expiration
	GetTokenRemainingTTL(tokenString string) (time.Duration, error)
}

// jwtService is the implementation of JwtService.
type jwtService struct {
	secretKey                      []byte
	accessTokenExpirationInMinutes int
	refreshTokenExpirationInDays   int
}

// NewJwtService creates a new instance of JwtService.
func NewJwtService(cfg *config.SecurityConfig) JwtService {
	return &jwtService{
		secretKey:                      []byte(cfg.JwtSecret),
		accessTokenExpirationInMinutes: cfg.JwtAccessTokenExpiresInMinutes,
		refreshTokenExpirationInDays:   cfg.JwtRefreshTokenExpiresInDays,
	}
}

// GenerateTokenPair creates both access and refresh tokens for a user
func (s *jwtService) GenerateTokenPair(userID uuid.UUID, username string) (*TokenPair, error) {
	accessToken, err := s.GenerateAccessToken(userID, username)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.GenerateRefreshToken(userID, username)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// GenerateAccessToken creates a new access token for a given user.
func (s *jwtService) GenerateAccessToken(userID uuid.UUID, username string) (string, error) {
	return s.generateToken(userID, username, AccessToken, time.Duration(s.accessTokenExpirationInMinutes)*time.Minute)
}

// GenerateRefreshToken creates a new refresh token for a given user.
func (s *jwtService) GenerateRefreshToken(userID uuid.UUID, username string) (string, error) {
	return s.generateToken(userID, username, RefreshToken, time.Duration(s.refreshTokenExpirationInDays)*24*time.Hour)
}

// generateToken is a helper method to generate tokens with specific type and duration
func (s *jwtService) generateToken(userID uuid.UUID, username string, tokenType TokenType, duration time.Duration) (string, error) {
	// Set custom claims
	claims := &JWTClaims{
		UserID:    userID,
		Username:  username,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "backend-app",
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and return it
	signedToken, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// ValidateToken validates a JWT string.
func (s *jwtService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// GetTokenRemainingTTL calculates the remaining time until token expiration
func (s *jwtService) GetTokenRemainingTTL(tokenString string) (time.Duration, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return 0, fmt.Errorf("invalid token: %w", err)
	}

	// Calculate remaining time until expiration
	expirationTime := claims.ExpiresAt.Time
	remainingTime := time.Until(expirationTime)
	
	// If token has already expired, return 0
	if remainingTime <= 0 {
		return 0, fmt.Errorf("token has already expired")
	}

	return remainingTime, nil
} 