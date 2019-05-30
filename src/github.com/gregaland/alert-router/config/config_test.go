package config

import (
	"gotest.tools/assert"
	"strings"
	"testing"
)

var data string = `
---
listen: :8000
routers:
 - id: gmail
   type: email
   enabled: true
   smtphost: smtp.gmail.com
   smtpport: 587
 - id: elastic
   type: webhook
   enabled: true
   url: https://elastic.rig.gregland.dev:9200
   username: elastic
   password: rigadmin
   query_parms:
     - token="foobar"
alerts_path: etc/alerts.d
`

var alertData = `
alert: greg
schedule:
  - id: all_day
    router_id: gmail
    email_addrs:
      - 9999999999@tmomail.net
  - id: after_hours
    start: "0 17 * * *"
    end: "0 6 * * *"
    router_id: gmail
    email_addrs:
      - john.doe@foobar.net
`

func checkRouterResult(t *testing.T, actual []*Routers) {
	idx1Expected := Routers{Type: EMAIL_RP,
		Parms: RouterParms{Id: "gmail", Enabled:true, SmtpHost: "smtp.gmail.com",
			SmtpPort: 587}}
	idx2Expected := Routers{Type: WEBHOOK_RP,
		Parms: RouterParms{Id: "elastic", Enabled:true, Url: "https://elastic.rig.gregland.dev:9200",
			Username: "elastic", Password: "rigadmin", QueryParms: []string{"token=\"foobar\""} }}
	assert.DeepEqual(t, idx1Expected, actual[0])
	assert.DeepEqual(t, idx2Expected, actual[1])
}

func TestNewRigConfig(t *testing.T) {
	yamlData := strings.NewReader(data)
	config, err := NewRigConfig(yamlData)
	if err != nil {
		t.Fatal(err)
	}
	checkRouterResult(t, config.Routers)
	assert.Equal(t, ":8000", config.Listen)
	assert.Equal(t, "etc/alerts.d", config.AlertsPath)
}

func TestRigConfig_LoadAlertConfig(t *testing.T) {
	yamlData := strings.NewReader(alertData)
	actual, err := LoadAlertConfig(yamlData)
	if err != nil {
		t.Fatal(err)
	}

	schedule := make([]RouterParms,0)
	schedule = append(schedule, RouterParms{Id: "all_day",
		ScheduleStart: "", ScheduleEnd: "", RouterId: "gmail",
		EmailAddrs: []string{"9999999999@tmomail.net"}})

	schedule = append(schedule, RouterParms{Id: "after_hours",
		ScheduleStart: "0 17 * * *", ScheduleEnd: "0 6 * * *",
		RouterId: "gmail", EmailAddrs: []string{"john.doe@foobar.net"}})

	expected := AlertConfig{AlertId:"greg", Schedule: schedule}
	assert.DeepEqual(t, expected, *actual)
}
