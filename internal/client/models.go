package client

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Subscription struct {
	ID               string    `json:"id"`
	Status           string    `json:"status"`
	Plan             string    `json:"plan"`
	CreatedAt        time.Time `json:"created_at"`
	CustomerID       string    `json:"customer_id"`
	ProductID        string    `json:"product_id"`
	NextBillingDate  time.Time `json:"next_billing_date"`
	CurrentPeriodEnd time.Time `json:"current_period_end"`
	Amount           float64   `json:"amount"`
	Currency         string    `json:"currency"`
	BillingCycle     string    `json:"billing_cycle"`
	PurchaseID       string    `json:"purchase_id"`
}

type BaseClient struct {
	httpClient *http.Client
	baseURL    string
	config     *CleverbridgeConfig
}

type CleverbridgeConfig struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	BaseURL      string `yaml:"base_url"`
	Debug        bool   `yaml:"debug"`
}

type Request struct {
	Method      string
	Path        string
	QueryParams map[string]string
	Headers     map[string]string
	Body        interface{}
}

type Response struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

type Logger struct {
	debug   bool
	logFile *os.File
	writer  io.Writer
}

// NewLogger creates a new logger with file support
func NewLogger(debug bool, logFile string) *Logger {
	var writer io.Writer = os.Stdout

	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Printf("Failed to open log file %s: %v, using stdout", logFile, err)
		} else {
			writer = file
			return &Logger{debug: debug, logFile: file, writer: writer}
		}
	}

	return &Logger{debug: debug, writer: writer}
}

// Close closes the log file if it's open
func (l *Logger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

// Info logging information
func (l *Logger) Info(message string, fields ...interface{}) {
	if l.debug {
		msg := fmt.Sprintf("INFO: %s", message)
		if len(fields) > 0 {
			msg += fmt.Sprintf(" %v", fields)
		}
		fmt.Fprintln(l.writer, msg)
	}
}

// Warn logging of warnings
func (l *Logger) Warn(message string, fields ...interface{}) {
	msg := fmt.Sprintf("WARN: %s", message)
	if len(fields) > 0 {
		msg += fmt.Sprintf(" %v", fields)
	}
	fmt.Fprintln(l.writer, msg)
}

// Error logging errors
func (l *Logger) Error(message string, err error, fields ...interface{}) {
	msg := fmt.Sprintf("ERROR: %s", message)
	if err != nil {
		msg += fmt.Sprintf(" - %v", err)
	}
	if len(fields) > 0 {
		msg += fmt.Sprintf(" %v", fields)
	}
	fmt.Fprintln(l.writer, msg)
}

// Json logging in JSON format (analog Perl Logger->json)
func (l *Logger) Json(data map[string]interface{}) {
	if l.debug {
		jsonData, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			l.Error("JSON marshaling failed", err)
			return
		}
		fmt.Fprintf(l.writer, "JSON LOG:\n%s\n", string(jsonData))
	}
}

func (c *APIClient) Close() error {
	if c.logger != nil {
		return c.logger.Close()
	}
	return nil
}
