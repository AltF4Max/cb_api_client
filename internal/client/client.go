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

func NewBaseClient(config *CleverbridgeConfig) *BaseClient {
	return &BaseClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: config.BaseURL,
		config:  config,
	}
}

func (c *BaseClient) getBasicAuth() string {
	auth := c.config.ClientID + ":" + c.config.ClientSecret
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (c *BaseClient) SendRequest(ctx context.Context, request *Request) (*Response, error) {
	fullURL := c.baseURL + request.Path
	if request.QueryParams != nil && len(request.QueryParams) > 0 {
		params := url.Values{}
		for key, value := range request.QueryParams {
			params.Add(key, value)
		}
		fullURL = fullURL + "?" + params.Encode()
	}
	var reqBody io.Reader
	if request.Body != nil {
		jsonData, err := json.Marshal(request.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)
	}
	req, err := http.NewRequestWithContext(ctx, request.Method, fullURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	for key, value := range request.Headers {
		req.Header.Set(key, value)
	}
	if c.config.Debug {
		fmt.Printf("🔧 Sending %s request to: %s\n", request.Method, fullURL)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	response := &Response{
		StatusCode: resp.StatusCode,
		Body:       body,
		Headers:    resp.Header,
	}
	if resp.StatusCode >= 400 {
		return response, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	return response, nil
}

func (c *BaseClient) GetSubscription(ctx context.Context, subscriptionID, isCurrent string) (*Subscription, error) {
	request := &Request{
		Method: "GET",
		Path:   "/subscription/getsubscription",
		QueryParams: map[string]string{
			"subscriptionId": subscriptionID,
			"isCurrent":      isCurrent,
		},
		Headers: map[string]string{
			"Authorization": "Basic " + c.baseClient.getBasicAuth(),
			"Content-Type":  "application/json",
			"Accept":        "application/json",
		},
	}

	response, err := c.baseClient.SendRequest(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	var subscription Subscription
	if err := json.Unmarshal(response.Body, &subscription); err != nil {
		return nil, fmt.Errorf("failed to parse subscription: %w", err)
	}

	return &subscription, nil
}
func (c *BaseClient) GetSubscriptionsByPurchase(ctx context.Context, purchaseID string) ([]Subscription, error) {
	request := &Request{
		Method: "GET",
		Path:   "/subscription/getsubscriptionsbypurchase",
		QueryParams: map[string]string{
			"purchaseId": purchaseID,
		},
		Headers: map[string]string{
			"Authorization": "Basic " + c.baseClient.getBasicAuth(),
			"Content-Type":  "application/json",
			"Accept":        "application/json",
		},
	}

	response, err := c.baseClient.SendRequest(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions by purchase: %w", err)
	}

	var subscriptions []Subscription
	if err := json.Unmarshal(response.Body, &subscriptions); err != nil {
		return nil, fmt.Errorf("failed to parse subscriptions: %w", err)
	}

	return subscriptions, nil
}
func (c *BaseClient) GetSubscriptionsForCustomer(ctx context.Context, customerID string) ([]Subscription, error) {
	request := &Request{
		Method: "GET",
		Path:   "/subscription/getsubscriptionsforcustomer",
		QueryParams: map[string]string{
			"customerId": customerID,
		},
		Headers: map[string]string{
			"Authorization": "Basic " + c.baseClient.getBasicAuth(),
			"Content-Type":  "application/json",
			"Accept":        "application/json",
		},
	}

	response, err := c.baseClient.SendRequest(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions for customer: %w", err)
	}

	var subscriptions []Subscription
	if err := json.Unmarshal(response.Body, &subscriptions); err != nil {
		return nil, fmt.Errorf("failed to parse subscriptions: %w", err)
	}

	return subscriptions, nil
}
