package result

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/OpenBanking-Brasil/MQD_Client/configuration"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/log"
	"github.com/golang-jwt/jwt/v5"
)

const CERTIFICATE_PATH = "ParameterData//certificates//" //
const TOKEN_PATH = "/token"
const REPORT_PATH = "/report"

const ENV_CLIENT_CRT_FILE = "CLIENT_CRT_FILE" // Certificate file
const ENV_CLIENT_KEY_FILE = "CLIENT_KEY_FILE" // Private key file

// JWKToken struct
type JWKToken struct {
	AccessToken      string `json:"access_token"`       // Access token to be used
	TokenType        string `json:"token_type"`         // Type of token
	ExpiresIn        int    `json:"expires_in"`         // Indicates the expiration of the token
	RefreshExpiresIn int    `json:"refresh_expires_in"` // Indicates the expiration of the refresh token
	NotBeforePolicy  int    `json:"not-before-policy"`  // Indicates the time before which the token cannot be used
	Scope            string `json:"scope"`              // Scope of the token
}

var (
	certificates = tls.Certificate{}
	token        = &JWKToken{}
)

// Func: loadCertificates Load certificate files (crt and key) into memmory for further use
// @author AB
// @params
// @return
func loadCertificates() {
	log.Info("Loading certificates", "Server", "loadCertificates")

	certFile := crosscutting.GetEnvironmentValue(ENV_CLIENT_CRT_FILE, "client.crt")
	keyFile := crosscutting.GetEnvironmentValue(ENV_CLIENT_KEY_FILE, "client.key")

	keyPath := fmt.Sprintf("%s%s", CERTIFICATE_PATH, keyFile)
	crtPath := fmt.Sprintf("%s%s", CERTIFICATE_PATH, certFile)

	// Load client certificate and private key
	cert, err := tls.LoadX509KeyPair(crtPath, keyPath)
	if err != nil {
		log.Fatal(err, "Error loading client certificate and key", "Server", "loadCertificates")
		return
	}

	certificates = cert
}

// Func: getHttpClient Returns a client configured to use certificates for mTLS communication
// @author AB
// @return
// http client: Client created with certificate info
func getHttpClient() *http.Client {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates:       []tls.Certificate{certificates},
				InsecureSkipVerify: true,
			},
		},
	}

	return httpClient
}

// Func: validateExpiration validates the expiration date of a jwt token
// @author AB
// @params
// token: JWT token to be validated
// @return
// bool: true if token is still valid
func validateExpiration(token *JWKToken) bool {
	log.Info("Validating expiration", "server", "validateExpiration")

	parsedToken, _, err := jwt.NewParser().ParseUnverified(token.AccessToken, jwt.MapClaims{})
	if err != nil {
		log.Error(err, "Error parsing or validating token", "server", "validateExpiration")
		return false
	}

	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
		expirationTime := time.Unix(int64(claims["exp"].(float64)), 0)
		log.Debug("Token expiration time: "+expirationTime.String(), "jwt", "ValidateToken")
		currentTime := time.Now()
		if currentTime.After(expirationTime) {
			log.Info("jwt token has expired", "server", "validateExpiration")
			return false
		}
	} else {
		log.Info("Invalid JWT token", "server", "validateExpiration")
		return false
	}

	return true
}

func GetJWKToken() (*JWKToken, error) {
	log.Info("Loading JWT token", "server", "GetJWKToken")

	httpClient := getHttpClient()

	if token.AccessToken != "" && validateExpiration(token) {
		log.Info("Token is valid, using previous token", "server", "GetJWKToken")
		return token, nil
	}

	log.Info("Token is invalid, Requesting new token", "server", "GetJWKToken")

	// Create the request body with grant_type and client_id
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", configuration.ClientID)
	requestBody := strings.NewReader(data.Encode())

	request, err := http.NewRequest("POST", configuration.ServerURL+TOKEN_PATH, requestBody)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}

	// Set the Content-Type header to application/x-www-form-urlencoded
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send the request
	resp, err := httpClient.Do(request)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Check the response
	if resp.StatusCode == http.StatusOK {
		// Decode the JSON response into a JWKToken struct
		err := json.NewDecoder(resp.Body).Decode(&token)
		if err != nil {
			fmt.Println("Error decoding JSON response:", err)
			return nil, err
		}

		// Access the fields of the JWKToken object
		log.Debug("Access Token: "+token.AccessToken, "server", "GetJWKToken")
		log.Debug("Token Type: "+token.TokenType, "server", "GetJWKToken")
		log.Debug("Expires In: "+strconv.Itoa(token.ExpiresIn), "server", "GetJWKToken")
		log.Debug("Refresh Token: "+strconv.Itoa(token.RefreshExpiresIn), "server", "GetJWKToken")
	} else {
		log.Warning("Request failed with status code: "+strconv.Itoa(resp.StatusCode), "server", "GetJWKToken")
		return nil, errors.New("request failed with status code:" + strconv.Itoa(resp.StatusCode))
	}

	return token, nil
}

// Func: postReport sends the report to the server using required authorization
// @author AB
// @params
// report: Report to send
// @return
// error: will be != of nil in case of error
func postReport(report Report) error {
	log.Info("Posting report", "server", "postReport")

	httpClient := getHttpClient()

	requestBody, err := json.Marshal(report)
	if err != nil {
		return err
	}

	// Create a new request
	req, err := http.NewRequest("POST", configuration.ServerURL+REPORT_PATH, bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return err
	}

	// Set the Authorization header with your token
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	// Send the request
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Error(err, "Error sending report.", "Result", "sendReportToAPI")
		return err
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		log.Warning("Error sending report, Status code: "+fmt.Sprint(resp.StatusCode), "Result", "sendReportToAPI")
	} else {
		// Read the body of the message
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		log.Info(string(body), "Result", "sendReportToAPI")
	}

	return nil
}

// Func: sendReportToAPI sends the report to server API, executes the process of requesting / validating the token
// @author AB
// @params
// report: Report to send
// @return
// error: will be != of nil in case of error
func sendReportToAPI(report Report) error {
	log.Info("Sending report to central Server", "Result", "sendReportToAPI")

	_, err := GetJWKToken()
	if err != nil {
		return err
	}

	err = postReport(report)
	if err != nil {
		return err
	}

	return nil
}
