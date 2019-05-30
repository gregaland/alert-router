package routers

import (
	"errors"
	"fmt"
	"github.com/gregaland/alert-router/config"
	log "github.com/sirupsen/logrus"
	"net/smtp"
)

const (
	EMAIL_DEFAULT_FROM         string = "alerts@gregland.dev"
	EMAIL_DEFAULT_MSG_HDR      string = "To: alerts@gregland.dev\r\nSubject: "
	EMAIL_DEFAULT_MAX_MSG_SIZE int    = 160
)

type EmailConfig struct {
	SmtpHost   string // required
	SmtpPort   int    // required
	Au         string // defaults
	Ap         string // defaults
	From       string // defaults
	MsgHdr     string // defaults
	MaxMsgSize int    // defaults
}

type EmailRouter struct {
	Config       *EmailConfig
	smtpHostPort string
	auth         smtp.Auth
}

func NewEmailRouter(config *EmailConfig) (Router, error) {

	er := &EmailRouter{Config: config}
	if er.Config.SmtpHost == "" {
		return nil, errors.New("smpthost must be provided")
	}
	if er.Config.SmtpPort == 0 {
		return nil, errors.New("smptport must be provided")
	}
	er.smtpHostPort = fmt.Sprintf("%s:%d", er.Config.SmtpHost, er.Config.SmtpPort)
	if er.Config.Au == "" {
		er.Config.Au = config.Au
	}
	if er.Config.Ap == "" {
		er.Config.Ap = config.Ap
	}
	if er.Config.From == "" {
		er.Config.From = EMAIL_DEFAULT_FROM
	}
	if er.Config.MsgHdr == "" {
		er.Config.MsgHdr = EMAIL_DEFAULT_MSG_HDR
	}
	if er.Config.MaxMsgSize == 0 {
		er.Config.MaxMsgSize = EMAIL_DEFAULT_MAX_MSG_SIZE
	}
	log.WithFields(log.Fields{
		"smtphost":   er.Config.SmtpHost,
		"smtpport":   er.Config.SmtpPort,
		"au":         er.Config.Au,
		"ap":         er.Config.Ap,
		"from":       er.Config.From,
		"msghdr":     er.Config.MsgHdr,
		"maxmsgsize": er.Config.MaxMsgSize,
	}).Info("email router constructed")

	return er, nil
}

func (e *EmailRouter) Init() error {
	e.auth = smtp.PlainAuth("", e.Config.Au, e.Config.Ap, e.Config.SmtpHost)
	return nil
}

func (e *EmailRouter) GetConfig() interface{} {
	return *e.Config
}

func (e *EmailRouter) Route(event *Event, t interface{}) error {
	log.Debug("entering email route")
	var err error = nil

	if params, ok := t.(config.RouterParms); ok {

		if len(event.Message) > e.Config.MaxMsgSize {
			event.Message = event.Message[:e.Config.MaxMsgSize]
		}
		msg := []byte(e.Config.MsgHdr + event.Id + "\r\n" + event.Message)

		log.WithFields(log.Fields{
			"id":       event.Id,
			"message":  event.Message,
			"smtphost": e.smtpHostPort,
			"from":     e.Config.From,
			"to":       params.EmailAddrs,
		}).Info("routing")

		err = smtp.SendMail(e.smtpHostPort, e.auth, e.Config.From, params.EmailAddrs, msg)
		if err != nil {
			log.Error("Failed to send email routes")
		}
	} else {
		log.Error("expected RouterParms object")
		err = errors.New("expected RouterParms")
	}
	return err
}
