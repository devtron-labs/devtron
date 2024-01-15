# jsonpath [![build](https://github.com/mdaverde/jsonpath/actions/workflows/build.yml/badge.svg)](https://github.com/mdaverde/jsonpath/actions) [![Go Report Card](https://goreportcard.com/badge/github.com/mdaverde/jsonpath)](https://goreportcard.com/report/github.com/mdaverde/jsonpath) [![GoDoc](https://godoc.org/github.com/mdaverde/jsonpath?status.svg)](https://godoc.org/github.com/mdaverde/jsonpath)

Originally intended to be used with `json.Unmarshal`, this is a golang library used to able get and set jsonpaths (even nonexistent paths).

## Install

```bash
$ go get github.com/mdaverde/jsonpath
```

## Usage

```go
sample := `{ "owner": { "name": "john doe", "contact": { "phone": "555-555-5555" } } }`

var payload interface{}

err := json.Unmarshal([]byte(sample), &payload)
must(err)

err = jsonpath.Set(&payload, "owner.contact.phone", "333-333-3333")
must(err)

value, err := jsonpath.Get(payload, "owner.contact.phone")
must(err)

// value == "333-333-3333"
```

## API

### jsonpath.Get(data interface{}, path string) (interface{}, error)

Returns the value at that json path as `interface{}` and if an error occurred

### jsonpath.Set(data interface{}, path string, value interface{}) (error)

Sets `value` on `data` at that json path

Note: you'll want to pass in a pointer to `data` so that the side effect actually is usable

### jsonpath.DoesNotExist error

Returned by  `jsonpath.Get` on a nonexistent path:

```go
value, err := Get(data, "where.is.this")
if _, ok := err.(DoesNotExist); !ok && err != nil {
    // other error
}
```

## Testing

```bash
$ go test .
```

## License

MIT © [mdaverde](https://mdaverde.com)