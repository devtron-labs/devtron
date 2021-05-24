JSON Adapter [![Build Status](https://travis-ci.org/casbin/json-adapter.svg?branch=master)](https://travis-ci.org/casbin/json-adapter) [![Coverage Status](https://coveralls.io/repos/github/casbin/json-adapter/badge.svg?branch=master)](https://coveralls.io/github/casbin/json-adapter?branch=master) [![Godoc](https://godoc.org/github.com/casbin/json-adapter?status.svg)](https://godoc.org/github.com/casbin/json-adapter)
====

JSON Adapter is the [JSON (JavaScript Object Notation)](https://www.json.org/) adapter for [Casbin](https://github.com/casbin/casbin). With this library, Casbin can load policy from JSON string or save policy to it.

## Installation

    go get github.com/casbin/json-adapter

## Simple Example

```go
package main

import (
	"github.com/casbin/casbin"
	"github.com/casbin/json-adapter"
)

func main() {
	// Initialize a JSON adapter and use it in a Casbin enforcer:
	b := []byte{} // b stores Casbin policy in JSON bytes.
	a := jsonadapter.NewAdapter(&b) // Use b as the data source. 
	e := casbin.NewEnforcer("examples/rbac_model.conf", a)
	
	// Load the policy from JSON bytes b.
	e.LoadPolicy()
	
	// Check the permission.
	e.Enforce("alice", "data1", "read")
	
	// Modify the policy.
	// e.AddPolicy(...)
	// e.RemovePolicy(...)
	
	// Save the policy back to JSON bytes b.
	e.SavePolicy()
}
```

## Getting Help

- [Casbin](https://github.com/casbin/casbin)

## License

This project is under Apache 2.0 License. See the [LICENSE](LICENSE) file for the full license text.
