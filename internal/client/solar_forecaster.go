package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// RunRequest represents the expected JSON payload for the /run endpoint.
type RunRequest struct {
	PrometheusURL       string  `json:"PROMETHEUS_URL"`
	MetricName          string  `json:"METRIC_NAME"`
	TrainDays           int     `json:"TRAIN_DAYS"`
	BatteryCapacityWith float64 `json:"BATTERY_CAPACITY_WH"`
	InitialSocPercent   float64 `json:"INITIAL_SOC_PERCENT"`
	ConstantLoadW       float64 `json:"CONSTANT_LOAD_W"`
	DetailedSummary     bool    `json:"DETAILED_SUMMARY"`
	UseCython           bool    `json:"USE_CYTHON"`
}

// SolarForecasterClient is a client for interacting with the Solar Forecaster service.
type SolarForecasterClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewSolarForecasterClient(baseURL string) *SolarForecasterClient {
	return &SolarForecasterClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second, // Request timeout
		},
	}
}

// genericRequest, tüm istekler için ortak bir işleyici olarak çalışır.
func (sfc *SolarForecasterClient) genericRequest(method, path string, body io.Reader, headers map[string]string) (map[string]interface{}, error) {
	req, err := http.NewRequest(method, sfc.baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := sfc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// RunForecast, /run endpoint'ine POST isteği gönderir.
func (sfc *SolarForecasterClient) RunForecast(reqData RunRequest) (map[string]interface{}, error) {
	body, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request data: %w", err)
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	return sfc.genericRequest("POST", "/run", bytes.NewReader(body), headers)
}

// UploadEnvFile, /upload-env endpoint'ine dosya yükleme işlemi yapar.
func (sfc *SolarForecasterClient) UploadEnvFile(filePath string) (map[string]interface{}, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open env file: %w", err)
	}
	defer file.Close()

	// bir buffer ve multipart writer oluştur
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	// env_file adında bir form alanı oluştur ve dosyayı buraya ekle
	part, err := writer.CreateFormFile("env_file", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("failed to copy file content: %w", err)
	}
	// writer'ı kapat (bu, form verisinin sonunu belirtir)
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	// isteği oluştur ve content-type header'ını ayarla
	req, err := http.NewRequest("POST", sfc.baseURL+"/upload-env", body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// isteği gönder
	resp, err := sfc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// RunWithEnv, /run-with-env/{session_id} endpoint'ine POST isteği gönderir.
func (sfc *SolarForecasterClient) RunWithEnv(sessionID string, overrides map[string]interface{}) (map[string]interface{}, error) {
	body, err := json.Marshal(overrides)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal overrides: %w", err)
	}

	headers := map[string]string{
		"Content-Type": "application/json",
		"Session-ID":   sessionID,
	}

	return sfc.genericRequest("POST", "/run-with-env/"+sessionID, bytes.NewReader(body), headers)
}

// GetSessions, /sessions endpoint'ine GET isteği gönderir.
func (sfc *SolarForecasterClient) GetSessions() (map[string]interface{}, error) {
	return sfc.genericRequest("GET", "/sessions", nil, nil)
}

// DeleteSession, /delete-session/{session_id} endpoint'ine DELETE isteği gönderir.
func (sfc *SolarForecasterClient) DeleteSession(sessionID string) (map[string]interface{}, error) {
	headers := map[string]string{
		"Session-ID": sessionID,
	}
	return sfc.genericRequest("DELETE", "/sessions/"+sessionID, nil, headers)
}

// GetSampleEnv, /sample-env endpoint'ine GET isteği gönderir.
func (sfc *SolarForecasterClient) GetSampleEnv() (map[string]interface{}, error) {
	return sfc.genericRequest("GET", "/sample-env", nil, nil)
}
