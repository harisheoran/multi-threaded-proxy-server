package main

import (
	cache_lru "harisheoran/multithreaded-proxy-web-server/internal/cache"
	"log"
)

type app struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	LRUCache    *cache_lru.CacheList
}
