package prober

import (
	"fmt"
	"net"

	"github.com/shonenada/prober-server/status"
)

func UDPProbe(prober *Prober) {
	udpSettings := prober.UDPSettings
	target := fmt.Sprintf("%s:%d", udpSettings.Host, udpSettings.Port)
	conn, err := net.Dial("udp", target)
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
}
