package prober

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type BodyConfig struct {
	Plain    string `yaml:"plain"`
	Template string `yaml:"template"`
}

type WebhookConfig struct {
	Headers map[string]string `yaml:"headers"`
	Body    BodyConfig        `yaml:"body"`
}

func ConfigFromFile(path string) (WebhookConfig, error) {
	config := WebhookConfig{}
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}
