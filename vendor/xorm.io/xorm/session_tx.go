// Copyright 2016 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"time"

	"xorm.io/xorm/log"
)

// Begin a transaction
func (session *Session) Begin() error {
	if session.isAutoCommit {
		tx, err := session.DB().BeginTx(session.ctx, nil)
		if err != nil {
			return err
		}
		session.isAutoCommit = false
		session.isCommitedOrRollbacked = false
		session.tx = tx

		session.saveLastSQL("BEGIN TRANSACTION")
	}
	return nil
}

// Rollback When using transaction, you can rollback if any error
func (session *Session) Rollback() error {
	if !session.isAutoCommit && !session.isCommitedOrRollbacked {
		session.saveLastSQL("ROLL BACK")
		session.isCommitedOrRollbacked = true
		session.isAutoCommit = true

		start := time.Now()
		needSQL := session.DB().NeedLogSQL(session.ctx)
		if needSQL {
			session.engine.logger.BeforeSQL(log.LogContext{
				Ctx: session.ctx,
				SQL: "ROLL BACK",
			})
		}
		err := session.tx.Rollback()
		if needSQL {
			session.engine.logger.AfterSQL(log.LogContext{
				Ctx:         session.ctx,
				SQL:         "ROLL BACK",
				ExecuteTime: time.Now().Sub(start),
				Err:         err,
			})
		}
		return err
	}
	return nil
}

// Commit When using transaction, Commit will commit all operations.
func (session *Session) Commit() error {
	if !session.isAutoCommit && !session.isCommitedOrRollbacked {
		session.saveLastSQL("COMMIT")
		session.isCommitedOrRollbacked = true
		session.isAutoCommit = true

		start := time.Now()
		needSQL := session.DB().NeedLogSQL(session.ctx)
		if needSQL {
			session.engine.logger.BeforeSQL(log.LogContext{
				Ctx: session.ctx,
				SQL: "COMMIT",
			})
		}
		err := session.tx.Commit()
		if needSQL {
			session.engine.logger.AfterSQL(log.LogContext{
				Ctx:         session.ctx,
				SQL:         "COMMIT",
				ExecuteTime: time.Now().Sub(start),
				Err:         err,
			})
		}

		if err != nil {
			return err
		}

		// handle processors after tx committed
		closureCallFunc := func(closuresPtr *[]func(interface{}), bean interface{}) {
			if closuresPtr != nil {
				for _, closure := range *closuresPtr {
					closure(bean)
				}
			}
		}

		for bean, closuresPtr := range session.afterInsertBeans {
			closureCallFunc(closuresPtr, bean)

			if processor, ok := interface{}(bean).(AfterInsertProcessor); ok {
				processor.AfterInsert()
			}
		}
		for bean, closuresPtr := range session.afterUpdateBeans {
			closureCallFunc(closuresPtr, bean)

			if processor, ok := interface{}(bean).(AfterUpdateProcessor); ok {
				processor.AfterUpdate()
			}
		}
		for bean, closuresPtr := range session.afterDeleteBeans {
			closureCallFunc(closuresPtr, bean)

			if processor, ok := interface{}(bean).(AfterDeleteProcessor); ok {
				processor.AfterDelete()
			}
		}
		cleanUpFunc := func(slices *map[interface{}]*[]func(interface{})) {
			if len(*slices) > 0 {
				*slices = make(map[interface{}]*[]func(interface{}), 0)
			}
		}
		cleanUpFunc(&session.afterInsertBeans)
		cleanUpFunc(&session.afterUpdateBeans)
		cleanUpFunc(&session.afterDeleteBeans)
	}
	return nil
}
