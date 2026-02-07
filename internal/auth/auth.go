package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenType string

const (
	// TokenTypeAccess -
	TokenTypeAccess TokenType = "chirpy-access"
)

func HashPassword(password string) (string, error) {
	hashedPW, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}

	return hashedPW, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	pass, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}
	return pass, nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	signingKey := []byte(tokenSecret)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    string(TokenTypeAccess),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:   userID.String(),
	})
	return token.SignedString(signingKey)
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claimsStruct := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &claimsStruct, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, err
	}

	userIDstring, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		return uuid.Nil, err
	}
	if issuer != string(TokenTypeAccess) {
		return uuid.Nil, fmt.Errorf("invalid issuer")
	}

	id, err := uuid.Parse(userIDstring)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	tokenHeader := headers.Get("Authorization")

	token := strings.Split(tokenHeader, "Bearer ")
	if len(token) != 2 {
		return "", fmt.Errorf("invalid token format")
	}

	return token[1], nil
}

func TokenAuth(headers http.Header, tokenSecret string) (uuid.UUID, error) {
	httpToken, err := GetBearerToken(headers)
	if err != nil {
		return uuid.Nil, fmt.Errorf("error extracting token: %v", err)
	}

	userID, err := ValidateJWT(httpToken, tokenSecret)
	if err != nil {
		return uuid.Nil, fmt.Errorf("error validating token: %v", err)
	}

	return userID, nil
}

func MakeRefeshToken() (string, error) {
	tokenByte := make([]byte, 32)
	rand.Read(tokenByte)
	return hex.EncodeToString(tokenByte), nil
}

func GetAPIKey(headers http.Header) (string, error) {
	keyHeader := headers.Get("Authorization")
	key := strings.Split(keyHeader, "ApiKey ")
	if len(key) != 2 {
		return "", fmt.Errorf("invalid API key format")
	}
	return key[1], nil
}
