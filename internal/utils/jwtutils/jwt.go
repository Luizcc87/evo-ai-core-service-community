package jwtutils

import (
	"errors"
	"evo-ai-core-service/internal/config"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type CustomClaims struct {
	Email  string    `json:"email"`
	Role   string    `json:"role"`
	UserID uuid.UUID `json:"user_id"`
	Name   string    `json:"name"`
	jwt.RegisteredClaims
}

var signingMethodsMap = map[string]jwt.SigningMethod{
	"HS256": jwt.SigningMethodHS256,
	"HS384": jwt.SigningMethodHS384,
	"HS512": jwt.SigningMethodHS512,
	"RS256": jwt.SigningMethodRS256,
	"RS384": jwt.SigningMethodRS384,
	"RS512": jwt.SigningMethodRS512,
	"ES256": jwt.SigningMethodES256,
	"ES384": jwt.SigningMethodES384,
	"ES512": jwt.SigningMethodES512,
}

func getSigningMethodFromString(alg string) (jwt.SigningMethod, error) {
	method, ok := signingMethodsMap[alg]
	if !ok {
		return nil, fmt.Errorf("JWT algorithm not supported or configured: %s", alg)
	}

	return method, nil
}

type JWTUtils interface {
	Parse(token string) (*CustomClaims, error)
	GenerateToken(claims *CustomClaims) (string, error)
}

type jwtutils struct {
	cfgJWT *config.JWTConfig
}

func NewJWTUtils(cfgJWT *config.JWTConfig) JWTUtils {
	return &jwtutils{
		cfgJWT: cfgJWT,
	}
}

func (j *jwtutils) Parse(token string) (*CustomClaims, error) {

	if token == "" || !strings.HasPrefix(token, "Bearer ") {
		return nil, errors.New("Missing or invalid token")
	}

	token = strings.TrimPrefix(token, "Bearer ")
	signingMethod, err := getSigningMethodFromString(j.cfgJWT.Algorithm)

	if err != nil {
		return nil, errors.New("Invalid token")
	}

	tokenParsed, err := jwt.ParseWithClaims(token, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != signingMethod.Alg() {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.cfgJWT.SecretKey), nil
	})

	if err != nil || !tokenParsed.Valid {
		log.Printf("Error parsing token: %v", err)
		return nil, errors.New("Invalid token")
	}

	claims, ok := tokenParsed.Claims.(*CustomClaims)
	if !ok {
		return nil, errors.New("Invalid token")
	}

	if claims.Subject == "" {
		return nil, errors.New("Token without id (sub)")
	}

	_, errUUID := uuid.Parse(claims.Subject)
	if errUUID != nil {
		return nil, errors.New("Token with invalid id (sub)")
	}

	if claims.Email == "" {
		return nil, errors.New("Token without email")
	}

	if claims.ExpiresAt == nil || claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("Expired Token")
	}

	return claims, nil
}

func (j *jwtutils) GenerateToken(claims *CustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.cfgJWT.SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
