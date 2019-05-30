package config

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
)

type RouteProcessor string

const (
	EMAIL_RP   RouteProcessor = "email"
	WEBHOOK_RP RouteProcessor = "webhook"
)

type Email RouteProcessor
type Webhook RouteProcessor

type AlertConfig struct {
	AlertId  string        `yaml:"alert" json:"alert"`
	Schedule []RouterParms `yaml:"schedule" json:"schedule"`
}

type RouterParms struct {
	Id            string   `yaml:"id" json:"id"`
	RouterId      string   `yaml:"router_id,omitempty" json:"router_id,omitempty"`
	Enabled       bool     `yaml:"enabled" json:"enabled,omitempty"`
	SmtpHost      string   `yaml:"smtphost,omitempty" json:"smtphost,omitempty"`
	SmtpPort      int      `yaml:"smtpport,omitempty" json:"smtpport,omitempty"`
	SmtpAuthUser  string   `yaml:"smtpauthuser,omitempty" json:"smtpauthuser,omitempty"`
	SmtpAuthPass  string   `yaml:"smtpauthpass,omitempty" json:"smtpauthpass,omitempty"`
	EmailAddrs    []string `yaml:"email_addrs,omitempty" json:"email_addrs,omitempty"`
	Url           string   `yaml:"url,omitempty" json:"url,omitempty"`
	Username      string   `yaml:"username,omitempty" json:"username,omitempty"`
	Password      string   `yaml:"password,omitempty" json:"password,omitempty"`
	QueryParms    []string `yaml:"query_parms,omitempty" json:"query_parms,omitempty"`
	ScheduleStart string   `yaml:"start,omitempty" json:"start,omitempty"`
	ScheduleEnd   string   `yaml:"end,omitempty" json:"end,omitempty"`
}

type Routers struct {
	Type  RouteProcessor `yaml:"type"`
	Parms RouterParms    `yaml:",inline"`
}

type RigConfig struct {
	Listen       string     `yaml:"listen"`
	Routers      []*Routers `yaml:"routers"`
	AlertsPath   string     `yaml:"alerts_path"`
	LogLevelStr  string     `yaml:"log_level"`
	LogFormatStr string     `yaml:"log_format"`
}

func (rc *RigConfig) loadEnvVars() (map[string]string, error) {

	var err error
	var envs = make(map[string]string)
	if au, ok := os.LookupEnv("SMTP_AUTH_USER"); ok {
		log.Debugf("SMTP_AUTH_USER: %s", au)
		envs["SMTP_AUTH_USER"] = au
	} else {
		err = errors.New("SMTP_AUTH_USER must be set")
	}
	if ap, ok := os.LookupEnv("SMTP_AUTH_PASS"); ok {
		log.Debugf("SMTP_AUTH_USER: %s", ap)
		envs["SMTP_AUTH_PASS"] = ap
	} else {
		err = errors.New("SMTP_AUTH_USER must be set")
	}

	return envs, err

}
func LoadAlertConfig(r io.Reader) (*AlertConfig, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	log.Debugf("received: %s", string(data))
	alertConfig := &AlertConfig{}
	err = yaml.Unmarshal(data, &alertConfig)
	if err != nil {
		return nil, err
	}
	return alertConfig, err

}

// NewRigConfig returns a new RigConfig instance.
func NewRigConfig(r io.Reader) (*RigConfig, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	rigConfig := &RigConfig{}

	err = yaml.Unmarshal(data, &rigConfig)
	if err != nil {
		return nil, err
	}
	envs, err := rigConfig.loadEnvVars()
	if err == nil {
		for _, c := range rigConfig.Routers {
			if EMAIL_RP == c.Type {
				c.Parms.SmtpAuthUser = envs["SMTP_AUTH_USER"]
				c.Parms.SmtpAuthPass = envs["SMTP_AUTH_PASS"]
			}
		}
	}

	return rigConfig, err
}

func (rc *RigConfig) LogLevel() log.Level {

	var level log.Level

	switch rc.LogLevelStr {
	case "debug":
		level = log.DebugLevel
		break
	case "info":
		level = log.InfoLevel
		break
	case "warning":
		level = log.WarnLevel
		break
	case "error":
		level = log.ErrorLevel
		break
	case "fatal":
		level = log.FatalLevel
		break
	default:
		level = log.InfoLevel
		break
	}

	return level
}

func (rc *RigConfig) LogFormat() log.Formatter {

	var format log.Formatter

	switch rc.LogLevelStr {
	case "text":
		format = &log.TextFormatter{}
		break
	case "json":
		format = &log.JSONFormatter{}
		break
	default:
		format = &log.JSONFormatter{}
		break
	}
	return format
}
