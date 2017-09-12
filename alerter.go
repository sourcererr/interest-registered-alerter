package main

import "fmt"

// Alerter - interface
type Alerter interface {
	InterestRegistered(emailAddress string) error
}

// SlackAlerter - sends alerts to slack
type SlackAlerter struct {
	slackUrl string
}

func (a *SlackAlerter) InterestRegistered(emailAddress string) error {
	fmt.Printf("Interest registered alert sent to slack")
	return nil
}

// NewSlackAlerter - create new slack alerter instance
func NewSlackAlerter(slackUrl string) *SlackAlerter {
	return &SlackAlerter{
		slackUrl: "www.slackurl.com",
	}
}
