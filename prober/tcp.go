package prober

import (
	"fmt"
	"net"

	"github.com/shonenada/prober-server/status"
)

func TCPProbe(prober *Prober) {
	tcpSettings := prober.TCPSettings
	target := fmt.Sprintf("%s:%d", tcpSettings.Host, tcpSettings.Port)
	conn, err := net.Dial("tcp", target)
	if err != nil {
		if prober.Retry > status.Status.RetryTimes {
			status.Status.SetMessage(err.Error())
			status.Status.Retrying()
		} else {
			status.Status.SetMessage(err.Error())
			status.Status.Failed()
		}
		return
	}
	defer conn.Close()
	status.Status.FlushMessage()
	status.Status.Success()
	return
}
