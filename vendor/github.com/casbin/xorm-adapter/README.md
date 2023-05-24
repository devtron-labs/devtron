Xorm Adapter [![Build Status](https://travis-ci.org/casbin/xorm-adapter.svg?branch=master)](https://travis-ci.org/casbin/xorm-adapter) [![Coverage Status](https://coveralls.io/repos/github/casbin/xorm-adapter/badge.svg?branch=master)](https://coveralls.io/github/casbin/xorm-adapter?branch=master) [![Godoc](https://godoc.org/github.com/casbin/xorm-adapter?status.svg)](https://godoc.org/github.com/casbin/xorm-adapter)
====

Xorm Adapter is the [Xorm](https://github.com/go-xorm/xorm) adapter for [Casbin](https://github.com/casbin/casbin). With this library, Casbin can load policy from Xorm supported database or save policy to it.

Based on [Xorm Drivers Support](https://github.com/go-xorm/xorm#drivers-support), The current supported databases are:

- Mysql: [github.com/go-sql-driver/mysql](https://github.com/go-sql-driver/mysql)
- MyMysql: [github.com/ziutek/mymysql/godrv](https://github.com/ziutek/mymysql/godrv)
- Postgres: [github.com/lib/pq](https://github.com/lib/pq)
- Tidb: [github.com/pingcap/tidb](https://github.com/pingcap/tidb)
- SQLite: [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)
- MsSql: [github.com/denisenkom/go-mssqldb](https://github.com/denisenkom/go-mssqldb)
- Oracle: [github.com/mattn/go-oci8](https://github.com/mattn/go-oci8) (experiment)

## Installation

    go get github.com/casbin/xorm-adapter

## Simple MySQL Example

```go
package main

import (
	"github.com/casbin/casbin"
	"github.com/casbin/xorm-adapter"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Initialize a Xorm adapter and use it in a Casbin enforcer:
	// The adapter will use the MySQL database named "casbin".
	// If it doesn't exist, the adapter will create it automatically.
	a := xormadapter.NewAdapter("mysql", "mysql_username:mysql_password@tcp(127.0.0.1:3306)/") // Your driver and data source. 

	// Or you can use an existing DB "abc" like this:
	// The adapter will use the table named "casbin_rule".
	// If it doesn't exist, the adapter will create it automatically.
	// a := xormadapter.NewAdapter("mysql", "mysql_username:mysql_password@tcp(127.0.0.1:3306)/abc", true)

	e := casbin.NewEnforcer("examples/rbac_model.conf", a)
	
	// Load the policy from DB.
	e.LoadPolicy()
	
	// Check the permission.
	e.Enforce("alice", "data1", "read")
	
	// Modify the policy.
	// e.AddPolicy(...)
	// e.RemovePolicy(...)
	
	// Save the policy back to DB.
	e.SavePolicy()
}
```

## Simple Postgres Example

```go
package main

import (
	"github.com/casbin/casbin"
	"github.com/casbin/xorm-adapter"
	_ "github.com/lib/pq"
)

func main() {
	// Initialize a Xorm adapter and use it in a Casbin enforcer:
	// The adapter will use the Postgres database named "casbin".
	// If it doesn't exist, the adapter will create it automatically.
	a := xormadapter.NewAdapter("postgres", "user=postgres_username password=postgres_password host=127.0.0.1 port=5432 sslmode=disable") // Your driver and data source.

	// Or you can use an existing DB "abc" like this:
	// The adapter will use the table named "casbin_rule".
	// If it doesn't exist, the adapter will create it automatically.
	// a := xormadapter.NewAdapter("postgres", "dbname=abc user=postgres_username password=postgres_password host=127.0.0.1 port=5432 sslmode=disable", true)

	e := casbin.NewEnforcer("../examples/rbac_model.conf", a)

	// Load the policy from DB.
	e.LoadPolicy()

	// Check the permission.
	e.Enforce("alice", "data1", "read")

	// Modify the policy.
	// e.AddPolicy(...)
	// e.RemovePolicy(...)

	// Save the policy back to DB.
	e.SavePolicy()
}
```

## Getting Help

- [Casbin](https://github.com/casbin/casbin)

## License

This project is under Apache 2.0 License. See the [LICENSE](LICENSE) file for the full license text.
