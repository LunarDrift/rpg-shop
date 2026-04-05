// Package auth .
package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassword(password string) (string, error) {
	hashedPass, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}
	return hashedPass, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, hash)
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	// set up claims struct
	expiration := time.Now().UTC().Add(expiresIn)
	claims := jwt.RegisteredClaims{
		Issuer:    "shop-access",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(expiration),
		Subject:   userID.String(),
	}

	// create new token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// create signed JWT
	return token.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	// create pointer to empty RegisteredClaims to pass in to ParseWithClaims
	// gets populated during parsing; how we get the Subject out after
	claims := &jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.UUID{}, err
	}
	if !token.Valid {
		return uuid.UUID{}, errors.New("invalid token")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.UUID{}, err
	}

	return userID, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	// look for the Authorization header
	authHeader := headers.Get("Authorization")

	// check if header exists
	if authHeader == "" {
		return "", errors.New("authorization header not found")
	}

	// strip prefix and whitespace
	token := strings.TrimPrefix(authHeader, "Bearer ")

	// check header again in case it's not a Bearer token
	if token == authHeader {
		return "", errors.New("authorization header is not a bearer token")
	}

	return token, nil
}

func GetAPIKey(headers http.Header) (string, error) {
	// look for the Authorization header
	authHeader := headers.Get("Authorization")

	// check if header exists
	if authHeader == "" {
		return "", errors.New("authorization header not found")
	}

	// strip prefix and whitespace
	apiKey := strings.TrimPrefix(authHeader, "ApiKey ")

	// check header again in case it's not an ApiKey
	if apiKey == authHeader {
		return "", errors.New("authorization header is not an ApiKey")
	}

	return apiKey, nil
}

func MakeRefreshToken() string {
	// generate 32 bytes of random data
	b := make([]byte, 32)
	_, _ = rand.Read(b)

	// convert to a hex string
	return hex.EncodeToString(b)
}
