//go:build pq
// +build pq

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

package postgresql

import (
	"net"
	"sort"
	"strings"
)

// String reassembles the parsed PostgreSQL connection URL into a valid DSN.
func (c ConnectionURL) String() (s string) {
	u := []string{}

	// TODO: This surely needs some sort of escaping.
	if c.User != "" {
		u = append(u, "user="+escaper.Replace(c.User))
	}

	if c.Password != "" {
		u = append(u, "password="+escaper.Replace(c.Password))
	}

	if c.Host != "" {
		host, port, err := net.SplitHostPort(c.Host)
		if err == nil {
			if host == "" {
				host = "127.0.0.1"
			}
			if port == "" {
				port = "5432"
			}
			u = append(u, "host="+escaper.Replace(host))
			u = append(u, "port="+escaper.Replace(port))
		} else {
			u = append(u, "host="+escaper.Replace(c.Host))
		}
	}

	if c.Socket != "" {
		u = append(u, "host="+escaper.Replace(c.Socket))
	}

	if c.Database != "" {
		u = append(u, "dbname="+escaper.Replace(c.Database))
	}

	// Is there actually any connection data?
	if len(u) == 0 {
		return ""
	}

	if c.Options == nil {
		c.Options = map[string]string{}
	}

	// If not present, SSL mode is assumed "prefer".
	if sslMode, ok := c.Options["sslmode"]; !ok || sslMode == "" {
		c.Options["sslmode"] = "prefer"
	}

	for k, v := range c.Options {
		u = append(u, escaper.Replace(k)+"="+escaper.Replace(v))
	}

	sort.Strings(u)

	return strings.Join(u, " ")
}
