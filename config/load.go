package config

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/dearcode/libbeat/logp"
	"golang.org/x/net/context"
)

var (
	etcdApp  = flag.String("app", "", "etcd app key.")
	etcdAddr = flag.String("etcd", "", "etcd addr list.")

	client  *clientv3.Client
	version int64
)

func init() {
	flag.Parse()

	if *etcdAddr != "" {
		c, err := clientv3.New(clientv3.Config{
			Endpoints:   strings.Split(*etcdAddr, ","),
			DialTimeout: etcdTimeout,
		})
		if err != nil {
			panic(err)
		}
		client = c
	}
}

const (
	etcdTimeout = time.Second * 3
)

//LoadConfig load config file from etcd.
func LoadConfig() (bool, error) {
	if client == nil {
		return false, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), etcdTimeout)
	resp, err := clientv3.NewKV(client).Get(ctx, *etcdApp)
	cancel()

	if err != nil {
		panic(err)
	}

	if len(resp.Kvs) == 0 {
		return true, fmt.Errorf("%v not found in etcd", *etcdApp)
	}

	log.Printf("local version:%v, modRevision:%v", version, resp.Kvs[0].ModRevision)

	if version != 0 && version == resp.Kvs[0].ModRevision {
		return false, nil
	}

	version = resp.Kvs[0].ModRevision

	log.Printf("config version:%v", version)
	logp.Info("config version:%v", version)

	return true, ioutil.WriteFile("./filebeat.yml", resp.Kvs[0].Value, 0644)
}
