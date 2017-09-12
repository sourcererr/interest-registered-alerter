package main

import (
	"bytes"
	"fmt"
	"net/http"
)

// Alerter - interface
type Alerter interface {
	InterestRegistered(emailAddress string) error
}

// SlackAlerter - sends alerts to slack
type SlackAlerter struct {
	slackUrl string
}

func (a *SlackAlerter) InterestRegistered(emailAddress string) error {
	var jsonStr = []byte(`{"text": "Interest registered by: ` + emailAddress + `"}`)

	req, err := http.NewRequest("POST", a.slackUrl, bytes.NewBuffer(jsonStr))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Sending alert to slack failed with code: %v\n", resp.StatusCode)
	}

	return nil
}

// NewSlackAlerter - create new slack alerter instance
func NewSlackAlerter(slackUrl string) *SlackAlerter {
	return &SlackAlerter{
		slackUrl: slackUrl,
	}
}
