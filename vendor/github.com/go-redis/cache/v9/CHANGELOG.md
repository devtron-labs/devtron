# Changelog

> :heart: [**Uptrace.dev** - distributed traces, logs, and errors in one place](https://uptrace.dev)

## v8

- Added s2 (snappy) compression. That means that v8 can't read the data set by v7.
- Replaced LRU with TinyLFU for local cache.
- Requires go-redis v8.
