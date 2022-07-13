package client

import (
	"context"
	"fmt"
	"github.com/busgo/pink-go/etcd"
	"log"
	"math/rand"
	"testing"
	"time"
)

type logExtension struct {
}

func (l logExtension) Before(ctx context.Context, task Job) context.Context {
	fmt.Println(task.Target(), "开始执行")
	return ctx
}

func (l logExtension) After(ctx context.Context, task Job) context.Context {
	fmt.Println(task.Target(), "执行结束")
	return ctx
}

func TestNewPinkClient(t *testing.T) {

	e, _ := etcd.NewEtcdCli(etcd.OptionDialTimeout(time.Second*5), etcd.OptionEndpoints([]string{"10.10.20.162:2379"}))
	client, _ := NewPinkClient(OptionCli(e), OptionGroup("xd-design"), OptionExtensions([]Extension{&logExtension{}}), OptionIp("10.10.20.89"))
	client.Subscribe(&UserJob{})

	for {
		//t.Logf("client:%+v", client)
		time.Sleep(time.Second)
	}
}

type UserJob struct {
}

func (u *UserJob) Target() string {
	return "irm.invalidate_expire_idea_auth"
}
func (u *UserJob) Execute(ctx context.Context, param string) (result string, err error) {

	log.Printf("receive param:%s", param)
	n := rand.Int63() % 20
	time.Sleep(time.Second * time.Duration(n))
	return "调用成功", err
}
