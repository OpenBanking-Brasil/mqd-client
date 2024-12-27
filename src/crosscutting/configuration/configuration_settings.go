package configuration

import "github.com/google/uuid"

// Settings Manages the configuration values for the application
type Settings struct {
	// ConfigurationSettings stores the settings for the current instance
	ConfigurationSettings struct {
		LoggingLevel  string `yaml:"LoggingLevel" env:"LOGGING_LEVEL, overwrite"`
		Environment   string `yaml:"Environment" env:"ENVIRONMENT, overwrite"`
		APIPort       string `yaml:"APIPort" env:"API_PORT, overwrite"`
		ApplicationID uuid.UUID
	} `yaml:"ConfigurationSettings"`

	// ApplicationSettings stores the settings for the application
	ApplicationSettings struct {
		Mode           string `yaml:"Mode" env:"APPLICATION_MODE, overwrite"`
		OrganisationID string `yaml:"OrganisationID" env:"SERVER_ORG_ID, overwrite"`
	} `yaml:"ApplicationSettings"`

	// ReportSettings stores the settings for reporting
	ReportSettings struct {
		ExecutionWindow int `yaml:"ExecutionWindow" env:"REPORT_EXECUTION_WINDOW, overwrite"`
		ExecutionNumber int `yaml:"ExecutionNumber" env:"REPORT_EXECUTION_NUMBER, overwrite"`
	} `yaml:"ReportSettings"`

	// ReportSettings stores the security settings of the application
	SecuritySettings struct {
		EnableHTTPS  bool   `yaml:"EnableHTTPS" env:"ENABLE_HTTPS, overwrite"`
		ProxyURL     string `yaml:"ProxyURL" env:"PROXY_URL, overwrite"`
		CertFilePath string
		KeyFilePath  string
	} `yaml:"SecuritySettings"`

	// ResultSettings stores the settings for result management
	ResultSettings struct {
		Enabled            bool `yaml:"Enabled" env:"RESULT_ENABLED, overwrite"`
		FilesPerDay        int  `yaml:"FilesPerDay" env:"RESULT_FILES_PER_DAY, overwrite"`
		DaysToStore        int  `yaml:"DaysToStore" env:"RESULT_DAYS_TO_STORE, overwrite"`
		SamplesPerError    int  `yaml:"SamplesPerError" env:"RESULT_SAMPLES_PER_ERROR, overwrite"`
		MaskPrivateContent bool `yaml:"MaskPrivateContent" env:"RESULT_MASK_PRIVATE_CONTENT, overwrite"`
	} `yaml:"ResultSettings"`
}
