/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package sql

import "github.com/go-pg/pg"

type TransactionWrapper interface {
	StartTx() (*pg.Tx, error)
	RollbackTx(tx *pg.Tx) error
	CommitTx(tx *pg.Tx) error
}

type TransactionUtilImpl struct {
	dbConnection *pg.DB
}

func NewTransactionUtilImpl(db *pg.DB) *TransactionUtilImpl {
	return &TransactionUtilImpl{
		dbConnection: db,
	}
}
func (impl *TransactionUtilImpl) RollbackTx(tx *pg.Tx) error {
	return tx.Rollback()
}
func (impl *TransactionUtilImpl) CommitTx(tx *pg.Tx) error {
	return tx.Commit()
}
func (impl *TransactionUtilImpl) StartTx() (*pg.Tx, error) {
	return impl.dbConnection.Begin()
}
