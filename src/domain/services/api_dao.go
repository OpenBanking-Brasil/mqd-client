package services

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/log"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/security/jwt"
)

// RestAPI is the struct to handle connections to APIs
type RestAPI struct {
	crosscutting.OFBStruct               // Base structure
	token                  *jwt.JWKToken // Token used by the server
	serverURL              string
}

// loadCertificates Loads certificates from environment variables
// @author AB
// @params
// @return
// error: Error if any
// Response from server in case of success
func (ad *RestAPI) requestNewJWTToken(clientID string) (*jwt.JWKToken, error) {
	ad.Logger.Info("Requesting new token", ad.Pack, "requestNewJWTToken")

	// Create an HTTP client
	client := &http.Client{}

	// Define the parameters for the token request
	params := url.Values{}
	params.Set("grant_type", "client_credentials")
	params.Set("client_id", clientID)
	requestBody := params.Encode()

	ad.Logger.Debug("ServerURL:"+ad.serverURL+tokenPath, ad.Pack, "requestNewJWTToken")
	ad.Logger.Debug("Body:"+requestBody, ad.Pack, "requestNewJWTToken")

	// Create a new HTTP request
	req, err := http.NewRequest("POST", ad.serverURL+tokenPath, strings.NewReader(requestBody))
	if err != nil {
		ad.Logger.Error(err, "Error creating request", ad.Pack, "requestNewJWTToken")
		return nil, err
	}

	// Set the content type header
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send the request
	response, err := client.Do(req)
	if err != nil {
		ad.Logger.Error(err, "Error sending request", ad.Pack, "requestNewJWTToken")
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			ad.Logger.Error(err, "Error closing body", ad.Pack, "requestNewJWTToken")
		}
	}(response.Body)

	var result *jwt.JWKToken
	if response.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		result, _ = jwt.GetTokenFromBinary(ad.Logger, bodyBytes)
	} else {
		ad.Logger.Warning("Request failed with status code: "+strconv.Itoa(response.StatusCode), ad.Pack, "requestNewJWTToken")
		if ad.Logger.GetLoggingGlobalLevel() == log.DebugLevel {
			bodyBytes, _ := io.ReadAll(response.Body)
			ad.Logger.Warning("Response Body:"+string(bodyBytes), ad.Pack, "requestNewJWTToken")
		}

		return nil, errors.New("request failed with status code:" + strconv.Itoa(response.StatusCode))
	}

	return result, nil
}

// getJWKToken returns a valid Token to be used in a secure communication
// @author AB
// @params
// @return
// Error if any
func (ad *RestAPI) getJWKToken(clientID string) error {
	ad.Logger.Info("Loading JWT token", ad.Pack, "getJWKToken")

	if ad.token != nil && jwt.ValidateExpiration(ad.Logger, ad.token) {
		ad.Logger.Info("Token is valid, using previous token", ad.Pack, "getJWKToken")
		return nil
	}

	ad.Logger.Info("Token is invalid, Requesting new token", ad.Pack, "getJWKToken")

	token, err := ad.requestNewJWTToken(clientID)
	if err != nil {
		ad.Logger.Error(err, "Error sending request", ad.Pack, "getJWKToken")
		return err
	}

	ad.token = token
	return nil
}

// getHTTPClient Returns a client configured to use certificates for mTLS communication
// @author AB
// @return
// http client: Client created with certificate info
func (ad *RestAPI) getHTTPClient() *http.Client {
	httpClient := &http.Client{
		Transport: &http.Transport{
			// TLSClientConfig: &tls.Config{
			// 	Certificates:       []tls.Certificate{ad.certificates},
			// 	InsecureSkipVerify: true,
			// },
		},
	}

	return httpClient
}

// executeGet returns the response body of a GET request
func (ad *RestAPI) executeGet(url string, retryTimes int) ([]byte, error) {
	ad.Logger.Info("Executing Get Request", ad.Pack, "executeGet")
	ad.Logger.Debug("URL: "+url, ad.Pack, "executeGet")
	httpClient := ad.getHTTPClient()

	// Create a new request
	response, err := httpClient.Get(url)
	if err != nil {
		ad.Logger.Error(err, "Error executing request", ad.Pack, "executeGet")
		if retryTimes > 0 {
			ad.Logger.Info("Retrying request", ad.Pack, "executeGet")
			time.Sleep(1 * time.Second)
			return ad.executeGet(url, retryTimes-1)
		}

		return nil, err
	}

	if response.StatusCode == http.StatusForbidden {
		ad.Logger.Warning("Forbidden status code", ad.Pack, "executeGet")
		return nil, errors.New("forbidden status code")
	}

	// Check the status code of the response
	if response.StatusCode != http.StatusOK {
		ad.Logger.Warning("Unexpected status code: "+http.StatusText(response.StatusCode), ad.Pack, "executeGet")
		if retryTimes > 0 {
			ad.Logger.Info("Retrying request", ad.Pack, "executeGet")
			time.Sleep(1 * time.Second)
			return ad.executeGet(url, retryTimes-1)
		}
		return nil, errors.New("invalid status code: " + strconv.Itoa(response.StatusCode))
	}

	defer func() {
		if err := response.Body.Close(); err != nil {
			ad.Logger.Error(err, "Error closing response body", ad.Pack, "executeGet")
		}
	}()

	// defer response.Body.Close()

	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		ad.Logger.Error(err, "Error reading response body", ad.Pack, "executeGet")
		return nil, err
	}

	// Check the status code of the response
	if strings.Contains(string(body), "NoSuchKey") {
		ad.Logger.Warning("configuration file not found.", ad.Pack, "executeGet")
		return nil, errors.New("configuration file not found: " + url)
	}

	return body, nil
}
