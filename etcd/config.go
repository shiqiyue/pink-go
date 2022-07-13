package etcd

import (
	"github.com/busgo/pink-go/log"
	"time"
)

//go:generate go-options config
// etcd cli config
type config struct {
	endpoints   []string
	userName    string
	password    string
	dialTimeout time.Duration
	l           log.Logger
}
