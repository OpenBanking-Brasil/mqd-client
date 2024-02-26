package jwt

import (
	"bytes"
	"encoding/json"
	"io"
	"strconv"
	"time"

	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/log"
	"github.com/golang-jwt/jwt/v5"
)

// JWKToken Struct that stores the properties oj a JWT token
type JWKToken struct {
	AccessToken      string `json:"access_token"`       // Access token to be used
	TokenType        string `json:"token_type"`         // Type of token
	ExpiresIn        int    `json:"expires_in"`         // Indicates the expiration of the token
	RefreshExpiresIn int    `json:"refresh_expires_in"` // Indicates the expiration of the refresh token
	NotBeforePolicy  int    `json:"not-before-policy"`  // Indicates the time before which the token cannot be used
	Scope            string `json:"scope"`              // Scope of the token
}

// ValidateExpiration validates the expiration date of a jwt token
//
// Parameters:
//   - logger: Logger to be used
//   - token: JWT token to be validated
//
// Returns:
//   - bool: true if token is still valid
func ValidateExpiration(logger log.Logger, token *JWKToken) bool {
	logger.Info("Validating expiration", "jwt", "validateExpiration")

	if token == nil {
		logger.Info("Empty token.", "jwt", "validateExpiration")
		return false
	}

	parsedToken, _, err := jwt.NewParser().ParseUnverified(token.AccessToken, jwt.MapClaims{})
	if err != nil {
		logger.Error(err, "Error parsing or validating token", "jwt", "validateExpiration")
		return false
	}

	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
		expirationTime := time.Unix(int64(claims["exp"].(float64)), 0)
		logger.Debug("Token expiration time: "+expirationTime.String(), "jwt", "ValidateToken")
		currentTime := time.Now()
		if currentTime.After(expirationTime) {
			logger.Info("jwt token has expired", "jwt", "validateExpiration")
			return false
		}
	} else {
		logger.Info("Invalid JWT token", "jwt", "validateExpiration")
		return false
	}

	return true
}

// GetTokenFromReader reads a jwt token from a reader
//
// Parameters:
//   - logger: Logger to be used
//   - reader: Reader that contains the information of the token
//
// Returns:
//   - JWKToken: Token read from the reader
//   - error: error if any
func GetTokenFromReader(logger log.Logger, reader io.Reader) (*JWKToken, error) {
	result := &JWKToken{}

	// Decode the JSON response into a JWKToken struct
	err := json.NewDecoder(reader).Decode(&result)
	if err != nil {
		logger.Error(err, "Error decoding JSON response", "jwt", "GetTokenFromReader")
		return nil, err
	}

	// Access the fields of the JWKToken object
	logger.Debug("Access Token: "+result.AccessToken, "jwt", "GetTokenFromReader")
	logger.Debug("Token Type: "+result.TokenType, "jwt", "GetTokenFromReader")
	logger.Debug("Expires In: "+strconv.Itoa(result.ExpiresIn), "jwt", "GetTokenFromReader")
	logger.Debug("Refresh Token: "+strconv.Itoa(result.RefreshExpiresIn), "jwt", "GetTokenFromReader")

	return result, nil
}

// GetTokenFromBinary reads a jwt token from a reader
//
// Parameters:
//   - logger: Logger to be used
//   - data: Byte array with the token information
//
// Returns:
//   - JWKToken: Token read from the reader
//   - error: error if any
func GetTokenFromBinary(logger log.Logger, data []byte) (*JWKToken, error) {
	result := &JWKToken{}

	// Create a reader from the binary array
	reader := bytes.NewReader(data)

	// Decode the JSON response into a JWKToken struct
	err := json.NewDecoder(reader).Decode(&result)
	if err != nil {
		logger.Error(err, "Error decoding JSON response", "jwt", "GetTokenFromReader")
		return nil, err
	}

	// Access the fields of the JWKToken object
	logger.Debug("Access Token: "+result.AccessToken, "jwt", "GetTokenFromReader")
	logger.Debug("Token Type: "+result.TokenType, "jwt", "GetTokenFromReader")
	logger.Debug("Expires In: "+strconv.Itoa(result.ExpiresIn), "jwt", "GetTokenFromReader")
	logger.Debug("Refresh Token: "+strconv.Itoa(result.RefreshExpiresIn), "jwt", "GetTokenFromReader")

	return result, nil
}
