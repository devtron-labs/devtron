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

package utils

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

var chars = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

// Generates random string
func Generate(size int) string {
	rand.Seed(time.Now().UnixNano())
	var b strings.Builder
	for i := 0; i < size; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	str := b.String()
	return str
}

func hasScheme(url string) bool {
	return len(url) >= 7 && (url[:7] == "http://" || url[:8] == "https://")
}

func GetUrlWithScheme(url string) (urlWithScheme string) {
	urlWithScheme = url
	if !hasScheme(url) {
		urlWithScheme = fmt.Sprintf("https://%s", url)
	}
	return urlWithScheme
}
