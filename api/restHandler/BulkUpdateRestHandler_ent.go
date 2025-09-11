/*
 * Copyright (c) 2024. Devtron Inc.
 */

package restHandler

import (
	"net/http"
)

type BulkEditV1Beta2RestHandlerEnt interface {
}

func (handler BulkUpdateRestHandlerImpl) BulkHibernateV1(w http.ResponseWriter, r *http.Request) {
	handler.BulkHibernate(w, r) // For backward compatibility, redirect to the new handler
}
