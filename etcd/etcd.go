package etcd

import (
	"context"
	"github.com/busgo/pink-go/log"
	"github.com/pkg/errors"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"sync"
)

type KeyChangeEvent int32

type KeyChangeChan <-chan *KeyChange

const (
	KeyCreateChangeEvent = iota + 1 //  create event
	KeyUpdateChangeEvent            //  update event
	KeyDeleteChangeEvent            //  delete event
	KeyCancelChangeEvent            //  cancel event
	defaultKeyChangeSize = 32
)

//  etcd cli
type Cli struct {
	c         *clientv3.Client
	kv        clientv3.KV
	lease     clientv3.Lease
	elections map[string]*concurrency.Election
	l         log.Logger
	sync.RWMutex
}

type WatchKeyResponse struct {
	Watcher     clientv3.Watcher
	KeyChangeCh chan *KeyChange
}

type KeyChange struct {
	Event KeyChangeEvent
	Key   string
	Value string
}

// new etcd cli
func NewEtcdCli(options ...Option) (*Cli, error) {
	cfg, err := newConfig(options...)
	if err != nil {
		return nil, err
	}
	if cfg.endpoints == nil || len(cfg.endpoints) == 0 {
		return nil, errors.New("endpoints can not be empty")
	}

	c, err := clientv3.New(clientv3.Config{
		Endpoints:   cfg.endpoints,
		Username:    cfg.userName,
		Password:    cfg.password,
		DialTimeout: cfg.dialTimeout,
	})

	if err != nil {
		return nil, err
	}
	if cfg.l == nil {
		cfg.l = &log.StdLogger{}
	}

	return &Cli{
		c:         c,
		kv:        clientv3.NewKV(c),
		lease:     clientv3.NewLease(c),
		elections: make(map[string]*concurrency.Election),
		l:         cfg.l,
	}, err
}

// get with key
func (cli *Cli) Get(ctx context.Context, key string) (string, error) {

	resp, err := cli.kv.Get(ctx, key)
	if err != nil {
		return "", err
	}

	if len(resp.Kvs) == 0 {
		return "", nil
	}
	return string(resp.Kvs[0].Value), err

}

// delete a key
func (cli *Cli) Delete(ctx context.Context, key string) error {
	_, err := cli.kv.Delete(ctx, key, clientv3.WithPrevKV())
	return err
}

// transfer a key
func (cli *Cli) Transfer(ctx context.Context, from, to string, value string) error {
	_, err := cli.c.Txn(ctx).Then(clientv3.OpDelete(from), clientv3.OpPut(to, value)).Else(clientv3.OpPut(to, value)).Commit()
	return err
}

// get with prefix
func (cli *Cli) GetWithPrefix(ctx context.Context, prefix string) ([]string, []string, error) {

	resp, err := cli.kv.Get(ctx, prefix, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		return make([]string, 0), make([]string, 0), err
	}

	if len(resp.Kvs) == 0 {
		return make([]string, 0), make([]string, 0), nil
	}

	keys := make([]string, 0)
	values := make([]string, 0)
	for _, kv := range resp.Kvs {

		keys = append(keys, string(kv.Key))
		values = append(values, string(kv.Value))
	}

	return keys, values, err
}

// put a key
func (cli *Cli) Put(ctx context.Context, key, value string) error {

	_, err := cli.kv.Put(ctx, key, value)
	return err
}

// put a key with ttl
func (cli *Cli) PutWithTTL(ctx context.Context, key, value string, ttl int64) (int64, error) {

	leaseResponse, err := cli.lease.Grant(ctx, ttl)
	if err != nil {
		return 0, err
	}
	_, err = cli.kv.Put(ctx, key, value, clientv3.WithLease(leaseResponse.ID))
	return int64(leaseResponse.ID), err
}

func (cli *Cli) PutWithNotExist(ctx context.Context, key, value string) error {
	tx := cli.c.Txn(ctx).If(clientv3.Compare(clientv3.Version(key), "=", 0)).
		Then(clientv3.OpPut(key, value))

	resp, err := tx.Commit()

	cli.l.Info(ctx, "resp:%+v", resp)

	return err
}

func (cli *Cli) PutWithNotExistTTL(ctx context.Context, key, value string, ttl int64) (int64, error) {
	leaseResponse, err := cli.lease.Grant(ctx, ttl)
	if err != nil {
		return 0, err
	}
	_, err = cli.c.Txn(ctx).If(clientv3.Compare(clientv3.Version(key), "=", 0)).
		Then(clientv3.OpPut(key, value, clientv3.WithLease(leaseResponse.ID))).
		Commit()
	return int64(leaseResponse.ID), nil
}

func (cli *Cli) Revoke(ctx context.Context, leaseId int64) error {

	if leaseId <= 0 {
		return nil
	}
	_, err := cli.lease.Revoke(ctx, clientv3.LeaseID(leaseId))
	return err
}

func (cli *Cli) Keepalive(ctx context.Context, key, value string, ttl int64) (int64, error) {
	resp, err := cli.lease.Grant(ctx, ttl)
	if err != nil {
		return 0, err
	}
	_, err = cli.kv.Put(ctx, key, value, clientv3.WithLease(resp.ID))
	if err != nil {
		return 0, err
	}

	// the key 'foo' will be kept forever
	ch, err := cli.lease.KeepAlive(context.Background(), resp.ID)
	if err != nil {
		return 0, err
	}
	go cli.keepaliveHandle(key, ch)
	return int64(resp.ID), nil
}

// handle keep alive
func (cli *Cli) keepaliveHandle(key string, ch <-chan *clientv3.LeaseKeepAliveResponse) {

	for {
		select {
		case c := <-ch:

			if c == nil {
				cli.l.Info(context.Background(), "the keep alive key:%s has closed", key)
				return
			}
			cli.l.Info(context.Background(), "keep alive for key:%s .................%+v", key, c)
		}
	}
}

func (cli *Cli) Watch(key string) *WatchKeyResponse {

	watcher := clientv3.NewWatcher(cli.c)
	watchChan := watcher.Watch(context.Background(), key)
	keyChangeCh := make(chan *KeyChange, defaultKeyChangeSize)

	// start watch
	go cli.keyChangeHandle(key, watchChan, keyChangeCh)
	return &WatchKeyResponse{
		Watcher:     watcher,
		KeyChangeCh: keyChangeCh,
	}

}

func (cli *Cli) WatchWithPrefix(prefix string) *WatchKeyResponse {

	watcher := clientv3.NewWatcher(cli.c)
	watchChan := watcher.Watch(context.Background(), prefix, clientv3.WithPrefix())

	keyChangeCh := make(chan *KeyChange, defaultKeyChangeSize)

	// start watch
	go cli.keyChangeHandle(prefix, watchChan, keyChangeCh)
	return &WatchKeyResponse{
		Watcher:     watcher,
		KeyChangeCh: keyChangeCh,
	}

}

func (cli *Cli) keyChangeHandle(prefix string, watchChan clientv3.WatchChan, keyChangeCh chan *KeyChange) {

	for {
		select {
		case ch, ok := <-watchChan:
			if !ok {
				cli.l.Info(context.Background(), "the watch prefix key:%s has cancel", prefix)
				keyChangeCh <- &KeyChange{
					Event: KeyCancelChangeEvent,
					Key:   prefix,
				}
				return
			}
			for _, event := range ch.Events {
				keyChangeEventHandle(event, keyChangeCh)
			}
		}

	}
}

func keyChangeEventHandle(event *clientv3.Event, ch chan *KeyChange) {

	c := &KeyChange{
		Key:   string(event.Kv.Key),
		Value: "",
	}
	switch event.Type {
	case mvccpb.PUT:
		c.Value = string(event.Kv.Value)
		c.Event = KeyCreateChangeEvent
		if event.IsModify() {
			c.Event = KeyUpdateChangeEvent
		}
	case mvccpb.DELETE:
		c.Event = KeyDeleteChangeEvent
	}
	ch <- c
}

// campaign become leader
func (cli *Cli) Campaign(ctx context.Context, id, prefix string, ttl int64) error {

	// create a session
	session, err := concurrency.NewSession(cli.c, concurrency.WithTTL(int(ttl)))
	if err != nil {
		cli.l.Info(ctx, "new session fail,id:%s,prefix:%s,%+v", id, prefix, err)
		return err
	}

	election := concurrency.NewElection(session, prefix)
	cli.elections[prefix] = election
	return election.Campaign(ctx, id)
}

func (cli *Cli) getElection(prefix string) (*concurrency.Election, error) {

	election := cli.elections[prefix]
	if election != nil {
		return election, nil
	}
	// create a session
	session, err := concurrency.NewSession(cli.c)
	if err != nil {
		cli.l.Info(context.Background(), "new session fail,prefix:%s,%+v", prefix, err)
		return nil, err
	}
	election = concurrency.NewElection(session, prefix)
	cli.elections[prefix] = election
	return election, nil
}

// find leader
func (cli *Cli) Leader(ctx context.Context, prefix string) (id string, err error) {

	election, err := cli.getElection(prefix)
	if err != nil {
		return
	}

	resp, err := election.Leader(ctx)
	if err != nil {
		return
	}
	return string(resp.Kvs[0].Value), nil

}
