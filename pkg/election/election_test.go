package election

import (
	"fmt"
	"log"
	"sync"
	"testing"
	"time"
	"xdp-banner/pkg/etcd"
	"xdp-banner/pkg/node"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestElection(t *testing.T) {
	cli, err := etcd.New(clientv3.Config{
		Endpoints: []string{"10.43.4.33:2379"},
		Username:  "root",
		Password:  "0FJomdKTgi",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	name, err := node.Name()
	if err != nil {
		log.Fatal(err)
	}
	ip, err := node.DefaultIP()
	if err != nil {
		log.Fatal(err)
	}

	ni1 := NodeInfo{
		Name:       name + "-1",
		ListenAddr: ip.String(),
	}
	if err != nil {
		log.Fatal(err)
	}
	ni2 := NodeInfo{
		Name:       name + "-2",
		ListenAddr: ni1.ListenAddr,
	}

	cctx := t.Context()

	e1, err := New(cctx, cli, ni1)
	if err != nil {
		log.Fatal(err)
	}

	e2, err := New(cctx, cli, ni2)
	if err != nil {
		log.Fatal(err)
	}

	msg := e1.Subscribe("e1")
	go func() {
		for {
			select {
			case msg := <-msg:
				fmt.Printf("e1 received: %s\n", msg)
			case <-cctx.Done():
				return
			}
		}
	}()

	// create competing candidates, with e1 initially losing to e2
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		e1.Campaign()
		fmt.Println("e1 completed first election")
		time.Sleep(6 * time.Second)
		fmt.Println("e1 resigning")
		e1.Resign(false)
	}()

	go func() {
		defer wg.Done()
		// delay candidacy so e1 wins first
		time.Sleep(3 * time.Second)
		e2.Campaign()
		fmt.Println("e2 completed first election")
	}()

	wg.Wait()
	time.Sleep(3 * time.Second)
}
