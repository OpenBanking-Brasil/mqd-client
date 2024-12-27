package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/configuration"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/log"
	"github.com/OpenBanking-Brasil/MQD_Client/domain/models"
)

const (
	tokenPath    = "/token"
	reportPath   = "/report"
	settingsPath = "/settings"

	configurationSettingsFile = "configurationSettings.json"
)

// ReportServerMQD Struct has the information to connect to the central server and send the Report
type ReportServerMQD struct {
	RestAPI
	settings configuration.Settings
}

// NewReportServerMQD Creates a new MQDServer
//
// Parameters:
//   - logger: Logger to be used
//
// Returns:
//   - ReportServerMQD: Server created
func NewReportServerMQD(logger log.Logger, serverURL string, settings configuration.Settings) *ReportServerMQD {
	result := &ReportServerMQD{
		RestAPI: RestAPI{
			OFBStruct: crosscutting.OFBStruct{
				Pack:   "services.ReportServerMQD",
				Logger: logger,
			},
			serverURL: serverURL,
		},
		settings: settings,
	}

	// result.loadCertificates()
	return result
}

// SendReport Sends a report to the central server
//
// Parameters:
//   - report: Report to be sent
//
// Returns:
//   - error: Error if any
func (rs *ReportServerMQD) SendReport(report models.Report) error {
	rs.Logger.Info("Sending report to central Server", rs.Pack, "sendReportToAPI")

	err := rs.getJWKToken(rs.settings.ApplicationSettings.OrganisationID)
	if err != nil {
		return err
	}

	err = rs.postReport(report)
	if err != nil {
		return err
	}

	return nil
}

// postReport sends the report to the server using required authorization
//
// Parameters:
//   - report: Report to be sent
//
// Returns:
//   - error: Error if any
func (rs *ReportServerMQD) postReport(report models.Report) error {
	rs.Logger.Info("Posting report", rs.Pack, "postReport")

	httpClient := rs.getHTTPClient()

	requestBody, err := json.Marshal(report)
	if err != nil {
		return err
	}

	// Create a new request
	req, err := http.NewRequest("POST", rs.serverURL+reportPath, bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return err
	}

	// Set the Authorization header with your token
	req.Header.Set("Authorization", "Bearer "+rs.token.AccessToken)

	// Send the request
	resp, err := httpClient.Do(req)
	if err != nil {
		rs.Logger.Error(err, "Error sending report.", rs.Pack, "postReport")
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			rs.Logger.Error(err, "Error closing resp.Body", rs.Pack, "postReport")
		}
	}(resp.Body)

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		rs.Logger.Warning("Error sending report, Status code: "+fmt.Sprint(resp.StatusCode), rs.Pack, "postReport")
	} else {
		// Read the body of the message
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		rs.Logger.Info(string(body), rs.Pack, "postReport")
	}

	return nil
}

// LoadAPIConfigurationFile Loads a json configuration file from the server
//
// Parameters:
//   - filePath: Path for the file on the server
//
// Returns:
//   - []byte: Byte array with the info
//   - error: Error if any
func (rs *ReportServerMQD) LoadAPIConfigurationFile(filePath string) ([]byte, error) {
	rs.Logger.Info("Loading API configuration", rs.Pack, "loadAPIConfiguration")
	serverPath := rs.serverURL + settingsPath + "/" + filePath
	return rs.executeGet(serverPath, 3)
}

// LoadConfigurationSettings Loads the main configuration file for the application
//
// Parameters:
//
// Returns:
//   - ConfigurationSettings: configuration file found on the server
//   - error: Error if any
func (rs *ReportServerMQD) LoadConfigurationSettings() (*models.ConfigurationSettings, error) {
	rs.Logger.Info("Loading ConfigurationSettings", rs.Pack, "LoadConfigurationSettings")
	serverPath := rs.serverURL + settingsPath + "/" + configurationSettingsFile

	body, err := rs.executeGet(serverPath, 3)
	if err != nil {
		return nil, err
	}

	var result models.ConfigurationSettings
	err = json.Unmarshal(body, &result)
	if err != nil {
		rs.Logger.Error(err, "error unmarshal file", rs.Pack, "loadAPIConfiguration")
		return nil, err
	}

	return &result, nil
}
