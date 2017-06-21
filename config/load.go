package config

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/juju/errors"
	"github.com/zssky/log"

	"github.com/dearcode/crab/http"
)

var (
	domain = flag.String("domain", "", "tracker manager domain.")
	app    = flag.String("app", "", "tracker app name.")
	module = flag.String("module", "", "tracker module name.")
	old    []byte
)

const (
	httpTimeout = time.Second * 3
)

//LoadConfig load topics from manager server.
func LoadConfig() (string, error) {
	flag.Parse()

	if *domain == "" {
		return "", nil
	}

	log.Infof("domain:%v app:%v module:%v", *domain, *app, *module)
	url := fmt.Sprintf("http://%s/api/module/?APP=%s&Module=%s", *domain, *app, *module)
	buf, _, err := http.NewClient(httpTimeout).Get(url, nil, nil)
	if err != nil {
		log.Errorf("Get module error:%v, domain:%v", errors.ErrorStack(err), *domain)
		return "", err
	}

	if bytes.Equal(old, buf) {
		return "", nil
	}

	path := fmt.Sprintf("/tmp/%v_%v.yml", *app, *module)
	if err = ioutil.WriteFile(path, buf, 0644); err != nil {
		log.Errorf("write file %v, data:%s", errors.ErrorStack(err), buf)
		return "", err
	}

	old = buf

	log.Infof("new config:%v", path)

	return path, nil
}
