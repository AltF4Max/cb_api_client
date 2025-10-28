package client

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

func NewAPIClient(config *CleverbridgeConfig) *APIClient {
	baseClient := BaseClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: config.BaseURL,
		config:  config,
	}

	return &APIClient{
		BaseClient: baseClient,
		logger:     NewLogger(config.Debug, ""),
	}
}

func (c *APIClient) getBasicAuth() string {
	auth := c.config.ClientID + ":" + c.config.ClientSecret
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (c *APIClient) sendRequest(ctx context.Context, method, path string, queryParams map[string]string, body interface{}) ([]byte, error) {
	fullURL := c.baseURL + path
	if queryParams != nil && len(queryParams) > 0 {
		params := url.Values{}
		for key, value := range queryParams {
			params.Add(key, value)
		}
		fullURL = fullURL + "?" + params.Encode()
	}

	c.logger.Info("Sending API request",
		"method", method,
		"url", fullURL,
		"path", path)

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			c.logger.Error("Failed to marshal request body", err,
				"method", method, "path", path)
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)

		if c.config.Debug {
			c.logger.Json(map[string]interface{}{
				"request_body": string(jsonData),
				"method":       method,
				"path":         path,
			})
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
		c.logger.Error("Failed to create HTTP request", err,
			"method", method, "url", fullURL)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Basic "+c.getBasicAuth())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	startTime := time.Now()
	resp, err := c.httpClient.Do(req)
	requestDuration := time.Since(startTime)

	if err != nil {
		c.logger.Error("HTTP request failed", err,
			"method", method,
			"url", fullURL,
			"duration", requestDuration.String())
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("Failed to read response body", err,
			"method", method,
			"url", fullURL,
			"status_code", resp.StatusCode)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	c.logger.Info("API response received",
		"method", method,
		"path", path,
		"status_code", resp.StatusCode,
		"duration", requestDuration.String(),
		"response_size", len(responseBody))

	if c.config.Debug && len(responseBody) > 0 {
		c.logger.Json(map[string]interface{}{
			"response_body": string(responseBody),
			"status_code":   resp.StatusCode,
			"method":        method,
			"path":          path,
		})
	}

	if resp.StatusCode >= 400 {
		c.logger.Error("API returned error response", nil,
			"method", method,
			"url", fullURL,
			"status_code", resp.StatusCode,
			"response", string(responseBody))
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(responseBody))
	}

	return responseBody, nil
}
func (c *APIClient) GetSubscription(ctx context.Context, subscriptionID, isCurrent string) (*Subscription, error) {
	c.logger.Info("Getting subscription",
		"subscription_id", subscriptionID,
		"is_current", isCurrent)

	queryParams := map[string]string{
		"subscriptionId": subscriptionID,
		"isCurrent":      isCurrent,
	}

	responseBody, err := c.sendRequest(ctx, "GET", "/subscription/getsubscription", queryParams, nil)
	if err != nil {
		c.logger.Error("Failed to get subscription", err,
			"subscription_id", subscriptionID)
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	var subscription Subscription
	if err := json.Unmarshal(responseBody, &subscription); err != nil {
		c.logger.Error("Failed to parse subscription response", err,
			"subscription_id", subscriptionID,
			"response_body", string(responseBody))
		return nil, fmt.Errorf("failed to parse subscription: %w", err)
	}

	c.logger.Info("Successfully retrieved subscription",
		"subscription_id", subscription.ID,
		"status", subscription.Status,
		"plan", subscription.Plan)

	return &subscription, nil
}

func (c *APIClient) GetSubscriptionsByPurchase(ctx context.Context, purchaseID string) ([]Subscription, error) {
	c.logger.Info("Getting subscriptions by purchase", "purchase_id", purchaseID)

	queryParams := map[string]string{
		"purchaseId": purchaseID,
	}

	responseBody, err := c.sendRequest(ctx, "GET", "/subscription/getsubscriptionsbypurchase", queryParams, nil)
	if err != nil {
		c.logger.Error("Failed to get subscriptions by purchase", err,
			"purchase_id", purchaseID)
		return nil, fmt.Errorf("failed to get subscriptions by purchase: %w", err)
	}

	var subscriptions []Subscription
	if err := json.Unmarshal(responseBody, &subscriptions); err != nil {
		c.logger.Error("Failed to parse subscriptions response", err,
			"purchase_id", purchaseID,
			"response_body", string(responseBody))
		return nil, fmt.Errorf("failed to parse subscriptions: %w", err)
	}

	c.logger.Info("Successfully retrieved subscriptions by purchase",
		"purchase_id", purchaseID,
		"subscriptions_count", len(subscriptions))

	return subscriptions, nil
}

func (c *APIClient) GetSubscriptionsForCustomer(ctx context.Context, customerID string) ([]Subscription, error) {
	c.logger.Info("Getting subscriptions for customer", "customer_id", customerID)

	queryParams := map[string]string{
		"customerId": customerID,
	}

	responseBody, err := c.sendRequest(ctx, "GET", "/subscription/getsubscriptionsforcustomer", queryParams, nil)
	if err != nil {
		c.logger.Error("Failed to get subscriptions for customer", err,
			"customer_id", customerID)
		return nil, fmt.Errorf("failed to get subscriptions for customer: %w", err)
	}

	var subscriptions []Subscription
	if err := json.Unmarshal(responseBody, &subscriptions); err != nil {
		c.logger.Error("Failed to parse subscriptions response", err,
			"customer_id", customerID,
			"response_body", string(responseBody))
		return nil, fmt.Errorf("failed to parse subscriptions: %w", err)
	}

	c.logger.Info("Successfully retrieved subscriptions for customer",
		"customer_id", customerID,
		"subscriptions_count", len(subscriptions))

	return subscriptions, nil
}
