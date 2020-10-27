/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package sse

type SSEMessage struct {
	Event     string
	Data      []byte
	Namespace string
}

func (msg SSEMessage) format() []byte {
	res := make([]byte, 0, 6+5+len(msg.Event)+len(msg.Data)+3)
	if msg.Event != "" {
		res = append(res, "event:"...)
		res = append(res, msg.Event...)
		res = append(res, '\n')
	}
	res = append(res, "data:"...)
	res = append(res, msg.Data...)
	res = append(res, '\n', '\n')
	return res
}
