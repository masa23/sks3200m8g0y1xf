package main

import (
	"fmt"

	"github.com/masa23/sks3200m8g0y1xf"
)

func main() {
	client := sks3200m8g0y1xf.NewClient("http://192.168.10.14")

	if err := client.Login("admin", "admin"); err != nil {
		panic(err)
	}

	ports, err := client.GetMonitoringPortStatics()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%5s%8s%12s%12s%12s%12s%12s\n",
		"Port", "State", "LinkStatus", "TxGoodPkt", "TxBadPkt", "RxGoodPkt", "RxBadPkt")
	for _, port := range ports {
		fmt.Printf("%5d%8s%12s%12d%12d%12d%12d\n",
			port.PortNumber, port.State, port.LinkStatus, port.TxGoodPkt, port.TxBadPkt, port.RxGoodPkt, port.RxBadPkt)
	}

	if err := client.Logout(); err != nil {
		panic(err)
	}
}
