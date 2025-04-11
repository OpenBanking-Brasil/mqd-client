package application

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"
)

func Test_ValidateResponseHeader(t *testing.T) {
	t.Run("should set response headers correctly for secure content", func(t *testing.T) {
		ctx := context.Background()

		// Load the docker-compose file
		composeFilePath := "../../infra/dockerfile/docker-compose.yaml"
		composeStack, err := compose.NewDockerCompose(composeFilePath)
		if err != nil {
			t.Fatalf("Failed to load docker-compose file: %v", err)
		}

		// Start the stack
		err = composeStack.
			WaitForService("mqd-client", wait.ForListeningPort("8080/tcp")).
			Up(ctx, compose.Wait(true))
		if err != nil {
			t.Fatalf("Failed to start docker-compose stack: %v", err)
		}
		defer func() {
			// Tear down the stack after the test
			err := composeStack.Down(ctx, compose.RemoveOrphans(true), compose.RemoveVolumes(true))
			if err != nil {
				t.Logf("Failed to stop docker-compose stack: %v", err)
			}
		}()

		// Perform a request to the /ValidateResponse endpoint
		url := "http://localhost:8080/ValidateResponse"
		reqBody := []byte(``)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Content-Type", "text/plain")
		req.Header.Set("serverOrgId", "test_id_1234")
		req.Header.Set("endpointNamee", "/accounts/v2/accounts")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
		assert.Equal(t, "max-age=4536000", resp.Header.Get("Strict-Transport-Security"))
		assert.Equal(t, "SAMEORIGIN", resp.Header.Get("X-Frame-Options"))
		assert.Contains(t, resp.Header.Get("Content-Security-Policy"), "default-src 'self';")
		assert.Contains(t, resp.Header.Get("Content-Security-Policy"), "script-src 'self' https://trusted-scripts.com;")
		assert.Contains(t, resp.Header.Get("Content-Security-Policy"), "style-src 'self' https://trusted-styles.com;")
		assert.Contains(t, resp.Header.Get("Content-Security-Policy"), "img-src 'self';")
		assert.Contains(t, resp.Header.Get("Content-Security-Policy"), "connect-src 'self';")
		assert.Contains(t, resp.Header.Get("Content-Security-Policy"), "font-src 'self';")
		assert.Contains(t, resp.Header.Get("Content-Security-Policy"), "object-src 'none';")
		assert.Contains(t, resp.Header.Get("Content-Security-Policy"), "media-src 'none';")
		assert.Contains(t, resp.Header.Get("Content-Security-Policy"), "frame-src 'self';")
		assert.Contains(t, resp.Header.Get("Content-Security-Policy"), "form-action 'self';")
		assert.Contains(t, resp.Header.Get("Content-Security-Policy"), "plugin-types 'application/json';")
		assert.Contains(t, resp.Header.Get("Content-Security-Policy"), "sandbox 'allow-scripts';")
		assert.Contains(t, resp.Header.Get("Content-Security-Policy"), "worker-src 'self';")
		assert.Contains(t, resp.Header.Get("Content-Security-Policy"), "manifest-src 'self';")
		assert.Contains(t, resp.Header.Get("Content-Security-Policy"), "prefetch-src 'self'")
	})
}
