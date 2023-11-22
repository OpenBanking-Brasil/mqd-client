package server

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

	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/configuration"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/log"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/security/jwt"
)

const CERTIFICATE_PATH = "ParameterData//certificates//"
const TOKEN_PATH = "/token"
const REPORT_PATH = "/report"

const ENV_CLIENT_CRT_FILE = "CLIENT_CRT_FILE" // Certificate file
const ENV_CLIENT_KEY_FILE = "CLIENT_KEY_FILE" // Private key file

// MQDServer Struct has the information to connect to the central server and send the Report
type MQDServer struct {
	pack         string          // Package name
	token        *jwt.JWKToken   // Token used by the server
	certificates tls.Certificate // Certificates to be used during server connection
	logger       log.Logger      // Logger to be used by the server
}

// NewMQDServer Creates a new MQDServer
// @author AB
// @params
// logger: Logger to be used
// @return
// MQDServer: Server created
func NewMQDServer(logger log.Logger) *MQDServer {
	result := &MQDServer{pack: "MQDServer", logger: logger}
	result.loadCertificates()
	return result
}

// SendReport Sends a report to the central server
// @author AB
// @param
// report: Report to be sent
// @return
// error: Error if any
func (m *MQDServer) SendReport(report Report) error {
	m.logger.Info("Sending report to central Server", m.pack, "sendReportToAPI")

	err := m.getJWKToken()
	if err != nil {
		return err
	}

	err = m.postReport(report)
	if err != nil {
		return err
	}

	return nil
}

// getHttpClient Returns a client configured to use certificates for mTLS communication
// @author AB
// @return
// http client: Client created with certificate info
func (m *MQDServer) getHttpClient() *http.Client {
	m.logger.Info("creating http Client with Certificates. ", m.pack, "getHttpClient")
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates:       []tls.Certificate{m.certificates},
				InsecureSkipVerify: true,
			},
		},
	}

	return httpClient
}

// loadCertificates Loads certificates from environment variables
// @author AB
// @params
// @return
// error: Error if any
// Response from server in case of success
func (m *MQDServer) requestNewToken() (*http.Response, error) {
	m.logger.Info("Requesting new token", m.pack, "requestNewToken")

	// Create the request body with grant_type and client_id
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", configuration.ClientID)
	requestBody := strings.NewReader(data.Encode())

	m.logger.Debug("ServerURL:"+configuration.ServerURL+TOKEN_PATH, m.pack, "requestNewToken")
	m.logger.Debug("Body:"+data.Encode(), m.pack, "requestNewToken")

	request, err := http.NewRequest("POST", configuration.ServerURL+TOKEN_PATH, requestBody)
	if err != nil {
		m.logger.Error(err, "Error creating request", m.pack, "requestNewToken")
		return nil, err
	}

	// Set the Content-Type header to application/x-www-form-urlencoded
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	httpClient := m.getHttpClient()

	// Send the request
	resp, err := httpClient.Do(request)
	if err != nil {
		m.logger.Error(err, "Error sending request", m.pack, "requestNewToken")
		return nil, err
	}

	return resp, nil
}

// getJWKToken returns a valid Token to be used in a secure communication
// @author AB
// @params
// @return
// Error if any
func (m *MQDServer) getJWKToken() error {
	m.logger.Info("Loading JWT token", m.pack, "getJWKToken")

	if m.token != nil && jwt.ValidateExpiration(m.logger, m.token) {
		m.logger.Info("Token is valid, using previous token", m.pack, "getJWKToken")
		return nil
	}

	m.logger.Info("Token is invalid, Requesting new token", m.pack, "getJWKToken")

	response, err := m.requestNewToken()
	if err != nil {
		m.logger.Error(err, "Error sending request", m.pack, "getJWKToken")
		return err
	}

	defer response.Body.Close()

	// Check the response
	if response.StatusCode == http.StatusOK {
		m.token, err = jwt.GetTokenFromReader(m.logger, response.Body)
		return err
	} else {
		m.logger.Warning("Request failed with status code: "+strconv.Itoa(response.StatusCode), m.pack, "getJWKToken")
		if m.logger.GetLoggingGlobalLevel() == log.DebugLevel {
			body, _ := io.ReadAll(response.Body)
			m.logger.Warning("Response Body:"+string(body), m.pack, "getJWKToken")
		}

		return errors.New("request failed with status code:" + strconv.Itoa(response.StatusCode))
	}
}

// postReport sends the report to the server using required authorization
// @author AB
// @params
// report: Report to send
// @return
// error: will be != of nil in case of error
func (m *MQDServer) postReport(report Report) error {
	m.logger.Info("Posting report", m.pack, "postReport")

	httpClient := m.getHttpClient()

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
	req.Header.Set("Authorization", "Bearer "+m.token.AccessToken)

	// Send the request
	resp, err := httpClient.Do(req)
	if err != nil {
		m.logger.Error(err, "Error sending report.", m.pack, "sendReportToAPI")
		return err
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		m.logger.Warning("Error sending report, Status code: "+fmt.Sprint(resp.StatusCode), "Result", "sendReportToAPI")
	} else {
		// Read the body of the message
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		m.logger.Info(string(body), "Result", "sendReportToAPI")
	}

	return nil
}

// loadCertificates Load certificate files (crt and key) into memmory for further use
// @author AB
// @params
// @return
func (m *MQDServer) loadCertificates() {
	m.logger.Info("Loading certificates", m.pack, "loadCertificates")

	certFile := crosscutting.GetEnvironmentValue(m.logger, ENV_CLIENT_CRT_FILE, "client.crt")
	keyFile := crosscutting.GetEnvironmentValue(m.logger, ENV_CLIENT_KEY_FILE, "client.key")

	keyPath := fmt.Sprintf("%s%s", CERTIFICATE_PATH, keyFile)
	crtPath := fmt.Sprintf("%s%s", CERTIFICATE_PATH, certFile)

	// Load client certificate and private key
	cert, err := tls.LoadX509KeyPair(crtPath, keyPath)
	if err != nil {
		m.logger.Fatal(err, "Error loading client certificate and key", "Server", "loadCertificates")
		return
	}

	m.certificates = cert
}
