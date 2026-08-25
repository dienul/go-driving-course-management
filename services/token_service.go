package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dienulhaq/go-driving-course-management/models"
	"github.com/golang-jwt/jwt/v5"
)

const tokenIssuer = "driving-course-management"

type Claims struct {
	UserID int64           `json:"user_id"`
	Role   models.UserRole `json:"role"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secret   []byte
	lifetime time.Duration
}

func NewJWTManager(secret, expiresIn string) (*JWTManager, error) {
	if len([]byte(strings.TrimSpace(secret))) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 bytes")
	}

	lifetime, err := time.ParseDuration(expiresIn)
	if err != nil || lifetime <= 0 {
		return nil, fmt.Errorf("JWT_EXPIRES_IN must be a positive Go duration")
	}

	return &JWTManager{
		secret:   []byte(secret),
		lifetime: lifetime,
	}, nil
}

func (m *JWTManager) Generate(user *models.User) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(m.lifetime)
	tokenID, err := randomTokenID()
	if err != nil {
		return "", time.Time{}, fmt.Errorf("generate token ID: %w", err)
	}

	claims := Claims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    tokenIssuer,
			Subject:   strconv.FormatInt(user.ID, 10),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        tokenID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign token: %w", err)
	}
	return signed, expiresAt, nil
}

func (m *JWTManager) Parse(rawToken string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(
		rawToken,
		claims,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, ErrInvalidToken
			}
			return m.secret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(tokenIssuer),
		jwt.WithExpirationRequired(),
		jwt.WithLeeway(30*time.Second),
	)
	if err != nil || !token.Valid || claims.UserID <= 0 {
		return nil, ErrInvalidToken
	}
	if claims.Subject != strconv.FormatInt(claims.UserID, 10) {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func randomTokenID() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return hex.EncodeToString(value), nil
}
