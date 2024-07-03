package services

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/configuration"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/errorhandling"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/log"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/security/jwt"
	"github.com/go-resty/resty/v2"
)

//
// type AuthenticationType int

// const (
// 	NoToken     AuthenticationType = iota // Indicates no authentication needed
// 	ApiToken                              // Indicates authentication using Api Token
// 	BearerToken                           // Indicates authentication using Bearer Token
// )

// RestAPI is the struct to handle connections to APIs
type RestAPI struct {
	crosscutting.OFBStruct               // Base structure
	client                 *http.Client  // Client to be used to access the API
	apiToken               string        // secure token (if needed)
	token                  *jwt.JWKToken // Token used by the server
	needsCertificates      bool          // Indicates if certificates are needed for mTLS
	clientID               string
}

// executeRequest executes a request to the API
// @author AB
// @param
// url: url of the API
// @return
// Byte array with the result id success
// error if any
func (ad *RestAPI) executeRequest(url string) ([]byte, *errorhandling.ErrorResponse) {
	ad.Logger.Debug("Executing request: "+url, ad.Pack, "executeRequest")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		ad.Logger.Error(err, "There was an error creating the request", ad.Pack, "executeRequest")
		return nil, &errorhandling.ErrorResponse{
			Error:            "Error creating the request",
			ErrorDescription: "There was an error creating the request",
			MainError:        err,
		}
	}

	// if ad.needsAuthorization {
	// 	req.Header.Set("Authorization", "Api-Token "+ad.apiToken)
	// }

	resp, err := ad.client.Do(req)
	if err != nil {
		ad.Logger.Error(err, "There was an error executing the request", ad.Pack, "executeRequest")
		return nil, &errorhandling.ErrorResponse{
			Error:            "Error executing the request",
			ErrorDescription: "There was an error while executing the request",
			MainError:        err,
		}
	}
	// defer resp.Body.Close()
	defer func() {
		if err := resp.Body.Close(); err != nil {
			ad.Logger.Error(err, "Error closing response body", ad.Pack, "executeGet")
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ad.Logger.Error(err, "There was an error while reading the response body", ad.Pack, "executeRequest")
		return nil, &errorhandling.ErrorResponse{
			Error:            "Error reading the request",
			ErrorDescription: "There was an error reading the request body",
			MainError:        err,
		}
	}

	if resp.StatusCode != http.StatusOK {
		ad.Logger.Warning("Status code not expected: "+http.StatusText(resp.StatusCode), ad.Pack, "executeRequest")
		ad.Logger.Debug("Body: "+string(body), ad.Pack, "executeRequest")
		return nil, &errorhandling.ErrorResponse{
			Error:            "There was an error during the request",
			ErrorDescription: "Unexpected status code :" + http.StatusText(resp.StatusCode),
			MainError:        fmt.Errorf("unexpected status code: %d", resp.StatusCode),
		}
	}

	return body, nil
}

// executeRequestWithBody executes a request to the API with a body
// @author AB
// @param
// method: method of the request
// url: url of the API
// body: body of the request
// @return
// error if any
func (ad *RestAPI) executeMethod(method string, url string, body []byte) *errorhandling.ErrorResponse {
	ad.Logger.Debug("Id URL: "+url, ad.Pack, "executeMethod")
	ad.Logger.Debug("monitor Body: "+string(body), ad.Pack, "executeMethod")

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		ad.Logger.Error(err, "Error creating server request", ad.Pack, "executeMethod")
		return &errorhandling.ErrorResponse{
			MainError: err,
		}
	}

	req.Header.Set("Content-Type", "application/json")
	// if ad.needsAuthorization {
	// 	req.Header.Set("Authorization", "Api-Token "+ad.apiToken)
	// }

	resp, err := ad.client.Do(req)
	if err != nil {
		ad.Logger.Error(err, "Error executing server request", ad.Pack, "executeMethod")
		return &errorhandling.ErrorResponse{
			MainError: err,
		}
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			ad.Logger.Error(err, "Error closing response body", ad.Pack, "executeGet")
		}
	}()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		ad.Logger.Error(err, "Error Reading server Response Body", ad.Pack, "executeMethod")
		return &errorhandling.ErrorResponse{
			Error:            "Error reading the request",
			ErrorDescription: "There was an error reading the request body",
			MainError:        err,
		}
	}

	if resp.StatusCode >= 300 {
		ad.Logger.Warning("Status code not expected: "+strconv.Itoa(resp.StatusCode), ad.Pack, "executeMethod")
		ad.Logger.Warning("Request Body: "+string(body), ad.Pack, "executeMethod")
		ad.Logger.Warning("Response Body: "+string(responseBody), ad.Pack, "executeMethod")
		ad.Logger.Panic("Status not expected.", ad.Pack, "executeMethod")
		return &errorhandling.ErrorResponse{
			MainError: fmt.Errorf("unexpected status code: %d", resp.StatusCode),
		}
	}

	return nil
}

// loadCertificates Loads certificates from environment variables
// @author AB
// @params
// @return
// error: Error if any
// Response from server in case of success
func (ad *RestAPI) requestNewJWTToken() (*jwt.JWKToken, error) {
	ad.Logger.Info("Requesting new token", ad.Pack, "requestNewJWTToken")

	// Create a Resty client
	client := resty.New()

	// Define the parameters for the token request
	params := url.Values{}
	params.Set("grant_type", "client_credentials")
	params.Set("client_id", configuration.ClientID)
	requestBody := params.Encode()

	ad.Logger.Debug("ServerURL:"+configuration.ServerURL+tokenPath, ad.Pack, "requestNewJWTToken")
	ad.Logger.Debug("Body:"+requestBody, ad.Pack, "requestNewJWTToken")

	// Send the token request
	response, err := client.R().
		SetBody(requestBody). // Encode the parameters as application/x-www-form-urlencoded
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		Post(configuration.ServerURL + tokenPath)
	if err != nil {
		ad.Logger.Error(err, "Error creating request", ad.Pack, "requestNewJWTToken")
		return nil, err
	}

	var result *jwt.JWKToken
	if response.StatusCode() == http.StatusOK {
		result, _ = jwt.GetTokenFromBinary(ad.Logger, response.Body())
	} else {
		ad.Logger.Warning("Request failed with status code: "+strconv.Itoa(response.StatusCode()), ad.Pack, "requestNewJWTToken")
		if ad.Logger.GetLoggingGlobalLevel() == log.DebugLevel {
			ad.Logger.Warning("Response Body:"+string(response.Body()), ad.Pack, "requestNewJWTToken")
		}

		return nil, errors.New("request failed with status code:" + strconv.Itoa(response.StatusCode()))
	}

	return result, nil
}

// getJWKToken returns a valid Token to be used in a secure communication
// @author AB
// @params
// @return
// Error if any
func (ad *RestAPI) getJWKToken() error {
	ad.Logger.Info("Loading JWT token", ad.Pack, "getJWKToken")

	if ad.token != nil && jwt.ValidateExpiration(ad.Logger, ad.token) {
		ad.Logger.Info("Token is valid, using previous token", ad.Pack, "getJWKToken")
		return nil
	}

	ad.Logger.Info("Token is invalid, Requesting new token", ad.Pack, "getJWKToken")

	token, err := ad.requestNewJWTToken()
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
func (ad *RestAPI) executeGet(url string, reetryTimes int) ([]byte, error) {
	ad.Logger.Info("Executing Get Request", ad.Pack, "executeGet")
	ad.Logger.Debug("URL: "+url, ad.Pack, "executeGet")
	httpClient := ad.getHTTPClient()

	// Create a new request
	response, err := httpClient.Get(url)
	if err != nil {
		ad.Logger.Error(err, "Error executing request", ad.Pack, "executeGet")
		if reetryTimes > 0 {
			ad.Logger.Info("Retrying request", ad.Pack, "executeGet")
			time.Sleep(1 * time.Second)
			return ad.executeGet(url, reetryTimes-1)
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
		if reetryTimes > 0 {
			ad.Logger.Info("Retrying request", ad.Pack, "executeGet")
			time.Sleep(1 * time.Second)
			return ad.executeGet(url, reetryTimes-1)
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
