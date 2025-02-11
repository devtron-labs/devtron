// Copyright (c) 2012-present The upper.io/db authors. All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
// OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
// WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package db

import (
	"sync"
	"sync/atomic"
	"time"
)

// Settings defines methods to get or set configuration values.
type Settings interface {
	// SetPreparedStatementCache enables or disables the prepared statement
	// cache.
	SetPreparedStatementCache(bool)

	// PreparedStatementCacheEnabled returns true if the prepared statement cache
	// is enabled, false otherwise.
	PreparedStatementCacheEnabled() bool

	// SetConnMaxLifetime sets the default maximum amount of time a connection
	// may be reused.
	SetConnMaxLifetime(time.Duration)

	// ConnMaxLifetime returns the default maximum amount of time a connection
	// may be reused.
	ConnMaxLifetime() time.Duration

	// SetConnMaxIdleTime sets the default maximum amount of time a connection
	// may remain idle.
	SetConnMaxIdleTime(time.Duration)

	// ConnMaxIdleTime returns the default maximum amount of time a connection
	// may remain idle.
	ConnMaxIdleTime() time.Duration

	// SetMaxIdleConns sets the default maximum number of connections in the idle
	// connection pool.
	SetMaxIdleConns(int)

	// MaxIdleConns returns the default maximum number of connections in the idle
	// connection pool.
	MaxIdleConns() int

	// SetMaxOpenConns sets the default maximum number of open connections to the
	// database.
	SetMaxOpenConns(int)

	// MaxOpenConns returns the default maximum number of open connections to the
	// database.
	MaxOpenConns() int

	// SetMaxTransactionRetries sets the number of times a transaction can
	// be retried.
	SetMaxTransactionRetries(int)

	// MaxTransactionRetries returns the maximum number of times a
	// transaction can be retried.
	MaxTransactionRetries() int
}

type settings struct {
	sync.RWMutex

	preparedStatementCacheEnabled uint32

	connMaxLifetime time.Duration
	connMaxIdleTime time.Duration
	maxOpenConns    int
	maxIdleConns    int

	maxTransactionRetries int
}

func (c *settings) binaryOption(opt *uint32) bool {
	return atomic.LoadUint32(opt) == 1
}

func (c *settings) setBinaryOption(opt *uint32, value bool) {
	if value {
		atomic.StoreUint32(opt, 1)
		return
	}
	atomic.StoreUint32(opt, 0)
}

func (c *settings) SetPreparedStatementCache(value bool) {
	c.setBinaryOption(&c.preparedStatementCacheEnabled, value)
}

func (c *settings) PreparedStatementCacheEnabled() bool {
	return c.binaryOption(&c.preparedStatementCacheEnabled)
}

func (c *settings) SetConnMaxLifetime(t time.Duration) {
	c.Lock()
	c.connMaxLifetime = t
	c.Unlock()
}

func (c *settings) ConnMaxLifetime() time.Duration {
	c.RLock()
	defer c.RUnlock()
	return c.connMaxLifetime
}

func (c *settings) SetConnMaxIdleTime(t time.Duration) {
	c.Lock()
	c.connMaxIdleTime = t
	c.Unlock()
}

func (c *settings) ConnMaxIdleTime() time.Duration {
	c.RLock()
	defer c.RUnlock()
	return c.connMaxIdleTime
}

func (c *settings) SetMaxIdleConns(n int) {
	c.Lock()
	c.maxIdleConns = n
	c.Unlock()
}

func (c *settings) MaxIdleConns() int {
	c.RLock()
	defer c.RUnlock()
	return c.maxIdleConns
}

func (c *settings) SetMaxTransactionRetries(n int) {
	c.Lock()
	c.maxTransactionRetries = n
	c.Unlock()
}

func (c *settings) MaxTransactionRetries() int {
	c.RLock()
	defer c.RUnlock()
	if c.maxTransactionRetries < 1 {
		return 1
	}
	return c.maxTransactionRetries
}

func (c *settings) SetMaxOpenConns(n int) {
	c.Lock()
	c.maxOpenConns = n
	c.Unlock()
}

func (c *settings) MaxOpenConns() int {
	c.RLock()
	defer c.RUnlock()
	return c.maxOpenConns
}

// NewSettings returns a new settings value prefilled with the current default
// settings.
func NewSettings() Settings {
	def := DefaultSettings.(*settings)
	return &settings{
		preparedStatementCacheEnabled: def.preparedStatementCacheEnabled,
		connMaxLifetime:               def.connMaxLifetime,
		connMaxIdleTime:               def.connMaxIdleTime,
		maxIdleConns:                  def.maxIdleConns,
		maxOpenConns:                  def.maxOpenConns,
		maxTransactionRetries:         def.maxTransactionRetries,
	}
}

// DefaultSettings provides default global configuration settings for database
// sessions.
var DefaultSettings Settings = &settings{
	preparedStatementCacheEnabled: 0,
	connMaxLifetime:               time.Duration(0),
	connMaxIdleTime:               time.Duration(0),
	maxIdleConns:                  10,
	maxOpenConns:                  0,
	maxTransactionRetries:         1,
}
