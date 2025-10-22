package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// NewCleverbridgeClient
func NewCleverbridgeClient(config *CleverbridgeConfig) *CleverbridgeClient {
	return &CleverbridgeClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		config:  config,
		baseURL: config.BaseURL,
		logger:  NewLogger(config.Debug, config.LogFile),
	}
}

// CleverbridgeDoRequest
func (c *CleverbridgeClient) CleverbridgeDoRequest(ctx context.Context, method, path string, queryParams map[string]string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			c.logger.Error("Failed to marshal request body", err,
				map[string]interface{}{"method": method, "path": path})
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)
	}

	fullURL := c.baseURL + path
	if queryParams != nil && len(queryParams) > 0 {
		params := url.Values{}
		for key, value := range queryParams {
			params.Add(key, value)
		}
		fullURL = fullURL + "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
		c.logger.Error("Failed to create HTTP request", err,
			map[string]interface{}{"method": method, "url": fullURL})
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Cleverbridge: Basic Auth
	req.SetBasicAuth(c.config.ClientID, c.config.ClientSecret)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	c.logger.Debug("Making Cleverbridge API request",
		map[string]interface{}{"method": method, "url": fullURL})

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("HTTP request failed", err,
			map[string]interface{}{"method": method, "url": fullURL})
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			c.logger.Error("Failed to read error response body", err,
				map[string]interface{}{
					"method":     method,
					"url":        fullURL,
					"statusCode": resp.StatusCode,
				})
			return nil, fmt.Errorf("request failed with status: %s", resp.Status)
		}

		// Cleverbridge может возвращать ошибки в другом формате
		var cleverbridgeError struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}

		if err := json.Unmarshal(bodyBytes, &cleverbridgeError); err == nil && cleverbridgeError.Error != "" {
			c.logger.Error("Cleverbridge API error", nil,
				map[string]interface{}{
					"method":  method,
					"url":     fullURL,
					"status":  resp.Status,
					"error":   cleverbridgeError.Error,
					"message": cleverbridgeError.Message,
				})
			return nil, fmt.Errorf("Cleverbridge API error: %s - %s", cleverbridgeError.Error, cleverbridgeError.Message)
		}

		c.logger.Error("Failed to decode Cleverbridge error response", nil,
			map[string]interface{}{
				"method":     method,
				"url":        fullURL,
				"statusCode": resp.StatusCode,
				"response":   string(bodyBytes),
			})
		return nil, fmt.Errorf("request failed with status: %s, response: %s", resp.Status, string(bodyBytes))
	}

	return resp, nil
}

func (c *CleverbridgeClient) GetSubscription(ctx context.Context, subscriptionID string, isCurrent string) (*CleverbridgeSubscription, error) {
	path := "/subscription/getsubscription"
	
	queryParams := map[string]string{
		"subscriptionId": subscriptionID,
		"isCurrent":      isCurrent,
	}

	resp, err := c.CleverbridgeDoRequest(ctx, "GET", path, queryParams, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result CleverbridgeSubscription
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Error("Failed to decode Cleverbridge subscription response", err,
			map[string]interface{}{
				"subscriptionID": subscriptionID,
				"isCurrent":      isCurrent,
				"path":           path,
			})
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
func (c *CleverbridgeClient) GetSubscriptionsByPurchase
func (c *CleverbridgeClient) GetSubscriptionsCustomer
func (c *CleverbridgeClient) 
func (c *CleverbridgeClient) 

