package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/labstack/gommon/log"
	"github.com/redis/go-redis/v9"
	"github.com/tetesuzawa/redisloadtest"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

// 90 days
func main() {
	host := flag.String("host", "127.0.0.1", "host")
	port := flag.Uint("port", 6379, "host")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())

	ch := make(chan int, 130_000_000)

	for j := 0; j < 130_000_000; j++ {
		if j%1000 == 0 {
			fmt.Println("inserting... ", j)
		}
		ch <- j
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	eg, ctx := errgroup.WithContext(ctx)

	for n := 0; n < runtime.NumCPU(); n++ {
		eg.Go(func() error {
			rdb := redis.NewClusterClient(&redis.ClusterOptions{
				Addrs:    []string{fmt.Sprintf("%s:%d", *host, *port)},
				Password: "", // no password set
			})
			defer rdb.Close()
			if err := rdb.Ping(ctx).Err(); err != nil {
				log.Print("failed to ping redis:", err)
				cancel()
			}

			for {
				select {
				case <-ctx.Done():
					fmt.Println("Stopping due to context done.")
					return ctx.Err()
				case s := <-sigCh:
					fmt.Printf("Received signal: %v. Stopping...\n", s)
					cancel() // contextをキャンセルして他のgoroutineも終了させる
					return nil
				case i := <-ch:
					if i%10000 == 0 {
						fmt.Println(i)
					}
					key, err := redisloadtest.GenerateKey(9)
					if err != nil {
						return err
					}
					err = rdb.SetEx(ctx, key, "", redisloadtest.ExpiresSeconds).Err()
					if err != nil {
						return err
					}
					if i >= 130_000_000-1 {
						cancel()
					}
				default:
				}
			}

		})
	}

	log.Print("waiting...")
	if err := eg.Wait(); err != nil && err != context.Canceled {
		log.Print("An error occurred:", err)
	}
}
