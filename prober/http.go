package prober

import (
	"fmt"
	"net/http"

	"github.com/shonenada/prober-server/status"
)

func HTTPProbe(prober *Prober) {
	httpSettings := prober.HTTPSettings
	resp, err := http.Get(httpSettings.URL)
	if err != nil {
		if prober.Retry > status.Status.RetryTime {
			status.Status.SetMessage(err.Error())
			status.Status.Retrying()
		} else {
			status.Status.SetMessage(err.Error())
			status.Status.Failed()
		}
		return
	}

	defer resp.Body.Close()

	if 200 <= resp.StatusCode && resp.StatusCode <= 299 {
		status.Status.FlushMessage()
		status.Status.Success()
		return
	} else {
		if prober.Retry > status.Status.RetryTime {
			status.Status.SetMessage(fmt.Sprintf("Response statusCode %d is not range of 200 ~ 299 ", resp.StatusCode))
			status.Status.Retrying()
		} else {
			status.Status.SetMessage(fmt.Sprintf("Response statusCode %d is not range of 200 ~ 299 ", resp.StatusCode))
			status.Status.Failed()
		}
		return
	}
}
