// Command mqd-server-mock is a local stand-in for the central MQD server that
// mqd-client (src/) talks to via PROXY_URL. It implements just enough of the
// real API (GET /settings/{file}, POST /token, POST /report) for mqd-client
// to start and run locally without a real server. See openspec/changes/mqd-server-mock
// for the spec this implements. Not representative of production validation rules.
package main

import (
	"embed"
	"encoding/json"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"
)

//go:embed fixtures
var embeddedFixtures embed.FS

// jwkToken mirrors src/crosscutting/security/jwt.JWKToken's JSON shape.
type jwkToken struct {
	AccessToken      string `json:"access_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	NotBeforePolicy  int    `json:"not-before-policy"`
	Scope            string `json:"scope"`
}

// reportSummary decodes only the fields of src/domain/models.Report needed for logging.
type reportSummary struct {
	ClientID    string `json:"ClientID"`
	DataOwnerID string `json:"DataOwnerID"`
}

type server struct {
	fixturesDir string
}

func (s *server) readFixture(fileName string) ([]byte, error) {
	if s.fixturesDir != "" {
		return os.ReadFile(filepath.Join(s.fixturesDir, fileName))
	}

	return fs.ReadFile(embeddedFixtures, path.Join("fixtures", fileName))
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// handleSettings implements GET /settings/{fileName}.
func (s *server) handleSettings(w http.ResponseWriter, r *http.Request) {
	fileName := r.PathValue("fileName")

	data, err := s.readFixture(fileName)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "settings file not found: "+fileName)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// handleToken implements POST /token.
func (s *server) handleToken(w http.ResponseWriter, r *http.Request) {
	// grant_type / client_id are accepted but not validated - this is a mock.
	_ = r.ParseForm()

	writeJSON(w, http.StatusOK, jwkToken{
		AccessToken:      "mock-access-token",
		TokenType:        "Bearer",
		ExpiresIn:        3600,
		RefreshExpiresIn: 0,
		NotBeforePolicy:  0,
		Scope:            "mock",
	})
}

// handleReport implements POST /report.
func (s *server) handleReport(w http.ResponseWriter, r *http.Request) {
	var report reportSummary
	if err := json.NewDecoder(r.Body).Decode(&report); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid report body: "+err.Error())
		return
	}

	log.Printf("received report: ClientID=%s DataOwnerID=%s", report.ClientID, report.DataOwnerID)
	writeJSON(w, http.StatusOK, map[string]string{"status": "accepted"})
}

// loggingMiddleware logs method, path and status of every request.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		log.Printf("%s %s -> %d", r.Method, r.URL.Path, sw.status)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(status int) {
	sw.status = status
	sw.ResponseWriter.WriteHeader(status)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}

func main() {
	port := flag.String("port", envOr("PORT", "8082"), "port to listen on")
	fixturesDir := flag.String("fixtures", envOr("FIXTURES_DIR", ""), "directory with fixture files (defaults to the embedded fixtures)")
	flag.Parse()

	s := &server{fixturesDir: *fixturesDir}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /settings/{fileName}", s.handleSettings)
	mux.HandleFunc("POST /token", s.handleToken)
	mux.HandleFunc("POST /report", s.handleReport)

	addr := ":" + *port
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           loggingMiddleware(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("mqd-server-mock listening on %s (fixtures: %s)", addr, fixturesSource(*fixturesDir))
	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func fixturesSource(fixturesDir string) string {
	if fixturesDir == "" {
		return "embedded defaults"
	}

	return fixturesDir
}
