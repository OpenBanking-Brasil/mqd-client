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
	API_DAO
}

// NewMQDServer Creates a new MQDServer
// @author AB
// @params
// logger: Logger to be used
// @return
// MQDServer: Server created
func NewReportServerMQD(logger log.Logger) *ReportServerMQD {
	result := &ReportServerMQD{
		API_DAO: API_DAO{
			OFBStruct: crosscutting.OFBStruct{
				Pack:   "services.ReportServerMQD",
				Logger: logger,
			},
		},
	}

	// result.loadCertificates()
	return result
}

// SendReport Sends a report to the central server
// @author AB
// @param
// report: Report to be sent
// @return
// error: Error if any
func (this *ReportServerMQD) SendReport(report models.Report) error {
	this.Logger.Info("Sending report to central Server", this.Pack, "sendReportToAPI")

	err := this.getJWKToken()
	if err != nil {
		return err
	}

	err = this.postReport(report)
	if err != nil {
		return err
	}

	return nil
}

// postReport sends the report to the server using required authorization
// @author AB
// @params
// report: Report to send
// @return
// error: will be != of nil in case of error
func (this *ReportServerMQD) postReport(report models.Report) error {
	this.Logger.Info("Posting report", this.Pack, "postReport")

	httpClient := this.getHttpClient()

	requestBody, err := json.Marshal(report)
	if err != nil {
		return err
	}

	// Create a new request
	req, err := http.NewRequest("POST", configuration.ServerURL+reportPath, bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return err
	}

	// Set the Authorization header with your token
	req.Header.Set("Authorization", "Bearer "+this.token.AccessToken)

	// Send the request
	resp, err := httpClient.Do(req)
	if err != nil {
		this.Logger.Error(err, "Error sending report.", this.Pack, "postReport")
		return err
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		this.Logger.Warning("Error sending report, Status code: "+fmt.Sprint(resp.StatusCode), this.Pack, "postReport")
	} else {
		// Read the body of the message
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		this.Logger.Info(string(body), this.Pack, "postReport")
	}

	return nil
}

func (this *ReportServerMQD) LoadAPIConfigurationFile(filePath string) ([]byte, error) {
	this.Logger.Info("Loading API configuration", this.Pack, "loadAPIConfiguration")
	serverPath := configuration.ServerURL + settingsPath + "/" + filePath
	return this.executeGet(serverPath, 3)
}

func (this *ReportServerMQD) LoadConfigurationSettings() (*models.ConfigurationSettings, error) {
	this.Logger.Info("Loading ConfigurationSettings", this.Pack, "LoadConfigurationSettings")
	serverPath := configuration.ServerURL + settingsPath + "/" + configurationSettingsFile

	body, err := this.executeGet(serverPath, 3)
	if err != nil {
		return nil, err
	}

	var result models.ConfigurationSettings
	err = json.Unmarshal(body, &result)
	if err != nil {
		this.Logger.Error(err, "error unmarshal file", this.Pack, "loadAPIConfiguration")
		return nil, err
	}

	return &result, nil
}
