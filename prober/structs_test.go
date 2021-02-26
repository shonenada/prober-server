package prober

import (
	"strings"
	"testing"
)

func buildProber(headers map[string]string) *Prober {
	return &Prober{
		Name: "testing",
		WebhookConfig: WebhookConfig{
			Version: "1",
			Headers: headers,
			Body: BodyConfig{
				Plain:    "",
				Template: "",
			},
		},
	}

}

func TestBuildHeaders_Default(t *testing.T) {
	prober := buildProber(map[string]string{})
	rv := prober.BuildHeaders()
	if rv.Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type should be `application/json`")
	}
}

func TestBuildHeaders_WithContentType(t *testing.T) {
	prober := buildProber(map[string]string{
		"Content-Type": "plain/text",
	})
	rv := prober.BuildHeaders()
	if rv.Get("Content-Type") != "plain/text" {
		t.Errorf("Content-Type should be `plain/text`")
	}
}

func TestBuildHeaders_WithLowerCaseKey(t *testing.T) {
	key := "user-agent"
	value := "test-prober"
	prober := buildProber(map[string]string{
		key: value,
	})
	rv := prober.BuildHeaders()
	if rv.Get(strings.ToUpper(key)) != value {
		t.Errorf("%s should be `%s`", key, value)
	}
}

func TestBuildHeaders_WithMultiKeys(t *testing.T) {
	prober := buildProber(map[string]string{
		"Accept":       "application/json",
		"User-Agent":   "TestProber",
		"Content-Type": "plain/html",
	})
	rv := prober.BuildHeaders()
	if rv.Get("Accept") != "application/json" ||
		rv.Get("User-Agent") != "TestProber" ||
		rv.Get("Content-Type") != "plain/html" {
		t.Errorf("Test failed")
	}
}
