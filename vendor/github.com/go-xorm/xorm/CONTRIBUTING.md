## Contributing to xorm

`xorm` has a backlog of [pull requests](https://help.github.com/articles/using-pull-requests), but contributions are still very
much welcome. You can help with patch review, submitting bug reports,
or adding new functionality. There is no formal style guide, but
please conform to the style of existing code and general Go formatting
conventions when submitting patches.

* [fork a repo](https://help.github.com/articles/fork-a-repo)
* [creating a pull request ](https://help.github.com/articles/creating-a-pull-request)

### Language

Since `xorm` is a world-wide open source project, please describe your issues or code changes in English as soon as possible.

### Sign your codes with comments
```
// !<you github id>! your comments

e.g.,

// !lunny! this is comments made by lunny
```

### Patch review

Help review existing open [pull requests](https://help.github.com/articles/using-pull-requests) by commenting on the code or
proposed functionality.

### Bug reports

We appreciate any bug reports, but especially ones with self-contained
(doesn't depend on code outside of xorm), minimal (can't be simplified
further) test cases. It's especially helpful if you can submit a pull
request with just the failing test case(you can find some example test file like [session_get_test.go](https://github.com/go-xorm/xorm/blob/master/session_get_test.go)).

If you implements a new database interface, you maybe need to add a test_<databasename>.sh file.
For example, [mysql_test.go](https://github.com/go-xorm/xorm/blob/master/test_mysql.sh)

### New functionality

There are a number of pending patches for new functionality, so
additional feature patches will take a while to merge. Still, patches
are generally reviewed based on usefulness and complexity in addition
to time-in-queue, so if you have a knockout idea, take a shot. Feel
free to open an issue discussion your proposed patch beforehand.
