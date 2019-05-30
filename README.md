# Alert Router

Alert router is a work in process.  The high level idea was to implement an alerting API that could be used to route critical alerts based on a schedule.  For example, we may always want a database failure to notify a team via a slack post, however, during off hours, we may also want the alert to be routed to an SMS an/or email address as well.

## Example

Given the following alert-router configuration:

```
listen: :8000
alerts_path: etc/alerts.d
log_level: info
log_format: json
routers:
  - id: gmail
    type: email
    enabled: true
    smtphost: smtp.gmail.com
    smtpport: 587
  - id: slack-alerts
    type: webhook
    url: https://hooks.slack.com/services/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
    enabled: true
```

AND

Given the following alert (example.json):

```
{
 "alert": "dbfail",
 "schedule": [
  {
    "id": "all_day",
    "router_id": "slack-alerts",
  },
  {
    "id": "after_hours",
    "start": "0 17 * * *",
    "end": "0 6 * * *",
    "router_id": "gmail",
    "email_addrs": ["8885551234@tmomail.net", "john.doe@gmail.com"] 
  }]
}
```

Add Alert Config:

> curl -d@./example.json http://alert-router/v1/alert/dbfail

Update Alert Config:

> curl -X PUT -d@./example.json http://alert-router/v1/alert/dbfail

Delete Alert Config:

> curl -X DELETE -d@./example.json http://alert-router/v1/alert/dbfail

Fire Alert:

> curl -d '{"msg": "db is down"}' http://alert-router/v1/alert/dbfail/fire

List Alerts

> curl http://alert-router/v1/alerts
