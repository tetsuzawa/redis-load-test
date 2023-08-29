package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/tetesuzawa/redisloadtest"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"
)

const maxDuration = time.Second * 10

func main() {
	addresses := flag.String("addresses", "127.0.0.1", "addresses")
	concurrency := flag.Int("concurrency", 1, "addresses")
	flag.Parse()

	ctx := context.Background()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	rdb := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        strings.Split(*addresses, ","),
		Password:     "", // no password set
		DialTimeout:  time.Second * 120,
		ReadTimeout:  time.Second * 3600,
		WriteTimeout: time.Second * 3600,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Print("failed to ping redis:", err)
		os.Exit(1)
	}

	go func() {
		select {
		case <-signalCh:
			log.Println("Interrupt signal received. Shutting down...")
			cancel()
		case <-ctx.Done():
		}
	}()

	const RequestsPerSecondOnThanosBid = 166666
	limiter := rate.NewLimiter(RequestsPerSecondOnThanosBid, 200000)
	histogram := make([]int, maxDuration/time.Millisecond)

	var mu sync.Mutex
	cnt := 0

	timeFormat := "2006-01-02 15:04:05"
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				fmt.Println("ticker done")
				return
			case <-ticker.C:
				mu.Lock()
				fmt.Printf("%v  cnt: %d\n", time.Now().Format(timeFormat), cnt)
				cnt = 0
				mu.Unlock()
			}
		}
	}()

	for i := 0; i < *concurrency; i++ {
		g.Go(func() error {
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					if err := limiter.Wait(ctx); err != nil {
						return err
					}

					key, err := redisloadtest.GenerateKey(16)
					if err != nil {
						return err
					}

					start := time.Now()
					_, err = rdb.Exists(ctx, key).Result()
					if err != nil && !errors.Is(err, redis.Nil) {
						return err
					}
					duration := time.Since(start)
					if duration > maxDuration {
						duration = maxDuration
					}

					mu.Lock()
					cnt++
					histogram[duration/time.Millisecond]++
					mu.Unlock()
				}
			}
		})
	}

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println("Duration(ms)\tCount")
	for i, count := range histogram {
		if count > 0 {
			fmt.Printf("%d\t\t%d\n", i, count)
		}
	}
}
