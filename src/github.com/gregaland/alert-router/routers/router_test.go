package routers

import (
	"gotest.tools/assert"
	"testing"
	"github.com/gregaland/alert-router/config"
)

func TestNewEmailRouter(t *testing.T) {

	// no config
	c := EmailConfig{}
	_, err := NewEmailRouter(&c)
	assert.Error(t, err, "smpthost must be provided")

	// test default config assignment
	// TODO need to figure out how to pass AU/AP
	c = EmailConfig{SmtpHost: "smtp.gmail.com", SmtpPort: 587}
	expect := EmailConfig{SmtpHost: "smtp.gmail.com", SmtpPort: 587,
		MsgHdr: EMAIL_DEFAULT_MSG_HDR, From: EMAIL_DEFAULT_FROM, Au: EMAIL_DEFAULT_AU,
	    Ap: EMAIL_DEFAULT_AP, MaxMsgSize: EMAIL_DEFAULT_MAX_MSG_SIZE}

	e, err := NewEmailRouter(&c)
	c = e.GetConfig().(EmailConfig)

	assert.Equal(t, expect, c)

	// test explicit config
	c = EmailConfig{SmtpHost: "smtp.gmail.com", SmtpPort: 587,
		MsgHdr: "foo", From: "foo", Au: "foo", Ap: "foo", MaxMsgSize: 100}
	e, err = NewEmailRouter(&c)
	c = e.GetConfig().(EmailConfig)
	assert.Equal(t, e.GetConfig(), c)

}
func TestEmailRouter_Route(t *testing.T) {
	c := &EmailConfig{SmtpHost: "smtp.gmail.com", SmtpPort: 587}
	e, err := NewEmailRouter(c)
	if err != nil {
		t.Fatal("failed to create email router")
	}
	e.Init()
	routes := map[string]Router{}
	routes["email"] = e

	event := &Event{Id: "watcher1.yml", Message: "woohoooo"}
	params := config.RouterParms{EmailAddrs: []string{"greg.land@gregland.dev"}}
	routes["email"].Route(event, params)

}
