package main

/*
alert-router will route an alert to one or more email addresses.  Alerts are routed based
on Alert ID.  Alert messages must be under 140 bytes.
*/

import (
	"flag"
	"fmt"
	"github.com/gregaland/alert-router/api"
	"github.com/gregaland/alert-router/config"
	"github.com/gregaland/alert-router/routemgr"
	log "github.com/sirupsen/logrus"
	"os"
)

// Version filled in by makefile from git tags
var version = "undefined"

func main() {
	fmt.Println("Starting. Version: " + version)
	var configFilePtr = flag.String("c", "/etc/alert-router.yml", "Path to configuration file")
	flag.Parse()

	// load configuration
	configData, err := os.Open(*configFilePtr)
	if err != nil {
		log.Fatal(err)
	}
	rigConfig, err := config.NewRigConfig(configData)
	if err != nil {
		log.Fatal(err)
	}

	log.SetLevel(rigConfig.LogLevel())
	log.SetFormatter(rigConfig.LogFormat())
	log.SetOutput(os.Stdout)
	log.SetReportCaller(true)

	routeMgr := routemgr.NewRouteMgr(rigConfig)
	alert := api.NewAlertApi(rigConfig, routeMgr)
	alert.ListenAndServe()
}
