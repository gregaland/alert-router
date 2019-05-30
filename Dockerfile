FROM golang:1.12.5-alpine3.9 AS build-env
ARG APP_VERSION=v0.0.7
RUN apk update && \
apk upgrade && \
apk add git

COPY src /go/src
RUN cd /go/src && \
go get github.com/gorilla/mux && \
go get gopkg.in/yaml.v2 && \
go get github.com/sirupsen/logrus && \
go get gotest.tools/assert && \
go get github.com/robfig/cron && \
go build -i -v -o ./bin/alert-router -ldflags="-X main.version=$APP_VERSION" github.com/gregaland/alert-router

FROM alpine
WORKDIR /app
COPY --from=build-env /go/src/bin/alert-router /app/
COPY etc/alert-router.yml /app
COPY etc/alerts.d /app/alerts.d
RUN apk update \
        && apk upgrade \
        && apk add --no-cache \
        ca-certificates \
        && update-ca-certificates 2>/dev/null || true
EXPOSE 8000
CMD ["/app/alert-router", "-c", "/app/alert-router.yml"]
