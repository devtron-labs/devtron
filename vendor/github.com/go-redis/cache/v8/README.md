# Redis cache library for Golang

[![Build Status](https://travis-ci.org/go-redis/cache.svg)](https://travis-ci.org/go-redis/cache)
[![GoDoc](https://godoc.org/github.com/go-redis/cache?status.svg)](https://pkg.go.dev/github.com/go-redis/cache/v8?tab=doc)

go-redis/cache library implements a cache using Redis as a
key/value storage. It uses [MessagePack](https://github.com/vmihailenco/msgpack) to marshal values.

Optinally you can use [TinyLFU](https://github.com/dgryski/go-tinylfu) or any other [cache algorithm](https://github.com/vmihailenco/go-cache-benchmark) as a local in-process cache.

## Installation

go-redis/cache supports 2 last Go versions and requires a Go version with
[modules](https://github.com/golang/go/wiki/Modules) support. So make sure to initialize a Go
module:

```shell
go mod init github.com/my/repo
```

And then install go-redis/cache/v8 (note _v8_ in the import; omitting it is a popular mistake):

```shell
go get github.com/go-redis/cache/v8
```

## Quickstart

```go
package cache_test

import (
    "context"
    "fmt"
    "time"

    "github.com/go-redis/redis/v8"
    "github.com/go-redis/cache/v8"
)

type Object struct {
    Str string
    Num int
}

func Example_basicUsage() {
    ring := redis.NewRing(&redis.RingOptions{
        Addrs: map[string]string{
            "server1": ":6379",
            "server2": ":6380",
        },
    })

    mycache := cache.New(&cache.Options{
        Redis:      ring,
        LocalCache: cache.NewTinyLFU(1000, time.Minute),
    })

    ctx := context.TODO()
    key := "mykey"
    obj := &Object{
        Str: "mystring",
        Num: 42,
    }

    if err := mycache.Set(&cache.Item{
        Ctx:   ctx,
        Key:   key,
        Value: obj,
        TTL:   time.Hour,
    }); err != nil {
        panic(err)
    }

    var wanted Object
    if err := mycache.Get(ctx, key, &wanted); err == nil {
        fmt.Println(wanted)
    }

    // Output: {mystring 42}
}

func Example_advancedUsage() {
    ring := redis.NewRing(&redis.RingOptions{
        Addrs: map[string]string{
            "server1": ":6379",
            "server2": ":6380",
        },
    })

    mycache := cache.New(&cache.Options{
        Redis:      ring,
        LocalCache: cache.NewTinyLFU(1000, time.Minute),
    })

    obj := new(Object)
    err := mycache.Once(&cache.Item{
        Key:   "mykey",
        Value: obj, // destination
        Do: func(*cache.Item) (interface{}, error) {
            return &Object{
                Str: "mystring",
                Num: 42,
            }, nil
        },
    })
    if err != nil {
        panic(err)
    }
    fmt.Println(obj)
    // Output: &{mystring 42}
}
```
