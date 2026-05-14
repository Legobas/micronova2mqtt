package micronova

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// --- Tests ---

func TestGetJwtExpiration_EmptyToken(t *testing.T) {
	_, err := getJwtExpiration("")
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestGetJwtExpiration_MissingExp(t *testing.T) {
	// create token with no exp
	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		"sub": "123",
	})
	tokenStr, _ := token.SigningString()       // unsigned format; ParseUnverified only needs a token string form
	_, err := getJwtExpiration(tokenStr + ".") // ensure parse form; ParseUnverified accepts typical JWT forms
	if err == nil {
		t.Fatal("expected error for missing exp claim")
	}
}

func TestGetJwtExpiration_ExpFloat64(t *testing.T) {
	// create token with exp as numeric (float64)
	exp := time.Now().Add(1 * time.Hour).Unix()
	claims := jwt.MapClaims{"exp": float64(exp)}
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenStr, _ := token.SigningString()
	expTime, err := getJwtExpiration(tokenStr + ".")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if expTime.Unix() != exp {
		t.Fatalf("expected %d got %d", exp, expTime.Unix())
	}
}
