package sqladapter

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	db "github.com/upper/db/v4"
	"github.com/upper/db/v4/internal/cache"
	"github.com/upper/db/v4/internal/sqladapter/compat"
	"github.com/upper/db/v4/internal/sqladapter/exql"
	"github.com/upper/db/v4/internal/sqlbuilder"
)

var (
	lastSessID uint64
	lastTxID   uint64
)

var (
	slowQueryThreshold          = time.Millisecond * 200
	retryTransactionWaitTime    = time.Millisecond * 10
	retryTransactionMaxWaitTime = time.Second * 1
)

// hasCleanUp is implemented by structs that have a clean up routine that needs
// to be called before Close().
type hasCleanUp interface {
	CleanUp() error
}

// statementExecer allows the adapter to have its own exec statement.
type statementExecer interface {
	StatementExec(sess Session, ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

// statementCompiler transforms an internal statement into a format
// database/sql can understand.
type statementCompiler interface {
	CompileStatement(sess Session, stmt *exql.Statement, args []interface{}) (string, []interface{}, error)
}

// sessValueConverter converts values before being passed to the underlying driver.
type sessValueConverter interface {
	ConvertValue(in interface{}) interface{}
}

// sessValueConverterContext converts values before being passed to the underlying driver.
type sessValueConverterContext interface {
	ConvertValueContext(ctx context.Context, in interface{}) interface{}
}

// valueConverter converts values before being passed to the underlying driver.
type valueConverter interface {
	ConvertValue(in interface{}) interface {
		sql.Scanner
		driver.Valuer
	}
}

// errorConverter converts an error value from the underlying driver into
// something different.
type errorConverter interface {
	Err(errIn error) (errOut error)
}

// AdapterSession defines methods to be implemented by SQL adapters.
type AdapterSession interface {
	Template() *exql.Template

	NewCollection() CollectionAdapter

	// Open opens a new connection
	OpenDSN(sess Session, dsn string) (*sql.DB, error)

	// Collections returns a list of non-system tables from the database.
	Collections(sess Session) ([]string, error)

	// TableExists returns an error if the given table does not exist.
	TableExists(sess Session, name string) error

	// LookupName returns the name of the database.
	LookupName(sess Session) (string, error)

	// PrimaryKeys returns all primary keys on the table.
	PrimaryKeys(sess Session, name string) ([]string, error)
}

// Session satisfies db.Session.
type Session interface {
	SQL() db.SQL

	// PrimaryKeys returns all primary keys on the table.
	PrimaryKeys(tableName string) ([]string, error)

	// Collections returns a list of references to all collections in the
	// database.
	Collections() ([]db.Collection, error)

	// Name returns the name of the database.
	Name() string

	// Close closes the database session
	Close() error

	// Ping checks if the database server is reachable.
	Ping() error

	// Reset clears all caches the session is using
	Reset()

	// Collection returns a new collection.
	Collection(string) db.Collection

	// ConnectionURL returns the ConnectionURL that was used to create the
	// Session.
	ConnectionURL() db.ConnectionURL

	// Open attempts to establish a connection to the database server.
	Open() error

	// TableExists returns an error if the table doesn't exists.
	TableExists(name string) error

	// Driver returns the underlying driver the session is using
	Driver() interface{}

	Save(db.Record) error

	Get(db.Record, interface{}) error

	Delete(db.Record) error

	// WaitForConnection attempts to run the given connection function a fixed
	// number of times before failing.
	WaitForConnection(func() error) error

	// BindDB sets the *sql.DB the session will use.
	BindDB(*sql.DB) error

	// Session returns the *sql.DB the session is using.
	DB() *sql.DB

	// BindTx binds a transaction to the current session.
	BindTx(context.Context, *sql.Tx) error

	// Returns the current transaction the session is using.
	Transaction() *sql.Tx

	// NewClone clones the database using the given AdapterSession as base.
	NewClone(AdapterSession, bool) (Session, error)

	// Context returns the default context the session is using.
	Context() context.Context

	// SetContext sets the default context for the session.
	SetContext(context.Context)

	NewTransaction(ctx context.Context, opts *sql.TxOptions) (Session, error)

	Tx(fn func(sess db.Session) error) error

	TxContext(ctx context.Context, fn func(sess db.Session) error, opts *sql.TxOptions) error

	WithContext(context.Context) db.Session

	IsTransaction() bool

	Commit() error

	Rollback() error

	db.Settings
}

// NewTx wraps a *sql.Tx and returns a Tx.
func NewTx(adapter AdapterSession, tx *sql.Tx) (Session, error) {
	sessTx := &sessionWithContext{
		session: &session{
			Settings: db.DefaultSettings,

			sqlTx:             tx,
			adapter:           adapter,
			cachedPKs:         cache.NewCache(),
			cachedCollections: cache.NewCache(),
			cachedStatements:  cache.NewCache(),
		},
		ctx: context.Background(),
	}
	return sessTx, nil
}

// NewSession creates a new Session.
func NewSession(connURL db.ConnectionURL, adapter AdapterSession) Session {
	sess := &sessionWithContext{
		session: &session{
			Settings: db.DefaultSettings,

			connURL:           connURL,
			adapter:           adapter,
			cachedPKs:         cache.NewCache(),
			cachedCollections: cache.NewCache(),
			cachedStatements:  cache.NewCache(),
		},
		ctx: context.Background(),
	}
	return sess
}

type session struct {
	db.Settings

	adapter AdapterSession

	connURL db.ConnectionURL

	builder db.SQL

	lookupNameOnce sync.Once
	name           string

	mu        sync.Mutex // guards ctx, txOptions
	txOptions *sql.TxOptions

	sqlDBMu sync.Mutex // guards sess, baseTx

	sqlDB *sql.DB
	sqlTx *sql.Tx

	sessID uint64
	txID   uint64

	cacheMu           sync.Mutex // guards cachedStatements and cachedCollections
	cachedPKs         *cache.Cache
	cachedStatements  *cache.Cache
	cachedCollections *cache.Cache

	template *exql.Template
}

type sessionWithContext struct {
	*session

	ctx context.Context
}

func (sess *sessionWithContext) WithContext(ctx context.Context) db.Session {
	if ctx == nil {
		panic("nil context")
	}
	newSess := &sessionWithContext{
		session: sess.session,
		ctx:     ctx,
	}
	return newSess
}

func (sess *sessionWithContext) Tx(fn func(sess db.Session) error) error {
	return TxContext(sess.Context(), sess, fn, nil)
}

func (sess *sessionWithContext) TxContext(ctx context.Context, fn func(sess db.Session) error, opts *sql.TxOptions) error {
	return TxContext(ctx, sess, fn, opts)
}

func (sess *sessionWithContext) SQL() db.SQL {
	return sqlbuilder.WithSession(
		sess,
		sess.adapter.Template(),
	)
}

func (sess *sessionWithContext) Err(errIn error) (errOur error) {
	if convertError, ok := sess.adapter.(errorConverter); ok {
		return convertError.Err(errIn)
	}
	return errIn
}

func (sess *sessionWithContext) PrimaryKeys(tableName string) ([]string, error) {
	h := cache.NewHashable(hashTypePrimaryKeys, tableName)

	cachedPK, ok := sess.cachedPKs.ReadRaw(h)
	if ok {
		return cachedPK.([]string), nil
	}

	pk, err := sess.adapter.PrimaryKeys(sess, tableName)
	if err != nil {
		return nil, err
	}

	sess.cachedPKs.Write(h, pk)
	return pk, nil
}

func (sess *sessionWithContext) TableExists(name string) error {
	return sess.adapter.TableExists(sess, name)
}

func (sess *sessionWithContext) NewTransaction(ctx context.Context, opts *sql.TxOptions) (Session, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	clone, err := sess.NewClone(sess.adapter, false)
	if err != nil {
		return nil, err
	}

	connFn := func() error {
		sqlTx, err := compat.BeginTx(clone.DB(), clone.Context(), opts)
		if err == nil {
			return clone.BindTx(ctx, sqlTx)
		}
		return err
	}

	if err := clone.WaitForConnection(connFn); err != nil {
		return nil, err
	}

	return clone, nil
}

func (sess *sessionWithContext) Collections() ([]db.Collection, error) {
	names, err := sess.adapter.Collections(sess)
	if err != nil {
		return nil, err
	}

	collections := make([]db.Collection, 0, len(names))
	for i := range names {
		collections = append(collections, sess.Collection(names[i]))
	}

	return collections, nil
}

func (sess *sessionWithContext) ConnectionURL() db.ConnectionURL {
	return sess.connURL
}

func (sess *sessionWithContext) Open() error {
	var sqlDB *sql.DB
	var err error

	connFn := func() error {
		sqlDB, err = sess.adapter.OpenDSN(sess, sess.connURL.String())
		if err != nil {
			return err
		}

		sqlDB.SetConnMaxLifetime(sess.ConnMaxLifetime())
		sqlDB.SetConnMaxIdleTime(sess.ConnMaxIdleTime())
		sqlDB.SetMaxIdleConns(sess.MaxIdleConns())
		sqlDB.SetMaxOpenConns(sess.MaxOpenConns())
		return nil
	}

	if err := sess.WaitForConnection(connFn); err != nil {
		return err
	}

	return sess.BindDB(sqlDB)
}

func (sess *sessionWithContext) Get(record db.Record, id interface{}) error {
	store := record.Store(sess)
	if getter, ok := store.(db.StoreGetter); ok {
		return getter.Get(record, id)
	}
	return store.Find(id).One(record)
}

func (sess *sessionWithContext) Save(record db.Record) error {
	if record == nil {
		return db.ErrNilRecord
	}

	if reflect.TypeOf(record).Kind() != reflect.Ptr {
		return db.ErrExpectingPointerToStruct
	}

	store := record.Store(sess)
	if saver, ok := store.(db.StoreSaver); ok {
		return saver.Save(record)
	}

	id := db.Cond{}
	keys, values, err := recordPrimaryKeyFieldValues(store, record)
	if err != nil {
		return err
	}
	for i := range values {
		if values[i] != reflect.Zero(reflect.TypeOf(values[i])).Interface() {
			id[keys[i]] = values[i]
		}
	}

	if len(id) > 0 && len(id) == len(values) {
		// check if record exists before updating it
		exists, _ := store.Find(id).Count()
		if exists > 0 {
			return recordUpdate(store, record)
		}
	}

	return recordCreate(store, record)
}

func (sess *sessionWithContext) Delete(record db.Record) error {
	if record == nil {
		return db.ErrNilRecord
	}

	if reflect.TypeOf(record).Kind() != reflect.Ptr {
		return db.ErrExpectingPointerToStruct
	}

	store := record.Store(sess)

	if hook, ok := record.(db.BeforeDeleteHook); ok {
		if err := hook.BeforeDelete(sess); err != nil {
			return err
		}
	}

	if deleter, ok := store.(db.StoreDeleter); ok {
		if err := deleter.Delete(record); err != nil {
			return err
		}
	} else {
		conds, err := recordID(store, record)
		if err != nil {
			return err
		}
		if err := store.Find(conds).Delete(); err != nil {
			return err
		}
	}

	if hook, ok := record.(db.AfterDeleteHook); ok {
		if err := hook.AfterDelete(sess); err != nil {
			return err
		}
	}

	return nil
}

func (sess *sessionWithContext) DB() *sql.DB {
	return sess.sqlDB
}

func (sess *sessionWithContext) SetContext(ctx context.Context) {
	sess.mu.Lock()
	sess.ctx = ctx
	sess.mu.Unlock()
}

func (sess *sessionWithContext) Context() context.Context {
	return sess.ctx
}

func (sess *sessionWithContext) SetTxOptions(txOptions sql.TxOptions) {
	sess.mu.Lock()
	sess.txOptions = &txOptions
	sess.mu.Unlock()
}

func (sess *sessionWithContext) TxOptions() *sql.TxOptions {
	sess.mu.Lock()
	defer sess.mu.Unlock()
	if sess.txOptions == nil {
		return nil
	}
	return sess.txOptions
}

func (sess *sessionWithContext) BindTx(ctx context.Context, tx *sql.Tx) error {
	sess.sqlDBMu.Lock()
	defer sess.sqlDBMu.Unlock()

	sess.sqlTx = tx
	sess.SetContext(ctx)

	sess.txID = newBaseTxID()

	return nil
}

func (sess *sessionWithContext) Commit() error {
	if sess.sqlTx != nil {
		return sess.sqlTx.Commit()
	}
	return db.ErrNotWithinTransaction
}

func (sess *sessionWithContext) Rollback() error {
	if sess.sqlTx != nil {
		return sess.sqlTx.Rollback()
	}
	return db.ErrNotWithinTransaction
}

func (sess *sessionWithContext) IsTransaction() bool {
	return sess.sqlTx != nil
}

func (sess *sessionWithContext) Transaction() *sql.Tx {
	return sess.sqlTx
}

func (sess *sessionWithContext) Name() string {
	sess.lookupNameOnce.Do(func() {
		if sess.name == "" {
			sess.name, _ = sess.adapter.LookupName(sess)
		}
	})

	return sess.name
}

func (sess *sessionWithContext) BindDB(sqlDB *sql.DB) error {

	sess.sqlDBMu.Lock()
	sess.sqlDB = sqlDB
	sess.sqlDBMu.Unlock()

	if err := sess.Ping(); err != nil {
		return err
	}

	sess.sessID = newSessionID()
	name, err := sess.adapter.LookupName(sess)
	if err != nil {
		return err
	}
	sess.name = name

	return nil
}

func (sess *sessionWithContext) Ping() error {
	if sess.sqlDB != nil {
		return sess.sqlDB.Ping()
	}
	return db.ErrNotConnected
}

func (sess *sessionWithContext) SetConnMaxLifetime(t time.Duration) {
	sess.Settings.SetConnMaxLifetime(t)
	if sessDB := sess.DB(); sessDB != nil {
		sessDB.SetConnMaxLifetime(sess.Settings.ConnMaxLifetime())
	}
}

func (sess *sessionWithContext) SetConnMaxIdleTime(t time.Duration) {
	sess.Settings.SetConnMaxIdleTime(t)
	if sessDB := sess.DB(); sessDB != nil {
		sessDB.SetConnMaxIdleTime(sess.Settings.ConnMaxIdleTime())
	}
}

func (sess *sessionWithContext) SetMaxIdleConns(n int) {
	sess.Settings.SetMaxIdleConns(n)
	if sessDB := sess.DB(); sessDB != nil {
		sessDB.SetMaxIdleConns(sess.Settings.MaxIdleConns())
	}
}

func (sess *sessionWithContext) SetMaxOpenConns(n int) {
	sess.Settings.SetMaxOpenConns(n)
	if sessDB := sess.DB(); sessDB != nil {
		sessDB.SetMaxOpenConns(sess.Settings.MaxOpenConns())
	}
}

// Reset removes all caches.
func (sess *sessionWithContext) Reset() {
	sess.cacheMu.Lock()
	defer sess.cacheMu.Unlock()

	sess.cachedPKs.Clear()
	sess.cachedCollections.Clear()
	sess.cachedStatements.Clear()

	if sess.template != nil {
		sess.template.Cache.Clear()
	}
}

func (sess *sessionWithContext) NewClone(adapter AdapterSession, checkConn bool) (Session, error) {

	newSess := NewSession(sess.connURL, adapter).(*sessionWithContext)

	newSess.name = sess.name
	newSess.sqlDB = sess.sqlDB
	newSess.cachedPKs = sess.cachedPKs

	if checkConn {
		if err := newSess.Ping(); err != nil {
			// Retry once if ping fails.
			return sess.NewClone(adapter, false)
		}
	}

	newSess.sessID = newSessionID()

	// New transaction should inherit parent settings
	copySettings(sess, newSess)

	return newSess, nil
}

func (sess *sessionWithContext) Close() error {
	defer func() {
		sess.sqlDBMu.Lock()
		sess.sqlDB = nil
		sess.sqlTx = nil
		sess.sqlDBMu.Unlock()
	}()

	if sess.sqlDB == nil {
		return nil
	}

	sess.cachedCollections.Clear()
	sess.cachedStatements.Clear() // Closes prepared statements as well.

	if !sess.IsTransaction() {
		if cleaner, ok := sess.adapter.(hasCleanUp); ok {
			if err := cleaner.CleanUp(); err != nil {
				return err
			}
		}
		// Not within a transaction.
		return sess.sqlDB.Close()
	}

	return nil
}

func (sess *sessionWithContext) Collection(name string) db.Collection {
	sess.cacheMu.Lock()
	defer sess.cacheMu.Unlock()

	h := cache.NewHashable(hashTypeCollection, name)

	col, ok := sess.cachedCollections.ReadRaw(h)
	if !ok {
		col = newCollection(name, sess.adapter.NewCollection())
		sess.cachedCollections.Write(h, col)
	}

	return &collectionWithSession{
		collection: col.(*collection),
		session:    sess,
	}
}

func queryLog(status *db.QueryStatus) {
	diff := status.End.Sub(status.Start)

	slowQuery := false
	if diff >= slowQueryThreshold {
		status.Err = db.ErrWarnSlowQuery
		slowQuery = true
	}

	if status.Err != nil || slowQuery {
		db.LC().Warn(status)
		return
	}

	db.LC().Debug(status)
}

func (sess *sessionWithContext) StatementPrepare(ctx context.Context, stmt *exql.Statement) (sqlStmt *sql.Stmt, err error) {
	var query string

	defer func(start time.Time) {
		queryLog(&db.QueryStatus{
			TxID:     sess.txID,
			SessID:   sess.sessID,
			RawQuery: query,
			Err:      err,
			Start:    start,
			End:      time.Now(),
			Context:  ctx,
		})
	}(time.Now())

	query, _, err = sess.compileStatement(stmt, nil)
	if err != nil {
		return nil, err
	}

	tx := sess.Transaction()
	if tx != nil {
		sqlStmt, err = compat.PrepareContext(tx, ctx, query)
		return
	}

	sqlStmt, err = compat.PrepareContext(sess.sqlDB, ctx, query)
	return
}

func (sess *sessionWithContext) ConvertValue(value interface{}) interface{} {
	if scannerValuer, ok := value.(sqlbuilder.ScannerValuer); ok {
		return scannerValuer
	}

	dv := reflect.Indirect(reflect.ValueOf(value))
	if dv.IsValid() {
		if converter, ok := dv.Interface().(valueConverter); ok {
			return converter.ConvertValue(dv.Interface())
		}
	}

	if converter, ok := sess.adapter.(sessValueConverterContext); ok {
		return converter.ConvertValueContext(sess.Context(), value)
	}

	if converter, ok := sess.adapter.(sessValueConverter); ok {
		return converter.ConvertValue(value)
	}

	return value
}

func (sess *sessionWithContext) StatementExec(ctx context.Context, stmt *exql.Statement, args ...interface{}) (res sql.Result, err error) {
	var query string

	defer func(start time.Time) {
		status := db.QueryStatus{
			TxID:     sess.txID,
			SessID:   sess.sessID,
			RawQuery: query,
			Args:     args,
			Err:      err,
			Start:    start,
			End:      time.Now(),
			Context:  ctx,
		}

		if res != nil {
			if rowsAffected, err := res.RowsAffected(); err == nil {
				status.RowsAffected = &rowsAffected
			}

			if lastInsertID, err := res.LastInsertId(); err == nil {
				status.LastInsertID = &lastInsertID
			}
		}

		queryLog(&status)
	}(time.Now())

	if execer, ok := sess.adapter.(statementExecer); ok {
		query, args, err = sess.compileStatement(stmt, args)
		if err != nil {
			return nil, err
		}
		res, err = execer.StatementExec(sess, ctx, query, args...)
		return
	}

	tx := sess.Transaction()
	if sess.Settings.PreparedStatementCacheEnabled() && tx == nil {
		var p *Stmt
		if p, query, args, err = sess.prepareStatement(ctx, stmt, args); err != nil {
			return nil, err
		}
		defer p.Close()

		res, err = compat.PreparedExecContext(p, ctx, args)
		return
	}

	query, args, err = sess.compileStatement(stmt, args)
	if err != nil {
		return nil, err
	}

	if tx != nil {
		res, err = compat.ExecContext(tx, ctx, query, args)
		return
	}

	res, err = compat.ExecContext(sess.sqlDB, ctx, query, args)
	return
}

// StatementQuery compiles and executes a statement that returns rows.
func (sess *sessionWithContext) StatementQuery(ctx context.Context, stmt *exql.Statement, args ...interface{}) (rows *sql.Rows, err error) {
	var query string

	defer func(start time.Time) {
		status := db.QueryStatus{
			TxID:     sess.txID,
			SessID:   sess.sessID,
			RawQuery: query,
			Args:     args,
			Err:      err,
			Start:    start,
			End:      time.Now(),
			Context:  ctx,
		}
		queryLog(&status)
	}(time.Now())

	tx := sess.Transaction()

	if sess.Settings.PreparedStatementCacheEnabled() && tx == nil {
		var p *Stmt
		if p, query, args, err = sess.prepareStatement(ctx, stmt, args); err != nil {
			return nil, err
		}
		defer p.Close()

		rows, err = compat.PreparedQueryContext(p, ctx, args)
		return
	}

	query, args, err = sess.compileStatement(stmt, args)
	if err != nil {
		return nil, err
	}
	if tx != nil {
		rows, err = compat.QueryContext(tx, ctx, query, args)
		return
	}

	rows, err = compat.QueryContext(sess.sqlDB, ctx, query, args)
	return
}

// StatementQueryRow compiles and executes a statement that returns at most one
// row.
func (sess *sessionWithContext) StatementQueryRow(ctx context.Context, stmt *exql.Statement, args ...interface{}) (row *sql.Row, err error) {
	var query string

	defer func(start time.Time) {
		status := db.QueryStatus{
			TxID:     sess.txID,
			SessID:   sess.sessID,
			RawQuery: query,
			Args:     args,
			Err:      err,
			Start:    start,
			End:      time.Now(),
			Context:  ctx,
		}
		queryLog(&status)
	}(time.Now())

	tx := sess.Transaction()

	if sess.Settings.PreparedStatementCacheEnabled() && tx == nil {
		var p *Stmt
		if p, query, args, err = sess.prepareStatement(ctx, stmt, args); err != nil {
			return nil, err
		}
		defer p.Close()

		row = compat.PreparedQueryRowContext(p, ctx, args)
		return
	}

	query, args, err = sess.compileStatement(stmt, args)
	if err != nil {
		return nil, err
	}
	if tx != nil {
		row = compat.QueryRowContext(tx, ctx, query, args)
		return
	}

	row = compat.QueryRowContext(sess.sqlDB, ctx, query, args)
	return
}

// Driver returns the underlying *sql.DB or *sql.Tx instance.
func (sess *sessionWithContext) Driver() interface{} {
	if sess.sqlTx != nil {
		return sess.sqlTx
	}
	return sess.sqlDB
}

// compileStatement compiles the given statement into a string.
func (sess *sessionWithContext) compileStatement(stmt *exql.Statement, args []interface{}) (string, []interface{}, error) {
	for i := range args {
		args[i] = sess.ConvertValue(args[i])
	}
	if statementCompiler, ok := sess.adapter.(statementCompiler); ok {
		return statementCompiler.CompileStatement(sess, stmt, args)
	}

	compiled, err := stmt.Compile(sess.adapter.Template())
	if err != nil {
		return "", nil, err
	}
	query, args := sqlbuilder.Preprocess(compiled, args)
	return query, args, nil
}

// prepareStatement compiles a query and tries to use previously generated
// statement.
func (sess *sessionWithContext) prepareStatement(ctx context.Context, stmt *exql.Statement, args []interface{}) (*Stmt, string, []interface{}, error) {
	sess.sqlDBMu.Lock()
	defer sess.sqlDBMu.Unlock()

	sqlDB, tx := sess.sqlDB, sess.Transaction()
	if sqlDB == nil && tx == nil {
		return nil, "", nil, db.ErrNotConnected
	}

	pc, ok := sess.cachedStatements.ReadRaw(stmt)
	if ok {
		// The statement was cachesess.
		ps, err := pc.(*Stmt).Open()
		if err == nil {
			_, args, err = sess.compileStatement(stmt, args)
			if err != nil {
				return nil, "", nil, err
			}
			return ps, ps.query, args, nil
		}
	}

	query, args, err := sess.compileStatement(stmt, args)
	if err != nil {
		return nil, "", nil, err
	}
	sqlStmt, err := func(query *string) (*sql.Stmt, error) {
		if tx != nil {
			return compat.PrepareContext(tx, ctx, *query)
		}
		return compat.PrepareContext(sess.sqlDB, ctx, *query)
	}(&query)
	if err != nil {
		return nil, "", nil, err
	}

	p, err := NewStatement(sqlStmt, query).Open()
	if err != nil {
		return nil, query, args, err
	}
	sess.cachedStatements.Write(stmt, p)
	return p, p.query, args, nil
}

var waitForConnMu sync.Mutex

// WaitForConnection tries to execute the given connectFn function, if
// connectFn returns an error, then WaitForConnection will keep trying until
// connectFn returns nil. Maximum waiting time is 5s after having acquired the
// lock.
func (sess *sessionWithContext) WaitForConnection(connectFn func() error) error {
	// This lock ensures first-come, first-served and prevents opening too many
	// file descriptors.
	waitForConnMu.Lock()
	defer waitForConnMu.Unlock()

	// Minimum waiting time.
	waitTime := time.Millisecond * 10

	// Waitig 5 seconds for a successful connection.
	for timeStart := time.Now(); time.Since(timeStart) < time.Second*5; {
		err := connectFn()
		if err == nil {
			return nil // Connected!
		}

		// Only attempt to reconnect if the error is too many clients.
		if sess.Err(err) == db.ErrTooManyClients {
			// Sleep and try again if, and only if, the server replied with a "too
			// many clients" error.
			time.Sleep(waitTime)
			if waitTime < time.Millisecond*500 {
				// Wait a bit more next time.
				waitTime = waitTime * 2
			}
			continue
		}

		// Return any other error immediately.
		return err
	}

	return db.ErrGivingUpTryingToConnect
}

// ReplaceWithDollarSign turns a SQL statament with '?' placeholders into
// dollar placeholders, like $1, $2, ..., $n
func ReplaceWithDollarSign(buf []byte) []byte {
	z := bytes.Count(buf, []byte{'?'})
	// the capacity is a quick estimation of the total memory required, this
	// reduces reallocations
	out := make([]byte, 0, len(buf)+z*3)

	var i, k = 0, 1
	for i < len(buf) {
		if buf[i] == '?' {
			out = append(out, buf[:i]...)
			buf = buf[i+1:]
			i = 0

			if len(buf) > 0 && buf[0] == '?' {
				out = append(out, '?')
				buf = buf[1:]
				continue
			}

			out = append(out, '$')
			out = append(out, []byte(strconv.Itoa(k))...)
			k = k + 1
			continue
		}
		i = i + 1
	}

	out = append(out, buf[:len(buf)]...)
	buf = nil

	return out
}

func copySettings(from Session, into Session) {
	into.SetPreparedStatementCache(from.PreparedStatementCacheEnabled())
	into.SetConnMaxLifetime(from.ConnMaxLifetime())
	into.SetConnMaxIdleTime(from.ConnMaxIdleTime())
	into.SetMaxIdleConns(from.MaxIdleConns())
	into.SetMaxOpenConns(from.MaxOpenConns())
}

func newSessionID() uint64 {
	if atomic.LoadUint64(&lastSessID) == math.MaxUint64 {
		atomic.StoreUint64(&lastSessID, 0)
		return 0
	}
	return atomic.AddUint64(&lastSessID, 1)
}

func newBaseTxID() uint64 {
	if atomic.LoadUint64(&lastTxID) == math.MaxUint64 {
		atomic.StoreUint64(&lastTxID, 0)
		return 0
	}
	return atomic.AddUint64(&lastTxID, 1)
}

// TxContext creates a transaction context and runs fn within it.
func TxContext(ctx context.Context, sess db.Session, fn func(tx db.Session) error, opts *sql.TxOptions) error {
	txFn := func(sess db.Session) error {
		tx, err := sess.(Session).NewTransaction(ctx, opts)
		if err != nil {
			return err
		}
		defer tx.Close()

		if err := fn(tx); err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return fmt.Errorf("%v: %w", rollbackErr, err)
			}
			return err
		}
		return tx.Commit()
	}

	retryTime := retryTransactionWaitTime

	var txErr error
	for i := 0; i < sess.MaxTransactionRetries(); i++ {
		txErr = sess.(*sessionWithContext).Err(txFn(sess))
		if txErr == nil {
			return nil
		}
		if errors.Is(txErr, db.ErrTransactionAborted) {
			time.Sleep(retryTime)

			retryTime = retryTime * 2
			if retryTime > retryTransactionMaxWaitTime {
				retryTime = retryTransactionMaxWaitTime
			}

			continue
		}
		return txErr
	}

	return fmt.Errorf("db: giving up trying to commit transaction: %w", txErr)
}

var _ = db.Session(&sessionWithContext{})
