package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/labstack/gommon/log"
	"github.com/redis/go-redis/v9"
	"github.com/tetesuzawa/redisloadtest"
	"os"
	"strings"
	"time"
)

// 90 days
func main() {
	addresses := flag.String("addresses", "127.0.0.1", "addresses")
	flag.Parse()

	ctx := context.Background()

	rdb := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        strings.Split(*addresses, ","),
		Password:     "", // no password set
		DialTimeout:  time.Second * 120,
		ReadTimeout:  time.Second * 3600,
		WriteTimeout: time.Second * 3600,
	})
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Print("failed to ping redis:", err)
		os.Exit(1)
	}

	const bigString = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	durations := make([]int64, 0, 1000)
	for i := 0; i < 800; i++ {
		//data := make([]string, 0, 130000*2)
		pipe := rdb.Pipeline()
		for j := 0; j < 100000; j++ {
			key, err := redisloadtest.RandomString(58)
			//data = append(data, key, "")
			//if err != nil {
			//	log.Errorf("failed to generate key: %v", err)
			//	os.Exit(1)
			//}
			err = pipe.SetEx(ctx, key, bigString, redisloadtest.ExpiresSeconds).Err()
			if err != nil {
				log.Errorf("failed SETEX on pipeline")
				os.Exit(1)
			}
		}
		start := time.Now()
		cmds, err := pipe.Exec(ctx)
		duration := time.Now().Sub(start).Milliseconds()
		durations = append(durations, duration)
		if err != nil {
			log.Errorf("failed Exec:", err)
			os.Exit(1)

		}
		for _, cmd := range cmds {
			if err := cmd.Err(); err != nil {
				log.Errorf("failed Exec:", err)
			}
		}
		log.Print("cmds len:", len(cmds))
		//if err := rdb.MSet(ctx, data).Err(); err != nil {
		//	log.Print("failed to set redis:", err)
		//	os.Exit(1)
		//}
		fmt.Println(i * 100000)
	}
	var max int64 = -1
	var min int64 = 10000
	var sum int64 = 0
	for _, dur := range durations {
		if dur > max {
			max = dur
		} else if dur < min {
			min = dur
		}
		sum += dur
	}
	avg := sum / int64(len(durations))

	fmt.Printf("Result: max=%d[ms] avg=%d[ms] min=%d[ms]\n", max, avg, min)
}
