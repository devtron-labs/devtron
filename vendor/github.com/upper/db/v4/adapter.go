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
	"fmt"
	"sync"
)

var (
	adapterMap   = make(map[string]Adapter)
	adapterMapMu sync.RWMutex
)

// Adapter interface defines an adapter
type Adapter interface {
	Open(ConnectionURL) (Session, error)
}

type missingAdapter struct {
	name string
}

func (ma *missingAdapter) Open(ConnectionURL) (Session, error) {
	return nil, fmt.Errorf("upper: Missing adapter %q, did you forget to import it?", ma.name)
}

// RegisterAdapter registers a generic database adapter.
func RegisterAdapter(name string, adapter Adapter) {
	adapterMapMu.Lock()
	defer adapterMapMu.Unlock()

	if name == "" {
		panic(`Missing adapter name`)
	}
	if _, ok := adapterMap[name]; ok {
		panic(`db.RegisterAdapter() called twice for adapter: ` + name)
	}
	adapterMap[name] = adapter
}

// LookupAdapter returns a previously registered adapter by name.
func LookupAdapter(name string) Adapter {
	adapterMapMu.RLock()
	defer adapterMapMu.RUnlock()

	if adapter, ok := adapterMap[name]; ok {
		return adapter
	}
	return &missingAdapter{name: name}
}

// Open attempts to stablish a connection with a database.
func Open(adapterName string, settings ConnectionURL) (Session, error) {
	return LookupAdapter(adapterName).Open(settings)
}
