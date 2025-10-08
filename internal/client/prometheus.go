package client

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/api"
	prometheusV1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// PrometheusClient is a client for interacting with Prometheus API.
type PrometheusClient struct {
	api prometheusV1.API
}

func NewPrometheusClient(prometheusURL string) (*PrometheusClient, error) {
	client, err := api.NewClient(api.Config{
		Address: prometheusURL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Prometheus client: %w", err)
	}

	return &PrometheusClient{
		api: prometheusV1.NewAPI(client),
	}, nil
}

// Query anlık bi PromQL sorgusu çalıştırır ve sonucu döner.
func (pc *PrometheusClient) Query(query string) (model.Value, error) {
	// 5 saniyelik bir zaman aşımı ile sorguyu çalıştır
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Sorguyu çalıştır
	result, warnings, err := pc.api.Query(ctx, query, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	if len(warnings) > 0 {
		fmt.Printf("Warnings: %v\n", warnings)
	}

	return result, nil
}
