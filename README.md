# Prober Server

Run your server in your server

```sh
$ export PROBER_SERVER_PORT="9078"
$ export PROBER_TYPE="HTTP"
$ export PROBER_DURATION="5s"
$ export PROBER_HTTP_URL="http://example.com:5000/healthcheck"
$ export PROBER_HTTP_RETRY="5"
$ go run main.go
```

Then you can get the status by HTTP API:

```sh
$ curl http://localhost:9078
{"code":1,"status":"RETRYING","message":"Get \"http://example.com:5000/healthcheck\": dial tcp 93.184.216.34:5000: i/o timeout","retry_time":1,"last_updated":"2020-10-23T06:35:26.304172Z"}
```
