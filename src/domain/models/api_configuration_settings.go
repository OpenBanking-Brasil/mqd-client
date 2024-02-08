package models

const (
	// ExtremelyHighTroughput defines the extremely high troughtput keyword
	ExtremelyHighTroughput = "EXTEMELY_HIGH"
	// HighTroughput defines High troughtput keyword
	HighTroughput = "HIGH"
	// MediumTroughput defines the Medium troughtput keyword
	MediumTroughput = "MEDIUM"
	// LowTroughput defines the Low troughtput keyword
	LowTroughput = "LOW"
	// VeryLowTroughput defines the Very Low Troughput keyword
	VeryLowTroughput = "VERY_LOW"
)

// APISetting Contains the settings needed to perform validations on API / endpoints
type APISetting struct {
	API          string               `json:"api"`           // API name
	BasePath     string               `json:"base_path"`     // Base path for the folder
	Version      string               `json:"version"`       // API version
	EndpointBase string               `json:"endpoint_base"` // Base URL of this endpoint
	EndpointList []APIEndpointSetting `json:"endpoint_List"` // List of settings for this endpoint
}

// APIEndpointSetting has the specific validation settings for an endpoint
type APIEndpointSetting struct {
	Endpoint              string `json:"endpoint"`                // Name of the endpoint requested
	HeaderValidationRules string `json:"header_validation_rules"` // Header validation rules
	BodyValidationRules   string `json:"body_validation_rules"`   // Body validation rules
	JSONHeaderSchema      string `json:"header_schema"`           // Schema for the Header
	JSONBodySchema        string `json:"body_schema"`             // JSON schema for the Body
	Throughput            string `json:"throughput"`              // Relation of the amount of requests for this endpoint
}

// APIGroupSetting Validation sattings for an API group
type APIGroupSetting struct {
	Group    string       `json:"group"`     // API Group name
	BasePath string       `json:"base_path"` // Base path for the folder
	APIList  []APISetting `json:"api_list"`  // List of APIs
}

// ValidationSettings stores the configuration for validations of the application
type ValidationSettings struct {
	APIGroupSettings                     []APIGroupSetting `json:"APIGroupSettings"`                     // API group validation settings
	TransmitterValidationRate            int               `json:"TransmitterValidationRate"`            // Validation rate in % for transmitter mode 1 - 100
	ReceiverValidationRate               int               `json:"ReceiverValidationRate"`               // Validation rate in % for receiver mode 1 - 100
	ExtremelyHighTroughputValidationRate int               `json:"ExtremelyHighTroughputValidationRate"` // Validation rate in % for extremely high throughput mode 1 - 100
	HighTroughputValidationRate          int               `json:"HighTroughputValidationRate"`          // Validation rate in % for high throughput mode 1 - 100
	MediumTroughputValidationRate        int               `json:"MediumTroughputValidationRate"`        // Validation rate in % for medium throughput mode 1 - 100
	LowTroughputValidationRate           int               `json:"LowTroughputValidationRate"`           // Validation rate in % for low throughput mode 1 - 100
	VeryLowTroughputValidationRate       int               `json:"VeryLowTroughputValidationRate"`       // Validation rate in % for very low throughput mode 1 - 100
}

// GetGroupSetting returns a group settings based on the group name
//
// Parameters:
//   - groupName: Group name to find the settings
//
// Returns:
//   - *APIGroupSetting: Group settings found, nil if it is not supported
func (vs *ValidationSettings) GetGroupSetting(groupName string) *APIGroupSetting {
	for i, setting := range vs.APIGroupSettings {
		if setting.Group == groupName {
			return &vs.APIGroupSettings[i]
		}
	}

	return nil
}

// GetAPISetting returns an API settings based on the name
//
// Parameters:
//   - apiName: API name to find the settings
//
// Returns:
//   - *APISetting: API settings found, nil if it is not supported
func (vs *APIGroupSetting) GetAPISetting(apiName string) *APISetting {
	for i, setting := range vs.APIList {
		if setting.API == apiName {
			return &vs.APIList[i]
		}
	}

	return nil
}
