// Copyright 2024 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package sasl

var anonymous = Mechanism{
	Name: "ANONYMOUS",
	Start: func(*Negotiator) (bool, []byte, interface{}, error) {
		// Per XEP-0175 we do not send any trace data.
		return false, nil, nil, nil
	},
	Next: func(m *Negotiator, _ []byte, _ interface{}) (_ bool, _ []byte, _ interface{}, err error) {
		if m.State()&Receiving != Receiving || m.State()&AuthTextSent != AuthTextSent {
			err = ErrTooManySteps
		}
		return
	},
}
