package vbasedata

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
)

func TestLruCache(t *testing.T) {
	lru := expirable.NewLRU[string, string](1, nil, 3*time.Second)

	for i := 0; i < 3; i++ {
		// 添加一个缓存
		log.Print(lru.Add(fmt.Sprintf("%v", i), fmt.Sprintf("%v", i)))
		// 获取一个缓存
		log.Print(lru.Get(fmt.Sprintf("%v", i)))
		// 获取一个缓存
		log.Print(lru.Get(fmt.Sprintf("%v", i-1)))
	}
}
