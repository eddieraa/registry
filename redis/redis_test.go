//go:build integration
// +build integration

package redis

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

func Test1(t *testing.T) {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	rdb.Set(ctx, "toto", "titi", time.Hour)
	res := rdb.Get(ctx, "toto")
	println(res.Val())
	pb := rdb.Subscribe(ctx, "channel")
	pb.Subscribe(ctx, "channel2")
	ch := pb.Channel()

	tk := time.NewTicker(time.Second * 20)

	select {
	case <-tk.C:
		println("Timed out")
		break
	case m := <-ch:
		fmt.Printf("Rcv channel: %s payload: %s\n", m.Channel, m.Payload)
		break
	}

}
