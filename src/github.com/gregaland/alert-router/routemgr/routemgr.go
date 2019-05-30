package routemgr

import (
	"encoding/json"
	"fmt"
	"github.com/gregaland/alert-router/config"
	"github.com/gregaland/alert-router/routers"
	"github.com/pkg/errors"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"net/smtp"
	"os"
	"path/filepath"
)

// RouteMgr
type RouteMgr struct {
	config       *config.RigConfig
	auth         smtp.Auth
	alertRouters map[string]routers.Router
	alerts       map[string][]*ScheduledAlert
	cron         *cron.Cron
}

// ScheduleAlert object used for cron schedules
type ScheduledAlert struct {
	Config  config.RouterParms
	enabled bool
}

// Used to enable a schedule via the cron interface
type ScheduleEnabler struct {
	s *ScheduledAlert
}

// Used to disable a schedule via the cron interface
type ScheduleDisabler struct {
	s *ScheduledAlert
}

// Schedule enabler.  Implements the cron interface Run function to
// flip the enabled flag to true
func (s *ScheduleEnabler) Run() {
	log.Infof("Enabling an alert: %s", s.s.Config.RouterId)
	s.s.enabled = true
}

// Schedule enabler.  Implements the cron interface Run function to
// flip the enabled flag to false
func (s *ScheduleDisabler) Run() {
	log.Infof("Disabling an alert: %s", s.s.Config.RouterId)
	s.s.enabled = false
}

func NewRouteMgr(config *config.RigConfig) *RouteMgr {
	rm := &RouteMgr{config: config}
	rm.cron = cron.New()
	err := rm.initRouters()
	if err != nil {
		log.Fatal(err)
	}
	err = rm.loadAlerts()
	if err != nil {
		log.Fatal(err)
	}

	rm.cron.Start()

	return rm
}

func (rm *RouteMgr) Route(event *routers.Event) error {

	var err error = nil
	if schedule, ok := rm.alerts[event.Id]; ok {
		for _, s := range schedule {
			if s.enabled {
				log.WithFields(log.Fields{
					"router_id": s.Config.RouterId,
					"id":        s.Config.Id,
					"enabled":   s.enabled,
				}).Info("found configured alert")

				if route, ok := rm.alertRouters[s.Config.RouterId]; !ok {
					err = errors.New("No schedule for router with id: " + s.Config.RouterId)
				} else {
					log.Info("Firing " + event.Id + ": " + event.Message)
					routeEvent := &routers.Event{Id: event.Id, Message: event.Message}
					go route.Route(routeEvent, s.Config)
				}
			} else {
				log.Infof("alert disabled.  id: %s", s.Config.Id)
			}
		}
	} else {
		err = errors.New("No alerts with id: " + event.Id)
	}
	return err
}
// Private function that uses the main config file to initialize routers.
func (rm *RouteMgr) initRouters() error {
	var err error = nil
	var r routers.Router = nil
	rm.alertRouters = make(map[string]routers.Router)

	log.Debug("initializing routers")
	for _, router := range rm.config.Routers {

		switch router.Type {
		case config.EMAIL_RP:

			log.WithFields(log.Fields{
				"type":     router.Type,
				"smtphost": router.Parms.SmtpHost,
				"smtpport": router.Parms.SmtpPort,
				"au":       router.Parms.SmtpAuthUser,
				"ap":       router.Parms.SmtpAuthPass,
			}).Info("loading email router")

			c := &routers.EmailConfig{SmtpHost: router.Parms.SmtpHost, SmtpPort: router.Parms.SmtpPort,
				Au: router.Parms.SmtpAuthUser, Ap: router.Parms.SmtpAuthPass}
			r, err = routers.NewEmailRouter(c)
			break
		case config.WEBHOOK_RP:
			log.WithFields(log.Fields{
				"type": router.Type,
			}).Info("loading slack router")

			c := &routers.SlackConfig{Url: router.Parms.Url}
			r, err = routers.NewSlackRouter(c)
			break
		default:
			log.Fatal("Unknown router type")
			return errors.New("Unknown router type")
		}

		// initialize the router
		if r != nil {
			err = r.Init()
			if err != nil {
				log.Fatal(err)
			}
			rm.alertRouters[router.Parms.Id] = r
		}
	}
	return err
}

// Private function that loads the alerts from the alerts directory
func (rm *RouteMgr) loadAlerts() error {
	var err error = nil

	log.WithFields(log.Fields{
		"path": rm.config.AlertsPath,
	}).Debug("loading alerts")

	files, err := ioutil.ReadDir(rm.config.AlertsPath)
	if err != nil {
		log.Fatal(err)
	}

	rm.alerts = make(map[string][]*ScheduledAlert)

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".yml" {

			fullpath := fmt.Sprintf("%s%c%s", rm.config.AlertsPath, filepath.Separator, file.Name())
			log.WithFields(log.Fields{
				"file": file.Name(),
			}).Debug("loading alert config")

			alertData, err := os.Open(fullpath)
			if err != nil {
				log.Fatal(err)
			}
			alertConfig, err := config.LoadAlertConfig(alertData)
			if err != nil {
				log.Fatal(err)
			}
			log.WithFields(log.Fields{
				"alert_id":   alertConfig.AlertId,
				"parameters": alertConfig.Schedule,
			}).Info("loaded alert config")
			rm.AddAlertConfig(alertConfig)

		}
		fmt.Println(file.Name())
	}
	return err
}

func (rm *RouteMgr) AddAlert(alertId string, r *http.Request) (bool,error) {

	ac := config.AlertConfig{}

	// Already have it?
	if _, ok := rm.alerts[alertId]; ok {
		// already have it
		return false, nil
	}

	// try to parse as json
	err := json.NewDecoder(r.Body).Decode(&ac)
	if err != nil {
		log.Errorf("failed to parse json: %v", err)
	} else {
		rm.AddAlertConfig(&ac)
		out, err := yaml.Marshal(ac)
		err = ioutil.WriteFile(rm.config.AlertsPath + "/" + alertId + ".yml", out, 0644)
		if err != nil {
			log.Error(err)
		}
	}
	return true, err
}

// Add a alert configuration and create its schedule
func (rm *RouteMgr) AddAlertConfig(alertConfig *config.AlertConfig) {
	schedules := make([]*ScheduledAlert, 0)
	for _, sap := range alertConfig.Schedule {
		sa := ScheduledAlert{Config: sap, enabled: false}

		// TODO: if both start and end are not given - then what?

		if sa.Config.ScheduleStart != "" {
			s := &ScheduleEnabler{s: &sa}
			err := rm.cron.AddJob("0 "+ sa.Config.ScheduleStart, s)
			if err != nil {
				log.Error(err)
			}
		} else {
			sa.enabled = true
		}
		if sa.Config.ScheduleEnd != "" {
			d := &ScheduleDisabler{s: &sa}
			err := rm.cron.AddJob("0 "+sa.Config.ScheduleEnd, d)
			if err != nil {
				log.Error(err)
			}
		}
		schedules = append(schedules, &sa)
		log.WithFields(log.Fields{
			"enabled": sa.enabled,
			"params":  sa.Config,
		}).Info("alert scheduled")
	}
	rm.alerts[alertConfig.AlertId] = schedules
}

// Private function to delete an alert
func (rm *RouteMgr) DeleteAlert(alertId string) (bool,error) {

	// do we have it?
	if _, ok := rm.alerts[alertId]; !ok {
		// nope
		return false, nil
	}

	delete(rm.alerts, alertId)

	// TODO use config path
	err := os.Remove(rm.config.AlertsPath + "/" + alertId + ".yml")
	return true, err
}

func (rm *RouteMgr) GetAlerts() map[string][]*ScheduledAlert {
	// returns a copy of the alerts
	result := make(map[string][]*ScheduledAlert)
	for k,v := range rm.alerts {
		sa := make([]*ScheduledAlert, 0, len(v))
		for _, alert := range v {
			sa = append(sa, alert)
		}
		result[k] = sa
	}
	return result
}
