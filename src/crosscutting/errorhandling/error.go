package errorhandling

// ErrorResponse contains the information of errors insede the app
type ErrorResponse struct {
	Error            string `json:"error"`             // Error contains the name of the error
	ErrorDescription string `json:"error_description"` // Short Description of the error
	MainError        error  `json:"-"`                 // Principal error if any
}
