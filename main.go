package main

import (
	"fmt"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/dearcode/filebeat/beater"
	"github.com/dearcode/filebeat/config"
	"github.com/dearcode/libbeat/beat"
)

var Name = "filebeat"

func fork() int {
	attr := syscall.ProcAttr{
		Env:   os.Environ(),
		Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()},
	}

	pid, err := syscall.ForkExec(os.Args[0], os.Args[:1], &attr)
	if err != nil {
		panic(err)
	}

	log.Printf("new child pid:%v", pid)
	return pid
}

func watcher() {
	t := time.NewTicker(time.Second * 3)
	pid := fork()

	for {
		<-t.C

		load, err := config.LoadConfig()
		if err != nil {
			log.Printf("load config error:%v", err)
			continue
		}

		if !load {
			log.Printf("no change")
			continue
		}

		syscall.Kill(pid, syscall.SIGHUP)

		var wstatus syscall.WaitStatus

		if _, err := syscall.Wait4(pid, &wstatus, 0, nil); err != nil {
			panic(err)
		}
		fmt.Printf("Wait4 pid:%v, status:%v\n", pid, wstatus)

		pid = fork()
	}
}

// The basic model of execution:
// - prospector: finds files in paths/globs to harvest, starts harvesters
// - harvester: reads a file, sends events to the spooler
// - spooler: buffers events until ready to flush to the publisher
// - publisher: writes to the network, notifies registrar
// - registrar: records positions of files read
// Finally, prospector uses the registrar information, on restart, to
// determine where in each file to restart a harvester.

func main() {
	load, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	if load {
		watcher()
	}

	if err := beat.Run(Name, "", beater.New); err != nil {
		os.Exit(1)
	}
}
