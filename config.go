package client

import (
	"github.com/busgo/pink-go/etcd"
	"github.com/busgo/pink-go/log"
)

//go:generate go-options config
// pink client config
type config struct {
	cli        *etcd.Cli
	group      string
	l          log.Logger
	ip         string
	extensions []Extension
}
