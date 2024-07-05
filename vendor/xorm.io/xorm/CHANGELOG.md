# Changelog

This changelog goes through all the changes that have been made in each release
without substantial changes to our git log.

## [1.0.3](https://gitea.com/xorm/xorm/pulls?q=&type=all&state=closed&milestone=1281) - 2020-07-10

* BUGFIXES
  * Fix dump of sqlite (#1639)
* ENHANCEMENTS
  * Fix index name parsing in SQLite dialect (#1737)
  * add hooks for Commit and Rollback (#1733)

## [1.0.2](https://gitea.com/xorm/xorm/pulls?q=&type=all&state=closed&milestone=1261) - 2020-06-16

* FEATURES
  * Add Hook (#1644)
* BUGFIXES
  * Fix bug when ID used but no reference table given (#1709)
  * Fix find and count bug (#1651)
* ENHANCEMENTS
  * chore: improve snakeCasedName performance (#1688)
  * Fix find with another struct (#1666)
  * fix GetColumns missing ordinal position (#1660)
* MISC
  * chore: improve titleCasedName performance (#1691)

## [1.0.1](https://gitea.com/xorm/xorm/pulls?q=&type=all&state=closed&milestone=1253) - 2020-03-25

* BUGFIXES
  * Oracle : Local Naming Method (#1515)
  * Fix find and count bug (#1618)
  * Fix duplicated deleted condition on FindAndCount (#1619)
  * Fix find and count bug with cache (#1622)
  * Fix postgres schema problem (#1624)
  * Fix quote with blank (#1626)

## [1.0.0](https://gitea.com/xorm/xorm/pulls?q=&type=all&state=closed&milestone=1242) - 2020-03-22

* BREAKING
  * Add context for dialects (#1558)
  * Move zero functions to a standalone package (#1548)
  * Merge core package back into the main repository and split into serval sub packages. (#1543)
* FEATURES
  * Use a new ContextLogger interface to implement logger (#1557)
* BUGFIXES
  * Fix setschema (#1606)
  * Fix dump/import bug (#1603)
  * Fix pk bug (#1602)
  * Fix master/slave bug (#1601)
  * Fix bug when dump (#1597)
  * Ignore schema when dbtype is not postgres (#1593)
  * Fix table name (#1590)
  * Fix find alias bug (#1581)
  * Fix rows bug (#1576)
  * Fix map with cols (#1575)
  * Fix bug on deleted with join (#1570)
  * Improve quote policy (#1567)
  * Fix break session sql enable feature (#1566)
  * Fix mssql quote (#1535)
  * Fix join table name quote bug (#1534)
  * Fix mssql issue with duplicate columns. (#1225)
  * Fix mysql8.0 sync failed (#808)
* ENHANCEMENTS
  * Fix batch insert interface slice be panic (#1598)
  * Move some codes to statement sub package (#1574)
  * Remove circle file (#1569)
  * Move statement as a sub package (#1564)
  * Move maptype to tag parser (#1561)
  * Move caches to manager (#1553)
  * Improve code (#1552)
  * Improve some codes (#1551)
  * Improve statement (#1549)
  * Move tag parser related codes as a standalone sub package (#1547)
  * Move reserve words related files into dialects sub package (#1544)
  * Fix `Conversion` method `ToDB() ([]byte, error)` return type is nil (#1296)
  * Check driver.Valuer response, and skip the column if nil (#1167)
  * Add cockroach support and tests (#896)
* TESTING
  * Improve tests (#1572)
* BUILD
  * Add changelog file and tool configuration (#1546)
* DOCS
  * Fix outdate changelog (#1565)

## old changelog

* **v0.6.5**
    * Postgres schema support
    * vgo support
    * Add FindAndCount
    * Database special params support via NewEngineWithParams
    * Some bugs fixed

* **v0.6.4**
    * Automatical Read/Write seperatelly
    * Query/QueryString/QueryInterface and action with Where/And
    * Get support non-struct variables
    * BufferSize on Iterate
    * fix some other bugs.

* **v0.6.3**
    * merge tests to main project
    * add `Exist` function
    * add `SumInt` function
    * Mysql now support read and create column comment.
    * fix time related bugs.
    * fix some other bugs.

* **v0.6.2**
    * refactor tag parse methods
    * add Scan features to Get
    * add QueryString method

* **v0.4.5**
    * many bugs fixed
    * extends support unlimited deep
    * Delete Limit support

* **v0.4.4**
    * ql database expriment support
    * tidb database expriment support
    * sql.NullString and etc. field support
    * select ForUpdate support
    * many bugs fixed

* **v0.4.3**
    * Json column type support
    * oracle expirement support
    * bug fixed

* **v0.4.2**
	* Transaction will auto rollback if not Rollback or Commit be called.
    * Gonic Mapper support
    * bug fixed

* **v0.4.1**
    * deleted tag support for soft delete
    * bug fixed

* **v0.4.0 RC1**
	Changes:
	* moved xorm cmd to [github.com/go-xorm/cmd](github.com/go-xorm/cmd)
	* refactored general DB operation a core lib at [github.com/go-xorm/core](https://github.com/go-xorm/core)
	* moved tests to github.com/go-xorm/tests [github.com/go-xorm/tests](github.com/go-xorm/tests)

	Improvements:
	* Prepared statement cache
	* Add Incr API
	* Specify Timezone Location

* **v0.3.2**
	Improvements:
	* Add AllCols & MustCols function
	* Add TableName for custom table name

	Bug Fixes:
	* #46
	* #51
	* #53
	* #89
	* #86
	* #92

* **v0.3.1**

	Features:
	* Support MSSQL DB via ODBC driver ([github.com/lunny/godbc](https://github.com/lunny/godbc));
	* Composite Key, using multiple pk xorm tag
	* Added Row() API as alternative to Iterate() API for traversing result set, provide similar usages to sql.Rows type
	* ORM struct allowed declaration of pointer builtin type as members to allow null DB fields
	* Before and After Event processors

	Improvements:
	* Allowed int/int32/int64/uint/uint32/uint64/string as Primary Key type
	* Performance improvement for Get()/Find()/Iterate()


* **v0.2.3** : Improved documents; Optimistic Locking support; Timestamp with time zone support; Mapper change to tableMapper and columnMapper & added PrefixMapper & SuffixMapper support custom table or column name's prefix and suffix;Insert now return affected, err instead of id, err; Added UseBool & Distinct;

* **v0.2.2** : Postgres drivers now support lib/pq; Added method Iterate for record by record to handler；Added SetMaxConns(go1.2+) support; some bugs fixed.

* **v0.2.1** : Added database reverse tool, now support generate go & c++ codes, see [Xorm Tool README](https://github.com/go-xorm/xorm/blob/master/xorm/README.md); some bug fixed.

* **v0.2.0** : Added Cache supported, select is speeder up 3~5x; Added SameMapper for same name between struct and table; Added Sync method for auto added tables, columns, indexes;

* **v0.1.9** : Added postgres and mymysql supported; Added ` and ? supported on Raw SQL even if postgres; Added Cols, StoreEngine, Charset function, Added many column data type supported, please see [Mapping Rules](#mapping).

* **v0.1.8** : Added union index and union unique supported, please see [Mapping Rules](#mapping).

* **v0.1.7** : Added IConnectPool interface and NoneConnectPool, SysConnectPool, SimpleConnectPool the three implements. You can choose one of them and the default is SysConnectPool. You can customrize your own connection pool. struct Engine added Close method, It should be invoked before system exit.

* **v0.1.6** : Added conversion interface support; added struct derive support; added single mapping support

* **v0.1.5** : Added multi threads support; added Sql() function for struct query; Get function changed return inteface; MakeSession and Create are instead with NewSession and NewEngine.

* **v0.1.4** : Added simple cascade load support; added more data type supports.

* **v0.1.3** : Find function now supports both slice and map; Add Table function for multi tables and temperory tables support

* **v0.1.2** : Insert function now supports both struct and slice pointer parameters, batch inserting and auto transaction

* **v0.1.1** : Add Id, In functions and improved README

* **v0.1.0** : Initial release.