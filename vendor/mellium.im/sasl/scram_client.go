// Copyright 2016 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package sasl

import (
	"bytes"
	"crypto/hmac"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"hash"
	"strconv"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

func scramClientNext(name string, fn func() hash.Hash, m *Negotiator, challenge []byte, data interface{}) (more bool, resp []byte, cache interface{}, err error) {
	_, password, _ := m.Credentials()
	state := m.State()

	switch state & StepMask {
	case AuthTextSent:
		iter := -1
		var salt, nonce []byte
		remain := challenge
		for {
			var field []byte
			field, remain = nextParam(remain)
			if len(field) < 3 || (len(field) >= 2 && field[1] != '=') {
				continue
			}
			switch field[0] {
			case 'i':
				ival := string(bytes.TrimRight(field[2:], "\x00"))

				if iter, err = strconv.Atoi(ival); err != nil {
					return false, nil, nil, err
				}
			case 's':
				salt = make([]byte, base64.StdEncoding.DecodedLen(len(field)-2))
				var n int
				n, err = base64.StdEncoding.Decode(salt, field[2:])
				salt = salt[:n]
				if err != nil {
					return false, nil, nil, err
				}
			case 'r':
				nonce = field[2:]
			case 'm':
				// RFC 5802:
				// m: This attribute is reserved for future extensibility.  In this
				// version of SCRAM, its presence in a client or a server message
				// MUST cause authentication failure when the attribute is parsed by
				// the other end.
				err = errors.New("server sent reserved attribute `m'")
				return false, nil, nil, err
			}
			if remain == nil {
				break
			}
		}

		switch {
		case iter < 0:
			err = errors.New("iteration count is invalid")
			return false, nil, nil, err
		case nonce == nil || !bytes.HasPrefix(nonce, m.Nonce()):
			err = errors.New("server nonce does not match client nonce")
			return false, nil, nil, err
		case salt == nil:
			err = errors.New("server sent empty salt")
			return false, nil, nil, err
		}

		gs2Header := getGS2Header(name, m)
		tlsState := m.TLSState()
		var channelBinding []byte
		switch plus := strings.HasSuffix(name, "-PLUS"); {
		case plus && tlsState == nil:
			err = errors.New("sasl: SCRAM with channel binding requires a TLS connection")
			return false, nil, nil, err
		case bytes.Contains(gs2Header, []byte(gs2HeaderCBSupportExporter)):
			keying, err := tlsState.ExportKeyingMaterial(exporterLabel, nil, exporterLen)
			if err != nil {
				return false, nil, nil, err
			}
			if len(keying) == 0 {
				err = errors.New("sasl: SCRAM with channel binding requires valid TLS keying material")
				return false, nil, nil, err
			}
			channelBinding = make([]byte, 2+base64.StdEncoding.EncodedLen(len(gs2Header)+len(keying)))
			channelBinding[0] = 'c'
			channelBinding[1] = '='
			base64.StdEncoding.Encode(channelBinding[2:], append(gs2Header, keying...))
		case bytes.Contains(gs2Header, []byte(gs2HeaderCBSupportUnique)):
			if len(tlsState.TLSUnique) == 0 {
				err = errors.New("sasl: SCRAM with channel binding requires valid tls-unique data")
				return false, nil, nil, err
			}
			channelBinding = make(
				[]byte,
				2+base64.StdEncoding.EncodedLen(len(gs2Header)+len(tlsState.TLSUnique)),
			)
			channelBinding[0] = 'c'
			channelBinding[1] = '='
			base64.StdEncoding.Encode(channelBinding[2:], append(gs2Header, tlsState.TLSUnique...))
		default:
			channelBinding = make(
				[]byte,
				2+base64.StdEncoding.EncodedLen(len(gs2Header)),
			)
			channelBinding[0] = 'c'
			channelBinding[1] = '='
			base64.StdEncoding.Encode(channelBinding[2:], gs2Header)
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
		_, err = h.Write(serverKeyInput)
		if err != nil {
			return false, nil, nil, err
		}
		serverKey := h.Sum(nil)
		h.Reset()

		_, err = h.Write(clientKeyInput)
		if err != nil {
			return false, nil, nil, err
		}
		clientKey := h.Sum(nil)

		h = hmac.New(fn, serverKey)
		_, err = h.Write(authMessage)
		if err != nil {
			return false, nil, nil, err
		}
		serverSignature := h.Sum(nil)

		h = fn()
		_, err = h.Write(clientKey)
		if err != nil {
			return false, nil, nil, err
		}
		storedKey := h.Sum(nil)
		h = hmac.New(fn, storedKey)
		_, err = h.Write(authMessage)
		if err != nil {
			return false, nil, nil, err
		}
		clientSignature := h.Sum(nil)
		clientProof := make([]byte, len(clientKey))
		subtle.XORBytes(clientProof, clientKey, clientSignature)

		encodedClientProof := make([]byte, base64.StdEncoding.EncodedLen(len(clientProof)))
		base64.StdEncoding.Encode(encodedClientProof, clientProof)
		clientFinalMessage := append(clientFinalMessageWithoutProof, []byte(",p=")...)
		clientFinalMessage = append(clientFinalMessage, encodedClientProof...)

		return true, clientFinalMessage, serverSignature, nil
	case ResponseSent:
		clientCalculatedServerFinalMessage := "v=" + base64.StdEncoding.EncodeToString(data.([]byte))
		if clientCalculatedServerFinalMessage != string(challenge) {
			return false, nil, nil, ErrAuthn
		}
		// Success!
		return false, nil, nil, nil
	default:
		return false, nil, nil, ErrInvalidState
	}
}
