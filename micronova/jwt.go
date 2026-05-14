package micronova

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func getJwtExpiration(tokenString string) (time.Time, error) {
	if tokenString == "" {
		return time.Time{}, errors.New("Token is empty")
	}

	claims := jwt.MapClaims{}
	_, _, err := jwt.NewParser().ParseUnverified(tokenString, claims)
	if err != nil {
		return time.Time{}, fmt.Errorf("JWT parse error: %w", err)
	}

	// for key, val := range claims {
	// 	log.Trace().Msgf("JWT Claim %v = %v", key, val)
	// }

	// Extract the 'exp' claim
	exp, ok := claims["exp"]
	if !ok {
		return time.Time{}, errors.New("JWT missing 'exp' claim")
	}

	// Handle numeric expiration time
	var expiration time.Time
	switch expType := exp.(type) {
	case float64:
		expiration = time.Unix(int64(expType), 0)
	case json.Number:
		val, err := expType.Int64()
		if err != nil {
			return time.Time{}, fmt.Errorf("JWT 'exp' claim is not a valid integer: %w", err)
		}
		expiration = time.Unix(val, 0)
	default:
		return time.Time{}, fmt.Errorf("JWT 'exp' claim has unexpected type: %T", expType)
	}

	return expiration, nil
}
