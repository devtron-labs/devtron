/*
 * Copyright (c) 2020-2024. Devtron Inc.
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
