package routers

import (
	"bytes"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"net/http"
)

const (
	SLACK_DEFAULT_MAX_MSG_SIZE int = 160
)

type SlackMessage struct {
	Text string `json:"text"`
}

type SlackConfig struct {
	Url        string
	MaxMsgSize int // defaults
}

type SlackRouter struct {
	Config *SlackConfig
}

func NewSlackRouter(config *SlackConfig) (Router, error) {

	r := &SlackRouter{Config: config}
	if r.Config.MaxMsgSize == 0 {
		r.Config.MaxMsgSize = SLACK_DEFAULT_MAX_MSG_SIZE
	}
	log.WithFields(log.Fields{
		"maxmsgsize": r.Config.MaxMsgSize,
	}).Info("slack router constructed")

	return r, nil
}

func (e *SlackRouter) Init() error {
	return nil
}

func (e *SlackRouter) GetConfig() interface{} {
	return *e.Config
}

func (e *SlackRouter) Route(event *Event, t interface{}) error {
	log.Debug("entering slack route")
	var err error = nil

	log.WithFields(log.Fields{
		"id":      event.Id,
		"message": event.Message,
		"url":     e.Config.Url,
	}).Info("routing")

	if len(event.Message) > e.Config.MaxMsgSize {
		event.Message = event.Message[:e.Config.MaxMsgSize]
	}
	msg, err := json.Marshal(&SlackMessage{Text: event.Id + ": " + event.Message})
	if err != nil {
		log.Error(err)
	} else {
		req, err := http.NewRequest("POST", e.Config.Url, bytes.NewBuffer(msg))
		if err != nil {
			log.Error(err)
		}
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Error(err)
		} else if resp.StatusCode != http.StatusOK {
			log.Errorf("Status Code: %d", resp.StatusCode)
		}
	}

	return err
}
