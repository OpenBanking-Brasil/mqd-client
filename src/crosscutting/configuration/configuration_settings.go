package configuration

import (
	"strings"

	"github.com/google/uuid"
)

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
		EnableHTTPS            bool   `yaml:"EnableHTTPS" env:"ENABLE_HTTPS, overwrite"`
		ProxyURL               string `yaml:"ProxyURL" env:"PROXY_URL, overwrite"`
		CertFilePath           string
		KeyFilePath            string
		SecurityResponseHeader SecurityResponseHeader `yaml:"SecurityResponseHeader"`
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

// SecurityResponseHeader represents the security-related HTTP headers that are included in the server's responses.
// These headers are used to enhance the security of the application by controlling browser behavior and mitigating
// potential attacks such as clickjacking, cross-site scripting (XSS), and other vulnerabilities.
//
// Fields:
//   - StrictTransportSecurity: Configures the "Strict-Transport-Security" header, which enforces HTTPS connections
//     by specifying a maximum age for which the browser should remember to only use HTTPS.
//   - XFrameOptions: Configures the "X-Frame-Options" header, which controls whether the page can be displayed
//     in a <frame>, <iframe>, or <object> element to prevent clickjacking attacks.
//   - ContentSecurityPolicy: Configures the "Content-Security-Policy" header, which defines policies for
//     controlling the sources of content that the browser is allowed to load (e.g., scripts, styles, images, etc.).
type SecurityResponseHeader struct {
	StrictTransportSecurity struct {
		MaxAge int `env:"STRICT_TRANSPORT_SECURITY_MAX_AGE, overwrite"`
	}
	XFrameOptions         XFrameOptions         `yaml:"XFrameOptions" env:"X_FRAME_OPTIONS, overwrite"`
	ContentSecurityPolicy ContentSecurityPolicy `yaml:"ContentSecurityPolicy"`
}

// XFrameOptions represents the possible values for the X-Frame-Options HTTP header.
// Possible values:
//   - DENY: Prevents the page from being displayed in any frame, regardless of the origin.
//   - SAME_ORIGIN: Allows the page to be displayed in a frame only if the frame is hosted on the same origin as the page.
type XFrameOptions string

const (
	XFrameDeny       XFrameOptions = "DENY"
	XFrameSameOrigin XFrameOptions = "SAMEORIGIN"
)

var xFrameOptionsEnum = map[XFrameOptions]struct{}{
	XFrameDeny:       {},
	XFrameSameOrigin: {},
}

func (xfo XFrameOptions) Get() string {
	if _, ok := xFrameOptionsEnum[xfo]; ok {
		return string(xfo)
	}
	return ""
}

// ContentSecurityPolicy represents the fields of the Content-Security-Policy header
type ContentSecurityPolicy struct {
	// DefaultSrc specifies the default policy for fetching resources
	DefaultSrc []string `yaml:"DefaultSrc" env:"CONTENT_SECURITY_POLICY_DEFAULT_SRC, overwrite"`

	// ScriptSrc specifies valid sources for JavaScript and WebAssembly resources
	ScriptSrc []string `yaml:"ScriptSrc" env:"CONTENT_SECURITY_POLICY_SCRIPT_SRC, overwrite"`

	// StyleSrc specifies valid sources for stylesheets
	StyleSrc []string `yaml:"StyleSrc" env:"CONTENT_SECURITY_POLICY_STYLE_SRC, overwrite"`

	// ImgSrc specifies valid sources for images
	ImgSrc []string `yaml:"ContentSecurityPolicy" env:"CONTENT_SECURITY_POLICY_IMG_SRC, overwrite"`

	// ConnectSrc specifies valid sources for XMLHttpRequest, WebSocket, etc.
	ConnectSrc []string `yaml:"ContentSecurityPolicy" env:"CONTENT_SECURITY_POLICY_CONNECT_SRC, overwrite"`

	// FontSrc specifies valid sources for fonts
	FontSrc []string `yaml:"ContentSecurityPolicy" env:"CONTENT_SECURITY_POLICY_FONT_SRC, overwrite"`

	// ObjectSrc specifies valid sources for <object>, <embed>, or <applet>
	ObjectSrc []string `yaml:"ContentSecurityPolicy" env:"CONTENT_SECURITY_POLICY_OBJECT_SRC, overwrite"`

	// MediaSrc specifies valid sources for audio and video
	MediaSrc []string `yaml:"ContentSecurityPolicy" env:"CONTENT_SECURITY_POLICY_MEDIA_SRC, overwrite"`

	// FrameSrc specifies valid sources for nested browsing contexts (e.g., <iframe>)
	FrameSrc []string `yaml:"ContentSecurityPolicy" env:"CONTENT_SECURITY_POLICY_FRAME_SRC, overwrite"`

	// FormAction specifies valid sources that can be used as an HTML <form> action
	FormAction []string `yaml:"ContentSecurityPolicy" env:"CONTENT_SECURITY_POLICY_FORM_ACTION, overwrite"`

	// PluginTypes specifies valid MIME types for plugins invoked via <object> and <embed>
	PluginTypes []string `yaml:"ContentSecurityPolicy" env:"CONTENT_SECURITY_POLICY_PLUGIN_TYPES, overwrite"`

	// Sandbox enables a sandbox for the requested resource
	Sandbox []string `yaml:"ContentSecurityPolicy" env:"CONTENT_SECURITY_POLICY_SANDBOX, overwrite"`

	// WorkerSrc restricts the URLs which may be loaded as a Worker, SharedWorker, or ServiceWorker
	WorkerSrc []string `yaml:"ContentSecurityPolicy" env:"CONTENT_SECURITY_POLICY_WORKER_SRC, overwrite"`

	// ManifestSrc restricts the URLs that application manifests can be loaded
	ManifestSrc []string `yaml:"ContentSecurityPolicy" env:"CONTENT_SECURITY_POLICY_MANIFEST_SRC, overwrite"`

	// PrefetchSrc defines valid sources for request prefetch and prerendering
	PrefetchSrc []string `yaml:"ContentSecurityPolicy" env:"CONTENT_SECURITY_POLICY_PREFETCH_SRC, overwrite"`
}

// Encode converts the ContentSecurityPolicy struct into a valid Content-Security-Policy header string
func (csp *ContentSecurityPolicy) Encode() string {
	var policies []string

	if len(csp.DefaultSrc) > 0 {
		policies = append(policies, "default-src "+strings.Join(csp.DefaultSrc, " "))
	}
	if len(csp.ScriptSrc) > 0 {
		policies = append(policies, "script-src "+strings.Join(csp.ScriptSrc, " "))
	}
	if len(csp.StyleSrc) > 0 {
		policies = append(policies, "style-src "+strings.Join(csp.StyleSrc, " "))
	}
	if len(csp.ImgSrc) > 0 {
		policies = append(policies, "img-src "+strings.Join(csp.ImgSrc, " "))
	}
	if len(csp.ConnectSrc) > 0 {
		policies = append(policies, "connect-src "+strings.Join(csp.ConnectSrc, " "))
	}
	if len(csp.FontSrc) > 0 {
		policies = append(policies, "font-src "+strings.Join(csp.FontSrc, " "))
	}
	if len(csp.ObjectSrc) > 0 {
		policies = append(policies, "object-src "+strings.Join(csp.ObjectSrc, " "))
	}
	if len(csp.MediaSrc) > 0 {
		policies = append(policies, "media-src "+strings.Join(csp.MediaSrc, " "))
	}
	if len(csp.FrameSrc) > 0 {
		policies = append(policies, "frame-src "+strings.Join(csp.FrameSrc, " "))
	}
	if len(csp.FormAction) > 0 {
		policies = append(policies, "form-action "+strings.Join(csp.FormAction, " "))
	}
	if len(csp.PluginTypes) > 0 {
		policies = append(policies, "plugin-types "+strings.Join(csp.PluginTypes, " "))
	}
	if len(csp.Sandbox) > 0 {
		policies = append(policies, "sandbox "+strings.Join(csp.Sandbox, " "))
	}
	if len(csp.WorkerSrc) > 0 {
		policies = append(policies, "worker-src "+strings.Join(csp.WorkerSrc, " "))
	}
	if len(csp.ManifestSrc) > 0 {
		policies = append(policies, "manifest-src "+strings.Join(csp.ManifestSrc, " "))
	}
	if len(csp.PrefetchSrc) > 0 {
		policies = append(policies, "prefetch-src "+strings.Join(csp.PrefetchSrc, " "))
	}

	return strings.Join(policies, "; ")
}
