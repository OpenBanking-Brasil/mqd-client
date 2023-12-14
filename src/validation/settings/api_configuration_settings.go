package settings

var (
// apiGroupSettings APIGroupSettings // Settings for each endpoint
)

type APISetting struct {
	API          string               `json:"api"`       // API name
	BasePath     string               `json:"base_path"` // Base path for the folder
	Version      string               `json:"version"`   // API version
	EndpointBase string               `json:"endpoint_base"`
	EndpointList []APIEndpointSetting `json:"endpoint_List"`
}

type APIEndpointSetting struct {
	Endpoint              string `json:"endpoint"`                // Name of the endpoint requested
	HeaderValidationRules string `json:"header_validation_rules"` // Header validation rules
	BodyValidationRules   string `json:"body_validation_rules"`   // Body validation rules
	JSONHeaderSchema      string `json:"header_schema"`           // Schema for the Header
	JSONBodySchema        string `json:"body_schema"`             // JSON schema for the Body
}

type APIGroupSetting struct {
	Group    string       `json:"group"`     // API Group name
	BasePath string       `json:"base_path"` // Base path for the folder
	ApiList  []APISetting `json:"api_list"`  // List of APIs
}

type APIGroupSettings struct {
	Settings []APIGroupSetting // Settings for each endpoint
}

func (s *APIGroupSettings) GetGroupSetting(groupName string) *APIGroupSetting {
	for i, setting := range s.Settings {
		if setting.Group == groupName {
			return &s.Settings[i]
		}
	}

	return nil
}

func (s *APIGroupSetting) GetAPISetting(apiName string) *APISetting {
	for i, setting := range s.ApiList {
		if setting.API == apiName {
			return &s.ApiList[i]
		}
	}

	return nil
}
