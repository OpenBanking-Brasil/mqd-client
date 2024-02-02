package services

import (
	"bytes"
	"crypto/tls"
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

type AuthenticationType int

const (
	NoToken     AuthenticationType = iota // Indicates no authentication needed
	ApiToken                              // Indicates authentication using Api Token
	BearerToken                           // Indicates authentication using Bearer Token
)

// API_DAO is the struct to handle connections to APIs
type API_DAO struct {
	crosscutting.OFBStruct               // Base structure
	client                 *http.Client  //Client to be used to access the API
	apiToken               string        // secure token (if needed)
	token                  *jwt.JWKToken // Token used by the server
	needsCertificates      bool          // Indicates if certificates are needed for mTLS
	// certificates           tls.Certificate // Certificates to be used during server connection
	clientID string
}

// executeRequest executes a request to the API
// @author AB
// @param
// url: url of the API
// @return
// Byte array with the result id success
// error if any
func (this *API_DAO) executeRequest(url string) ([]byte, *errorhandling.ErrorResponse) {
	this.Logger.Debug("Executing request: "+url, this.Pack, "executeRequest")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		this.Logger.Error(err, "There was an error creating the request", this.Pack, "executeRequest")
		return nil, &errorhandling.ErrorResponse{
			Error:            "Error creating the request",
			ErrorDescription: "There was an error creating the request",
			MainError:        err,
		}
	}

	// if this.needsAuthorization {
	// 	req.Header.Set("Authorization", "Api-Token "+this.apiToken)
	// }

	resp, err := this.client.Do(req)
	if err != nil {
		this.Logger.Error(err, "There was an error executing the request", this.Pack, "executeRequest")
		return nil, &errorhandling.ErrorResponse{
			Error:            "Error executing the request",
			ErrorDescription: "There was an error while executing the request",
			MainError:        err,
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		this.Logger.Error(err, "There was an error while reading the response body", this.Pack, "executeRequest")
		return nil, &errorhandling.ErrorResponse{
			Error:            "Error reading the request",
			ErrorDescription: "There was an error reading the request body",
			MainError:        err,
		}
	}

	if resp.StatusCode != http.StatusOK {
		this.Logger.Warning("Status code not expected: "+http.StatusText(resp.StatusCode), this.Pack, "executeRequest")
		this.Logger.Debug("Body: "+string(body), this.Pack, "executeRequest")
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
func (this *API_DAO) executeMethod(method string, url string, body []byte) *errorhandling.ErrorResponse {
	this.Logger.Debug("Id URL: "+url, this.Pack, "executeMethod")
	this.Logger.Debug("monitor Body: "+string(body), this.Pack, "executeMethod")

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		this.Logger.Error(err, "Error creating server request", this.Pack, "executeMethod")
		return &errorhandling.ErrorResponse{
			MainError: err,
		}
	}

	req.Header.Set("Content-Type", "application/json")
	// if this.needsAuthorization {
	// 	req.Header.Set("Authorization", "Api-Token "+this.apiToken)
	// }

	resp, err := this.client.Do(req)
	if err != nil {
		this.Logger.Error(err, "Error executing server request", this.Pack, "executeMethod")
		return &errorhandling.ErrorResponse{
			MainError: err,
		}
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		this.Logger.Error(err, "Error Reading server Response Body", this.Pack, "executeMethod")
		return &errorhandling.ErrorResponse{
			Error:            "Error reading the request",
			ErrorDescription: "There was an error reading the request body",
			MainError:        err,
		}
	}

	if resp.StatusCode >= 300 {
		this.Logger.Warning("Status code not expected: "+strconv.Itoa(resp.StatusCode), this.Pack, "executeMethod")
		this.Logger.Warning("Request Body: "+string(body), this.Pack, "executeMethod")
		this.Logger.Warning("Response Body: "+string(responseBody), this.Pack, "executeMethod")
		this.Logger.Panic("Status not expected.", this.Pack, "executeMethod")
		return &errorhandling.ErrorResponse{
			MainError: fmt.Errorf("unexpected status code: %d", resp.StatusCode),
		}
	}

	return nil
}

func (this *API_DAO) getRequest(authenticationType AuthenticationType) (*resty.Request, error) {
	// Create a Resty Client
	client := resty.New()
	request := client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}).R()
	switch authenticationType {
	case NoToken:
		return request, nil
	case ApiToken:
		request = request.SetAuthScheme("Api-Token").SetAuthToken(this.apiToken)
	case BearerToken:
		// err := this.getJWKToken()
		// if err != nil {
		// 	this.Logger.Error(err, "Error getting JWK Token", this.Pack, "getRequest")
		// 	return request, err
		// }

		// request = request.SetAuthToken(this.token.AccessToken)
	}

	return request, nil
}

// getFromAPI executes a request to the API and returns the indicated interface
// @author AB
// @param
// method: method of the request
// url: url of the API
// result: object to be mapped with the response
// @return
// error if any
func (this *API_DAO) getFromAPI(method string, url string, authenticationType AuthenticationType, result interface{}) *errorhandling.ErrorResponse {
	request, err := this.getRequest(authenticationType)
	if err != nil {
		this.Logger.Error(err, "Error getting JWK Token", this.Pack, "getRequest")
		return &errorhandling.ErrorResponse{
			Error:            "Error loading the request",
			ErrorDescription: "There was an error while loading the request",
			MainError:        err,
		}
	}

	resp, err := request.
		SetResult(&result).
		Execute(method, url)
	if err != nil {
		this.Logger.Error(err, "There was an error executing the request: "+url, this.Pack, "getFromAPI")
		return &errorhandling.ErrorResponse{
			Error:            "Error executing the request",
			ErrorDescription: "There was an error while executing the request",
			MainError:        err,
		}
	}

	if resp.StatusCode() >= 300 {
		this.Logger.Warning("Status code not expected: "+strconv.Itoa(resp.StatusCode()), this.Pack, "getFromAPI")
		this.Logger.Warning("For url: "+url, this.Pack, "getFromAPI")
		this.Logger.Warning("Body: "+string(resp.Body()), this.Pack, "getFromAPI")
		return &errorhandling.ErrorResponse{
			Error:            "There was an error during the request",
			ErrorDescription: "Unexpected status code :" + http.StatusText(resp.StatusCode()),
			MainError:        errors.New("unexpected status code: " + strconv.Itoa(resp.StatusCode())),
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
func (this *API_DAO) requestNewJWTToken() (*jwt.JWKToken, error) {
	this.Logger.Info("Requesting new token", this.Pack, "requestNewJWTToken")

	// Create a Resty client
	client := resty.New()

	// Define the parameters for the token request
	params := url.Values{}
	params.Set("grant_type", "client_credentials")
	params.Set("client_id", configuration.ClientID)
	requestBody := params.Encode()

	this.Logger.Debug("ServerURL:"+configuration.ServerURL+tokenPath, this.Pack, "requestNewJWTToken")
	this.Logger.Debug("Body:"+requestBody, this.Pack, "requestNewJWTToken")

	// Send the token request
	response, err := client.R().
		SetBody(requestBody). // Encode the parameters as application/x-www-form-urlencoded
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		Post(configuration.ServerURL + tokenPath)
	if err != nil {
		this.Logger.Error(err, "Error creating request", this.Pack, "requestNewJWTToken")
		return nil, err
	}

	var result *jwt.JWKToken
	if response.StatusCode() == http.StatusOK {
		result, _ = jwt.GetTokenFromBinary(this.Logger, response.Body())
	} else {
		this.Logger.Warning("Request failed with status code: "+strconv.Itoa(response.StatusCode()), this.Pack, "requestNewJWTToken")
		if this.Logger.GetLoggingGlobalLevel() == log.DebugLevel {
			this.Logger.Warning("Response Body:"+string(response.Body()), this.Pack, "requestNewJWTToken")
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
func (this *API_DAO) getJWKToken() error {
	this.Logger.Info("Loading JWT token", this.Pack, "getJWKToken")

	if this.token != nil && jwt.ValidateExpiration(this.Logger, this.token) {
		this.Logger.Info("Token is valid, using previous token", this.Pack, "getJWKToken")
		return nil
	}

	this.Logger.Info("Token is invalid, Requesting new token", this.Pack, "getJWKToken")

	token, err := this.requestNewJWTToken()
	if err != nil {
		this.Logger.Error(err, "Error sending request", this.Pack, "getJWKToken")
		return err
	}

	this.token = token
	return nil
}

// getHttpClient Returns a client configured to use certificates for mTLS communication
// @author AB
// @return
// http client: Client created with certificate info
func (this *API_DAO) getHttpClient() *http.Client {
	httpClient := &http.Client{
		Transport: &http.Transport{
			// TLSClientConfig: &tls.Config{
			// 	Certificates:       []tls.Certificate{this.certificates},
			// 	InsecureSkipVerify: true,
			// },
		},
	}

	return httpClient
}

// executeGet returns the response body of a GET request
func (this *API_DAO) executeGet(url string, reetryTimes int) ([]byte, error) {
	this.Logger.Info("Executing Get Request", this.Pack, "executeGet")
	this.Logger.Debug("URL: "+url, this.Pack, "executeGet")
	httpClient := this.getHttpClient()

	// Create a new request
	response, err := httpClient.Get(url)
	if err != nil {
		this.Logger.Error(err, "Error executing request", this.Pack, "executeGet")
		if reetryTimes > 0 {
			this.Logger.Info("Retrying request", this.Pack, "executeGet")
			time.Sleep(1 * time.Second)
			return this.executeGet(url, reetryTimes-1)
		}

		return nil, err
	}

	if response.StatusCode == http.StatusForbidden {
		this.Logger.Warning("Forbidden status code", this.Pack, "executeGet")
		return nil, errors.New("Forbidden status code")
	}

	// Check the status code of the response
	if response.StatusCode != http.StatusOK {
		this.Logger.Warning("Unexpected status code: "+http.StatusText(response.StatusCode), this.Pack, "executeGet")
		if reetryTimes > 0 {
			this.Logger.Info("Retrying request", this.Pack, "executeGet")
			time.Sleep(1 * time.Second)
			return this.executeGet(url, reetryTimes-1)
		}
		return nil, errors.New("invalid status code: " + strconv.Itoa(response.StatusCode))
	}

	defer response.Body.Close()

	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		this.Logger.Error(err, "Error reading response body", this.Pack, "executeGet")
		return nil, err
	}

	// Check the status code of the response
	if strings.Contains(string(body), "NoSuchKey") {
		this.Logger.Warning("configuration file not found.", this.Pack, "executeGet")
		return nil, errors.New("configuration file not found: " + url)
	}

	return body, nil
}
