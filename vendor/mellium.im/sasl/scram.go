// Copyright 2016 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause license that can be
// found in the LICENSE file.

package sasl

import (
	"bytes"
	"crypto/hmac"
	"encoding/base64"
	"errors"
	"hash"
	"strconv"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

const (
	gs2HeaderCBSupport         = "p=tls-unique,"
	gs2HeaderNoServerCBSupport = "y,"
	gs2HeaderNoCBSupport       = "n,"
)

var (
	clientKeyInput = []byte("Client Key")
	serverKeyInput = []byte("Server Key")
)

// The number of random bytes to generate for a nonce.
const noncerandlen = 16

func getGS2Header(name string, n *Negotiator) (gs2Header []byte) {
	_, _, identity := n.Credentials()
	switch {
	case n.TLSState() == nil || !strings.HasSuffix(name, "-PLUS"):
		// We do not support channel binding
		gs2Header = []byte(gs2HeaderNoCBSupport)
	case n.State()&RemoteCB == RemoteCB:
		// We support channel binding and the server does too
		gs2Header = []byte(gs2HeaderCBSupport)
	case n.State()&RemoteCB != RemoteCB:
		// We support channel binding but the server does not
		gs2Header = []byte(gs2HeaderNoServerCBSupport)
	}
	if len(identity) > 0 {
		gs2Header = append(gs2Header, []byte(`a=`)...)
		gs2Header = append(gs2Header, identity...)
	}
	gs2Header = append(gs2Header, ',')
	return
}

func scram(name string, fn func() hash.Hash) Mechanism {
	// BUG(ssw): We need a way to cache the SCRAM client and server key
	// calculations.
	return Mechanism{
		Name: name,
		Start: func(m *Negotiator) (bool, []byte, interface{}, error) {
			user, _, _ := m.Credentials()

			// Escape "=" and ",". This is mostly the same as bytes.Replace but
			// faster because we can do both replacements in a single pass.
			n := bytes.Count(user, []byte{'='}) + bytes.Count(user, []byte{','})
			username := make([]byte, len(user)+(n*2))
			w := 0
			start := 0
			for i := 0; i < n; i++ {
				j := start
				j += bytes.IndexAny(user[start:], "=,")
				w += copy(username[w:], user[start:j])
				switch user[j] {
				case '=':
					w += copy(username[w:], "=3D")
				case ',':
					w += copy(username[w:], "=2C")
				}
				start = j + 1
			}
			copy(username[w:], user[start:])

			clientFirstMessage := make([]byte, 5+len(m.Nonce())+len(username))
			copy(clientFirstMessage, "n=")
			copy(clientFirstMessage[2:], username)
			copy(clientFirstMessage[2+len(username):], ",r=")
			copy(clientFirstMessage[5+len(username):], m.Nonce())

			return true, append(getGS2Header(name, m), clientFirstMessage...), clientFirstMessage, nil
		},
		Next: func(m *Negotiator, challenge []byte, data interface{}) (more bool, resp []byte, cache interface{}, err error) {
			if challenge == nil || len(challenge) == 0 {
				return more, resp, cache, ErrInvalidChallenge
			}

			if m.State()&Receiving == Receiving {
				panic("not yet implemented")
			}
			return scramClientNext(name, fn, m, challenge, data)
		},
	}
}

func scramClientNext(name string, fn func() hash.Hash, m *Negotiator, challenge []byte, data interface{}) (more bool, resp []byte, cache interface{}, err error) {
	_, password, _ := m.Credentials()
	state := m.State()

	switch state & StepMask {
	case AuthTextSent:
		iter := -1
		var salt, nonce []byte
		for _, field := range bytes.Split(challenge, []byte{','}) {
			if len(field) < 3 || (len(field) >= 2 && field[1] != '=') {
				continue
			}
			switch field[0] {
			case 'i':
				ival := string(bytes.TrimRight(field[2:], "\x00"))

				if iter, err = strconv.Atoi(ival); err != nil {
					return
				}
			case 's':
				salt = make([]byte, base64.StdEncoding.DecodedLen(len(field)-2))
				var n int
				n, err = base64.StdEncoding.Decode(salt, field[2:])
				salt = salt[:n]
				if err != nil {
					return
				}
			case 'r':
				nonce = field[2:]
			case 'm':
				// RFC 5802:
				// m: This attribute is reserved for future extensibility.  In this
				// version of SCRAM, its presence in a client or a server message
				// MUST cause authentication failure when the attribute is parsed by
				// the other end.
				err = errors.New("Server sent reserved attribute `m'")
				return
			}
		}

		switch {
		case iter < 0:
			err = errors.New("Iteration count is missing")
			return
		case iter < 0:
			err = errors.New("Iteration count is invalid")
			return
		case nonce == nil || !bytes.HasPrefix(nonce, m.Nonce()):
			err = errors.New("Server nonce does not match client nonce")
			return
		case salt == nil:
			err = errors.New("Server sent empty salt")
			return
		}

		gs2Header := getGS2Header(name, m)
		tlsState := m.TLSState()
		var channelBinding []byte
		if tlsState != nil && strings.HasSuffix(name, "-PLUS") {
			channelBinding = make(
				[]byte,
				2+base64.StdEncoding.EncodedLen(len(gs2Header)+len(tlsState.TLSUnique)),
			)
			base64.StdEncoding.Encode(channelBinding[2:], append(gs2Header, tlsState.TLSUnique...))
			channelBinding[0] = 'c'
			channelBinding[1] = '='
		} else {
			channelBinding = make(
				[]byte,
				2+base64.StdEncoding.EncodedLen(len(gs2Header)),
			)
			base64.StdEncoding.Encode(channelBinding[2:], gs2Header)
			channelBinding[0] = 'c'
			channelBinding[1] = '='
		}
		clientFinalMessageWithoutProof := append(channelBinding, []byte(",r=")...)
		clientFinalMessageWithoutProof = append(clientFinalMessageWithoutProof, nonce...)

		clientFirstMessage := data.([]byte)
		authMessage := append(clientFirstMessage, ',')
		authMessage = append(authMessage, challenge...)
		authMessage = append(authMessage, ',')
		authMessage = append(authMessage, clientFinalMessageWithoutProof...)

		saltedPassword := pbkdf2.Key(password, salt, iter, fn().Size(), fn)

		h := hmac.New(fn, saltedPassword)
		h.Write(serverKeyInput)
		serverKey := h.Sum(nil)
		h.Reset()

		h.Write(clientKeyInput)
		clientKey := h.Sum(nil)

		h = hmac.New(fn, serverKey)
		h.Write(authMessage)
		serverSignature := h.Sum(nil)

		h = fn()
		h.Write(clientKey)
		storedKey := h.Sum(nil)
		h = hmac.New(fn, storedKey)
		h.Write(authMessage)
		clientSignature := h.Sum(nil)
		clientProof := make([]byte, len(clientKey))
		xorBytes(clientProof, clientKey, clientSignature)

		encodedClientProof := make([]byte, base64.StdEncoding.EncodedLen(len(clientProof)))
		base64.StdEncoding.Encode(encodedClientProof, clientProof)
		clientFinalMessage := append(clientFinalMessageWithoutProof, []byte(",p=")...)
		clientFinalMessage = append(clientFinalMessage, encodedClientProof...)

		return true, clientFinalMessage, serverSignature, nil
	case ResponseSent:
		clientCalculatedServerFinalMessage := "v=" + base64.StdEncoding.EncodeToString(data.([]byte))
		if clientCalculatedServerFinalMessage != string(challenge) {
			err = ErrAuthn
			return
		}
		// Success!
		return false, nil, nil, nil
	}
	err = ErrInvalidState
	return
}
