package main

import (
	"flag"
	"os"
	"syscall"
	"time"

	"github.com/dearcode/libbeat/beat"
	"github.com/juju/errors"
	"github.com/zssky/log"

	"github.com/dearcode/filebeat/beater"
	"github.com/dearcode/filebeat/config"
)

var Name = "filebeat"

const (
	watcherTimeout = time.Minute
)

func fork(path string) int {
	attr := syscall.ProcAttr{
		Env:   os.Environ(),
		Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()},
	}

	argv := []string{os.Args[0], "-c", path}
	if home := flag.Lookup("path.home"); home.Value.String() != "" {
		argv = append(argv, "-path.home")
		argv = append(argv, home.Value.String())
	}

	pid, err := syscall.ForkExec(os.Args[0], argv, &attr)
	if err != nil {
		panic(err)
	}

	log.Infof("new child pid:%v, argv:%v", pid, argv)

	return pid
}

func watcher(path string) {
	if path == "" {
		return
	}

	pid := fork(path)

	for range time.NewTicker(watcherTimeout).C {
		path, err := config.LoadConfig()
		if err != nil {
			log.Infof("load config %v", err)
			continue
		}

		if path == "" {
			log.Infof("the configuration file has not changed")
			continue
		}

		syscall.Kill(pid, syscall.SIGHUP)

		var wstatus syscall.WaitStatus

		log.Debugf("Wait4 pid:%v", pid)
		if _, err := syscall.Wait4(pid, &wstatus, 0, nil); err != nil {
			panic(err)
		}
		log.Debugf("pid:%v, status:%v", pid, wstatus)

		pid = fork(path)
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
	flag.Parse()
	path, err := config.LoadConfig()
	if err != nil {
		panic(errors.ErrorStack(err))
	}

	watcher(path)

	log.Debugf("argv:%#v", os.Args)
	if err = beat.Run(Name, "", beater.New); err != nil {
		os.Exit(1)
	}
}
