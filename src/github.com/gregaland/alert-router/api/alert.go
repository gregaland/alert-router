package api

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gregaland/alert-router/config"
	"github.com/gregaland/alert-router/routemgr"
	"github.com/gregaland/alert-router/routers"
	log "github.com/sirupsen/logrus"
	"net/http"
)

// API payload
type Event struct {
	Message string `json:"msg,omitempty"`
}

// RigAlert
type AlertApi struct {
	config       *config.RigConfig
	routeMgr     *routemgr.RouteMgr
	router       *mux.Router
}

// NewRigAlert returns a new instance
func NewAlertApi(config *config.RigConfig, mgr *routemgr.RouteMgr) *AlertApi {
	alertApi := &AlertApi{}
	alertApi.config = config
	alertApi.routeMgr = mgr

	alertApi.router = mux.NewRouter()
	alertApi.router.HandleFunc("/v1/alerts/{id}/fire", alertApi.SendAlert).Methods("POST")
	alertApi.router.HandleFunc("/v1/alerts/{id}", alertApi.AddAlert).Methods("POST")
	alertApi.router.HandleFunc("/v1/alerts/{id}", alertApi.UpdateAlert).Methods("PUT")
	alertApi.router.HandleFunc("/v1/alerts/{id}", alertApi.DeleteAlert).Methods("DELETE")
	alertApi.router.HandleFunc("/v1/alerts", alertApi.ListAlerts).Methods("GET")
	alertApi.router.HandleFunc("/v1/ekg", alertApi.Ekg).Methods("GET")

	return alertApi
}

// Listen and Serve until process terminated
func (aa *AlertApi) ListenAndServe() {
	log.Fatal(http.ListenAndServe(aa.config.Listen, aa.router))
}

// API Endpoint: /v1/alerts/{id}/fire
//
func (aa *AlertApi) SendAlert(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var event Event
	_ = json.NewDecoder(r.Body).Decode(&event)
	alertId := params["id"]

	err := aa.routeMgr.Route(&routers.Event{Id: alertId, Message: event.Message})
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	}
}



// Add an Alert id
//
// API Endpoint: POST /v1/alerts/{id}
//
func (aa *AlertApi) AddAlert(w http.ResponseWriter, r *http.Request) {
	alertId := mux.Vars(r)["id"]
	log.Infof("adding alert: %s", alertId)
    found, err := aa.routeMgr.AddAlert(alertId, r)
    if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusBadRequest)
	} else if !found {
		w.WriteHeader(http.StatusNotFound)
	}
}

// Update an Alert ID
// API Endpoint: PUT /v1/alerts/{id}
//
func (aa *AlertApi) UpdateAlert(w http.ResponseWriter, r *http.Request) {

	alertId := mux.Vars(r)["id"]
	log.Infof("updating alert: %s", alertId)
	found, err := aa.routeMgr.DeleteAlert(alertId)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
	} else if !found {
		w.WriteHeader(http.StatusNotFound)
	} else {
		_, err = aa.routeMgr.AddAlert(alertId, r)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

// Delete an Alert ID
// API Endpoint: DELETE /v1/alerts/{id}
//
func (aa *AlertApi) DeleteAlert(w http.ResponseWriter, r *http.Request) {

	alertId := mux.Vars(r)["id"]
	log.Infof("deleting alert: %s", alertId)

	found, err := aa.routeMgr.DeleteAlert(alertId)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
	} else if !found {
		w.WriteHeader(http.StatusNotFound)
	}
}

// Delete an Alert ID
// API Endpoint: DELETE /v1/alerts/{id}
//
func (aa *AlertApi) ListAlerts(w http.ResponseWriter, r *http.Request) {
	ac := make([]config.AlertConfig, 0)

	alerts := aa.routeMgr.GetAlerts()
	for k, v := range alerts {
		rp := make([]config.RouterParms, 0)
		for _, sa := range v {
			rp = append(rp, sa.Config)
		}
		ac = append(ac, config.AlertConfig{AlertId: k, Schedule: rp})
	}
	body, err := json.Marshal(ac)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		_, err = w.Write(body)
	}
}

// API Endpoint: /ekg
//
func (aa *AlertApi) Ekg(w http.ResponseWriter, r *http.Request) {
	_, err := fmt.Fprintf(w, "OK")
	if err != nil {
		log.Error(err)
	}
}
