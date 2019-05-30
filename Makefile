APP_VERSION := $(shell git describe --always --long --dirty)
RPM_VERSION := $(shell git describe --always)
OUT := alert-router
PKG := github.com/gregaland/alert-router

# ---------------------------------------------
# Nothing below this line should need to change
# ---------------------------------------------
# Set the GOPATH to the directory where the Makefile lives
export GOPATH=$(MAKEFILE_DIR)


# Spec file naming pattern is <appname>.spec
# Attempt to discover spec file - if multiple files match the pattern
# the result is indeterminate
ifneq ("$(wildcard *.spec)","")
  RPM_SPEC_FILE ?= $(wildcard *.spec)
  RPM_NAME=$(RPM_SPEC_FILE:.spec=)
  RPM_NAME_VERSION=${RPM_NAME}-${RPM_VERSION}
else
  $(error "Failed to find the spec file. Spec files should follow the pattern <app-name>.spec")
endif

RPMBUILD=/usr/bin/rpmbuild
TAR=/usr/bin/tar
SED=/usr/bin/sed

# Working files/dirs
MAKEFILE_DIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
RPM_DIR=$(MAKEFILE_DIR)/rpm
RPM_SRC_DIR=${RPM_DIR}/SOURCES
RPM_TAR_DIR=${RPM_SRC_DIR}/${RPM_NAME_VERSION}
RPM_TAR_FILE=${RPM_SRC_DIR}/${RPM_NAME_VERSION}.tar.gz

build: go_vars_check fmt bin

rpm: vars_check prep
	@${RPMBUILD} -D '%_topdir $(RPM_DIR)' -D '%version $(RPM_VERSION)' --clean --target=x86_64 -ta ${RPM_TAR_FILE}

docker: 
	docker build --build-arg APP_VERSION=$(APP_VERSION) . -t gregaland/alert-router:$(RPM_VERSION)

prep:
	@rm -rf ${RPM_DIR}
	@mkdir -p $(RPM_DIR) ${RPM_TAR_DIR}
	@rsync -a * ${RPM_TAR_DIR} --exclude rpm
	@tar cvfz ${RPM_TAR_FILE} -C ${RPM_SRC_DIR} ${RPM_NAME_VERSION}

fmt:
	go fmt ${PKG}

deps:
	go get github.com/gorilla/mux
	go get gopkg.in/yaml.v2
	go get github.com/sirupsen/logrus
	go get gotest.tools/assert
#	go get github.com/stretchr/testify/assert
	go get github.com/robfig/cron

bin: deps
	go build -i -v -o ./bin/${OUT} -ldflags="-X main.version=${APP_VERSION}" ${PKG}

go_vars_check: 
	@if [ ${GOPATH} = "NOT_SET" ] ; then \
		echo "Your GOPATH is not set."; \
		exit 1; \
	fi;

vars_check:
	@if [ ${RPM_VERSION} = "NOT_SET" ] ; then \
		echo "RPM_VERSION not set, check /vob/make/cfg/rpm.mk for required variables"; \
		exit 1; \
	fi;

clean:
	go clean
	@rm -rf ${GOPATH}/pkg
	@rm -rf ${GOPATH}/bin/alert-router
	@rm -rf ${GOPATH}/src/github.com/gorilla/mux
	@rm -rf ${GOPATH}/src/github.com/sirupsen/logrus
	@rm -rf ${GOPATH}/src/github.com/stretchr/testify/assert
	@rm -rf ${GOPATH}/src/github.com/robfig/cron
	@rm -rf ${GOPATH}/src/gopkg.in
	@rm -rf ${GOPATH}/src/golang.org
	@rm -rf ${GOPATH}/src/gotest.tools
	@rm -rf ${RPM_DIR}
	@rm -rf ./bin/${OUT}
