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
	Type         string
	Duration     time.Duration
	Retry        uint
	Webhook      string
	HTTPSettings HTTPSettings
	UDPSettings  UDPSettings
	TCPSettings  TCPSettings
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
	probeType := os.Getenv("PROBER_TYPE")
	probeDuration := os.Getenv("PROBER_DURATION")
	webhook := os.Getenv("PROBER_WEBHOOK")
	duration, err := time.ParseDuration(probeDuration)

	if err != nil {
		log.Printf("failed to parse `%s` as time.Duration, using default value: %s", probeDuration, DEFAULT_PROBE_DURATION)
		duration = DEFAULT_PROBE_DURATION
	}

	prober := Prober{
		Type:         strings.ToUpper(probeType),
		Duration:     duration,
		Retry:        GetUintEnvDefault("PROBER_RETRY", DEFAULT_HTTP_RETRY),
		Webhook:      webhook,
		HTTPSettings: HTTPSettings{},
		TCPSettings:  TCPSettings{},
		UDPSettings:  UDPSettings{},
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
	// Validate Duartion
	if prober.Duration < time.Millisecond {
		return errors.New("Duration is invalid")
	}

	// Validate Retry
	if prober.Retry < 0 {
		return errors.New("Retry is invalid")
	}

	if len(prober.Type) == 0 {
		return errors.New("Type of prober is not set")
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

func TriggerWebhook(prober *Prober) {
	if len(prober.Webhook) > 0 {
		output, err := json.Marshal(status.Status)
		if err != nil {
			log.Printf("Failed to marshal json")
			return
		}
		resp, err := http.Post(prober.Webhook, "application/json", bytes.NewBuffer(output))
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
		TriggerWebhook(prober)
		log.Printf("STATUS: %s - %s", status.Status.Status, time.Now().UTC())
		time.Sleep(prober.Duration)
	}
}
