# Prober Server

Prober Server probe your services deploy in kubernetes as a sidecar.

## Run in your host

```sh
$ export PROBER_SERVER_PORT="9078"
$ export PROBER_DURATION="5s"
$ export PROBER_RETRY="5"
$ export PROBER_TYPE="HTTP"
$ export PROBER_HTTP_URL="http://example.com/healthcheck"
$ go run main.go
2020/10/23 14:34:56 Env `PROBER_HTTP_TIMEOUT` is not set, using default value 30
2020/10/23 14:34:56 Prober server running in `HTTP` type; Probe Duration: 5s
2020/10/23 14:34:56 HTTP URL: http://example.com:5000/healthcheck; HTTP Timeout: 30
2020/10/23 14:35:26 STATUS: RETRYING - 2020-10-23 06:35:26.304173 +0000 UTC
2020/10/23 14:36:01 STATUS: RETRYING - 2020-10-23 06:36:01.312028 +0000 UTC
2020/10/23 14:36:36 STATUS: RETRYING - 2020-10-23 06:36:36.31496 +0000 UTC
2020/10/23 14:37:11 STATUS: FAILED - 2020-10-23 06:37:11.320504 +0000 UTC
```

Then you can get the status by HTTP API:

```sh
$ curl http://localhost:9078
{"code":0,"status":"SUCCESS","message":"","retry_time":0,"last_updated":"2020-10-23T06:32:49.779909Z"}

$ curl http://localhost:9078
{"code":1,"status":"RETRYING","message":"Get \"http://example.com:5000/healthcheck\": dial tcp 93.184.216.34:5000: i/o timeout","retry_time":1,"last_updated":"2020-10-23T06:35:26.304172Z"}

$ curl http://localhost:9078
{"code":2,"status":"FAILED","message":"Get \"http://example.com:5000/healthcheck\": dial tcp 93.184.216.34:5000: i/o timeout","retry_time":4,"last_updated":"2020-10-23T06:39:17.679375Z"}
```

## Using docker

```sh
$ docker build --tag prober-server:latest .
$ docker run -tid --name prober-server -p 9078:9078 \
  -e PROBER_TYPE=http \
  -e PROBER_DURATION=5s \
  -e PROBER_RETRY=5 \
  -e PROBER_HTTP_URL="http://example.com/healthcheck" \
  prober-server:latest
```

## TCP and UDP

You can also probe TCP service by set `PROBER_TYPE` environment variables:

```sh
$ export PROBER_SERVER_PORT="9078"
$ export PROBER_DURATION="5s"
$ export PROBER_RETRY="5"
$ export PROBER_TYPE="TCP"    # this changed
$ export PROBER_TCP_HOST="example.com"
$ export PROBER_TCP_PORT="8080"
$ go run main.go
```

And UDP service by:

```sh
$ export PROBER_SERVER_PORT="9078"
$ export PROBER_DURATION="5s"
$ export PROBER_RETRY="5"
$ export PROBER_TYPE="UDP"    # this changed
$ export PROBER_UDP_HOST="example.com"
$ export PROBER_UDP_PORT="4093"
$ go run main.go
```

## Adding Webhook

You can setup webhook URL for recieving result of prober. To achieve webhook
just setup `PROBER_NAME` and `PROBER_WEBHOOK` environment variables.

NOTE: `PROBER_NAME` is set for identity your service.


```sh
$ export PROBER_NAME="YOUR_SERVER"
$ export PROBER_WEBHOOK="http://your-webhook.com"
```

Then you will recieve a POST Request with following payload:

```json
{
  "code": 1,
  "last_updated": "2020-10-23T10:21:05.312644Z",
  "message": "Get \"http://localhost:80\": dial tcp [::1]:80: connect: connection refused",
  "name": "NGINX",
  "retry_times": 1,
  "status": "RETRYING"
}
```
