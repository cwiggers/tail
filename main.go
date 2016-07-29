package main

import (
	"flag"

	"github.com/cwiggers/tail/master"
	"github.com/cwiggers/tail/slave"
)

func main() {

	masterAddr := flag.String("m", "127.0.0.1:6678", "master server address")
	slaveAddr := flag.String("s", "127.0.0.1:6679", "slave server address")
	token := flag.String("t", "", "token")
	ddir := flag.String("d", "data", "data store")

	flag.Parse()

	if *masterAddr != "" && *slaveAddr != "" {
		// slave server
		slave.InitSlave(*ddir, *masterAddr, *slaveAddr, *token)
	} else if *masterAddr != "" {
		// master server
		master.InitMaster(*masterAddr, *token)
	}
}
