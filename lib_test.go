package redisloadtest

import (
	"fmt"
	"testing"
)

func Test_aa(t *testing.T) {
	key, _ := GenerateKey(9)
	fmt.Println(len([]byte(key)))
}
