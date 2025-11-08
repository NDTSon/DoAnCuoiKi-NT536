package auth

import (
    "fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenGenerator struct {
	issuer     string
	signingKey []byte
}

func NewTokenGenerator(issuer, key string) *TokenGenerator {
	return &TokenGenerator{issuer: issuer, signingKey: []byte(key)}
}

func (t *TokenGenerator) Generate(userID string, ttl time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"iss": t.issuer,
		"exp": time.Now().Add(ttl).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(t.signingKey)
}

// Parse validates a token string and returns its claims when valid.
func (t *TokenGenerator) Parse(tokenString string) (jwt.MapClaims, error) {
    claims := jwt.MapClaims{}
    _, err := jwt.ParseWithClaims(
        tokenString,
        claims,
        func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
            }
            return t.signingKey, nil
        },
        jwt.WithIssuer(t.issuer),
        jwt.WithLeeway(30*time.Second),
        jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
    )
    if err != nil {
        return nil, err
    }
    return claims, nil
}
