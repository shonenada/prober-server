package prober

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/shonenada/prober-server/status"
)

const DEFAULT_PROBE_DURATION = 60 * time.Second
const DEFAULT_HTTP_TIMEOUT = 30
const DEFAULT_HTTP_RETRY = 3
const DEFAULT_TCP_RETRY = 3
const DEFAULT_UDP_RETRY = 3

type HTTPSettings struct {
	URL     string
	Timeout uint
}

type TCPSettings struct {
	Host string
	Port uint
}

type UDPSettings struct {
	Host string
	Port uint
}

type Prober struct {
	Name          string
	Type          string
	Duration      time.Duration
	Retry         uint
	Webhook       string
	WebhookConfig WebhookConfig
	HTTPSettings  HTTPSettings
	UDPSettings   UDPSettings
	TCPSettings   TCPSettings
}

func GetUintEnvDefault(key string, defaultValue uint) uint {
	v := os.Getenv(key)
	if len(v) == 0 {
		log.Printf("Env `%s` is not set, using default value %d", key, defaultValue)
		return defaultValue
	}
	u, err := strconv.ParseUint(v, 10, 32)
	if err != nil {
		log.Printf("failed to parse `%s` as uint, using default value: %d", v, defaultValue)
		return defaultValue
	}
	return uint(u)
}

func BuildProber() (*Prober, error) {
	name := os.Getenv("PROBER_NAME")
	probeType := os.Getenv("PROBER_TYPE")
	probeDuration := os.Getenv("PROBER_DURATION")
	webhook := os.Getenv("PROBER_WEBHOOK")
	config := os.Getenv("PROBER_CONFIG")
	triggerOnStatusChange := os.Getenv("PROBER_TRIGGER_ON_STATUS_CHANGE")

	duration, err := time.ParseDuration(probeDuration)
	if err != nil {
		log.Printf("failed to parse `%s` as time.Duration, using default value: %s", probeDuration, DEFAULT_PROBE_DURATION)
		duration = DEFAULT_PROBE_DURATION
	}

	var webhookConfig WebhookConfig
	if len(config) > 0 {
		webhookConfig, err = ConfigFromFile(config)
		if err != nil {
			log.Printf("Failed to build config from file: %s", err)
			webhookConfig = WebhookConfig{}
		}
	} else {
		webhookConfig = WebhookConfig{}
	}

	webhookConfig.StatusChangeOnly = (triggerOnStatusChange == "true")

	prober := Prober{
		Name:          name,
		Type:          strings.ToUpper(probeType),
		Duration:      duration,
		Retry:         GetUintEnvDefault("PROBER_RETRY", DEFAULT_HTTP_RETRY),
		Webhook:       webhook,
		WebhookConfig: webhookConfig,
		HTTPSettings:  HTTPSettings{},
		TCPSettings:   TCPSettings{},
		UDPSettings:   UDPSettings{},
	}

	if prober.Type == "HTTP" {
		prober.HTTPSettings.URL = os.Getenv("PROBER_HTTP_URL")
		prober.HTTPSettings.Timeout = GetUintEnvDefault("PROBER_HTTP_TIMEOUT", DEFAULT_HTTP_TIMEOUT)
	} else if prober.Type == "TCP" {
		prober.TCPSettings.Host = os.Getenv("PROBER_TCP_HOST")
		prober.TCPSettings.Port = GetUintEnvDefault("PROBER_TCP_PORT", 0)
	} else if prober.Type == "UDP" {
		prober.UDPSettings.Host = os.Getenv("PROBER_UDP_HOST")
		prober.UDPSettings.Port = GetUintEnvDefault("PROBER_UDP_PORT", 0)
	}

	return &prober, nil
}

func (prober *Prober) Valid() error {
	if len(prober.Type) == 0 {
		return errors.New("Type of prober is not set")
	}

	// Validate Duartion
	if prober.Duration < time.Millisecond {
		return errors.New("Duration is invalid")
	}

	// Validate Retry
	if prober.Retry < 0 {
		return errors.New("Retry is invalid")
	}

	if len(prober.Webhook) > 0 {
		u, err := url.ParseRequestURI(prober.Webhook)
		if err != nil {
			return errors.New("Webhook is invalid")
		}
		scheme := strings.ToUpper(u.Scheme)
		if scheme != "HTTP" && scheme != "HTTPS" {
			return fmt.Errorf("Webhook URL with scheme %s is not supported, only HTTP/HTTPS supported", scheme)
		}

		if len(prober.Name) == 0 {
			return errors.New("Name cannot be empty when webhook is set")
		}
	}

	if prober.Type == "HTTP" {
		httpSettings := prober.HTTPSettings
		// Validate URL
		if len(httpSettings.URL) == 0 {
			return errors.New("HTTP URL is not set")
		}
		u, err := url.ParseRequestURI(httpSettings.URL)
		if err != nil {
			return errors.New("HTTP URL is invalid")
		}
		scheme := strings.ToUpper(u.Scheme)
		if scheme != "HTTP" && scheme != "HTTPS" {
			return fmt.Errorf("HTTP URL with scheme %s is not supported, only HTTP/HTTPS supported", scheme)
		}

		// Validate Timeout
		if httpSettings.Timeout <= 0 {
			return errors.New("HTTP Timeout is invalid")
		}

		return nil
	} else if prober.Type == "TCP" {
		tcpSettings := prober.TCPSettings
		// Host
		if len(tcpSettings.Host) == 0 {
			return errors.New("TCP Host is not set")
		}

		// Port
		if tcpSettings.Port < 0 || tcpSettings.Port > 65535 {
			return errors.New("TCP Port is invalid")
		}

		return nil

	} else if prober.Type == "UDP" {
		udpSettings := prober.UDPSettings
		// Host
		if len(udpSettings.Host) == 0 {
			return errors.New("UDP Host is not set")
		}

		// Port
		if udpSettings.Port < 0 || udpSettings.Port > 65535 {
			return errors.New("UDP Port is invalid")
		}

		return nil
	} else {
		return fmt.Errorf("Type of prober `%s` is invalid", prober.Type)
	}
}

type WebhookRequest struct {
	Name        string    `json:"name"`
	Code        uint      `json:"code"`
	Status      string    `json:"status"`
	LastStatus  string    `json:"last_status"`
	Message     string    `json:"message"`
	RetryTimes  uint      `json:"retry_times"`
	LastUpdated time.Time `json:"last_updated"`
}

func (prober *Prober) BuildHeaders() http.Header {
	config := prober.WebhookConfig
	headers := http.Header{}
	for k, v := range config.Headers {
		headers.Set(k, v)
	}
	if len(headers.Get("Content-Type")) == 0 {
		headers.Set("Content-Type", "application/json")
	}
	if len(headers.Get("Content-Type")) == 0 {
		headers.Set("User-Agent", "ProberServer")
	}
	return headers
}

func (prober *Prober) BuildBody() ([]byte, error) {
	config := prober.WebhookConfig
	if len(config.Body.Plain) > 0 {
		return []byte(config.Body.Plain), nil
	} else {
		wrq := WebhookRequest{
			Name:        prober.Name,
			Code:        status.Status.Code,
			Status:      status.Status.Status,
			LastStatus:  status.Status.LastStatus,
			Message:     status.Status.Message,
			RetryTimes:  status.Status.RetryTimes,
			LastUpdated: status.Status.LastUpdated,
		}
		if len(config.Body.Template) > 0 {
			template, err := template.New("template").Parse(config.Body.Template)
			if err != nil {
				return []byte{}, err
			}

			var buff bytes.Buffer
			err = template.Execute(&buff, wrq)
			if err != nil {
				return []byte{}, err
			}
			return buff.Bytes(), nil
		} else {
			output, err := json.Marshal(wrq)
			if err != nil {
				return []byte{}, errors.New("Failed to marshal json")
			}
			return output, nil
		}
	}
}

func (prober *Prober) TriggerWebhook() {
	if len(prober.Webhook) > 0 {
		body, err := prober.BuildBody()
		if err != nil {
			log.Printf("Failed to generate body")
			return
		}
		req, err := http.NewRequest("POST", prober.Webhook, bytes.NewBuffer(body))
		req.Header = prober.BuildHeaders()
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Failed to POST webhook")
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			log.Printf("Failed to POST webhook, status code %d", resp.StatusCode)
		}
	}
}

func (prober *Prober) RunForver() {
	for {
		if prober.Type == "HTTP" {
			HTTPProbe(prober)
		} else if prober.Type == "TCP" {
			TCPProbe(prober)
		} else if prober.Type == "UDP" {
			UDPProbe(prober)
		}
		if prober.WebhookConfig.StatusChangeOnly {
			if status.Status.IsStatusChanged() {
				go prober.TriggerWebhook()
			}
		} else {
			go prober.TriggerWebhook()
		}
		log.Printf("STATUS: %s - %s", status.Status.Status, time.Now().UTC())
		time.Sleep(prober.Duration)
	}
}
