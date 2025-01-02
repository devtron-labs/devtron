// Copyright 2016 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package sasl

import (
	"bytes"
	"crypto/hmac"
	"crypto/subtle"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"hash"
	"strconv"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

const (
	exporterLen                = 32
	exporterLabel              = "EXPORTER-Channel-Binding"
	gs2HeaderCBSupportUnique   = "p=tls-unique,"
	gs2HeaderCBSupportExporter = "p=tls-exporter,"
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
	tlsState := n.TLSState()
	switch {
	case tlsState == nil || !strings.HasSuffix(name, "-PLUS"):
		// We do not support channel binding
		gs2Header = []byte(gs2HeaderNoCBSupport)
	case n.State()&RemoteCB == RemoteCB:
		// We support channel binding and the server does too
		if tlsState.Version >= tls.VersionTLS13 {
			gs2Header = []byte(gs2HeaderCBSupportExporter)
		} else {
			gs2Header = []byte(gs2HeaderCBSupportUnique)
		}
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

			username := escapeSaslname(user)
			clientFirstMessage := concatenateByteSlices(
				[]byte("n="), username,
				[]byte(",r="), m.Nonce(),
			)

			return true, append(getGS2Header(name, m), clientFirstMessage...), clientFirstMessage, nil
		},
		Next: func(m *Negotiator, challenge []byte, data interface{}) (more bool, resp []byte, cache interface{}, err error) {
			if len(challenge) == 0 {
				return false, nil, nil, ErrInvalidChallenge
			}

			if m.State()&Receiving == Receiving {
				return scramServerNext(name, fn, m, challenge, data)
			} else {
				return scramClientNext(name, fn, m, challenge, data)
			}
		},
	}
}

func scramServerNext(name string, fn func() hash.Hash, m *Negotiator, challenge []byte, data interface{}) (more bool, resp []byte, cache interface{}, err error) {
	if strings.HasSuffix(name, "-PLUS") {
		panic("SCRAM *-PLUS mechanism does not implemented yet")
	}

	type scramServerCache struct {
		gs2Header              []byte
		nonce                  []byte
		clientFirstMessageBare []byte
		serverFirstMessage     []byte

		storedKey []byte
		serverKey []byte
	}

	switch m.State() & StepMask {
	case AuthTextSent:
		var message *clientFirstMessage

		message, err = parseClientFirstMessage(challenge)
		if err != nil {
			return false, nil, nil, fmt.Errorf("cannot parse %q as client first message: %w", challenge, err)
		}

		if bytes.HasPrefix(message.gs2CbindFlag, []byte("p=")) {
			return false, nil, nil, fmt.Errorf("client wants binding, but it is not implemented")
		}

		// RFC 5802, section 5.1. SCRAM Attributes, n attribute
		if bytes.Contains(message.username, []byte{'='}) {
			return false, nil, nil, fmt.Errorf("unescaped username contains '='")
		}

		var salt, saltedPassword []byte
		var iter int64
		salt, saltedPassword, iter, err = m.SaltedCredentials(message.username, message.authzID)
		if err != nil {
			return false, nil, nil, err
		}

		nonce := append(message.nonce, m.Nonce()...)

		clientKey := scramClientKey(fn, saltedPassword)
		storedKey := scramStoredKey(fn, clientKey)
		serverKey := scramServerKey(fn, saltedPassword)

		saltEncoded := make([]byte, 3+base64.StdEncoding.EncodedLen(len(salt)))
		copy(saltEncoded[:3], ",s=")
		base64.StdEncoding.Encode(saltEncoded[3:], salt)
		resp = concatenateByteSlices(
			[]byte("r="), nonce,
			saltEncoded,
			[]byte(",i="), []byte(strconv.FormatInt(int64(iter), 10)),
		)
		cache = scramServerCache{
			nonce:                  nonce,
			gs2Header:              message.gs2Header,
			clientFirstMessageBare: message.bare,
			serverFirstMessage:     resp,
			storedKey:              storedKey,
			serverKey:              serverKey,
		}
		return true, resp, cache, nil

	case ResponseSent:
		var message *clientFinalMessage
		message, err = parseClientFinalMessage(challenge)
		if err != nil {
			return false, []byte("e=invalid-encoding"), nil, fmt.Errorf("cannot parse %q as client final message: %w", challenge, err)
		}

		serverCache := data.(scramServerCache)

		if !bytes.Equal(message.channelBinding, serverCache.gs2Header) {
			return false, []byte("e=channel-bindings-dont-match"), nil, fmt.Errorf("expected %q channel-binding, but client sent %q", serverCache.gs2Header, message.channelBinding)
		}

		if !bytes.Equal(message.nonce, serverCache.nonce) {
			return false, []byte("e=other-error"), nil, fmt.Errorf("nonce expected to be %q, but client sent %q", serverCache.nonce, message.nonce)
		}

		authMessage := concatenateByteSlices(
			serverCache.clientFirstMessageBare, []byte{','},
			serverCache.serverFirstMessage, []byte{','},
			message.messageWithoutProof,
		)

		clientSignature := scramClientSignature(fn, serverCache.storedKey, authMessage)
		clientKey := make([]byte, len(message.proof))
		subtle.XORBytes(clientKey, message.proof, clientSignature)

		storedKey := scramStoredKey(fn, clientKey)
		if !hmac.Equal(storedKey, serverCache.storedKey) {
			return false, []byte("e=invalid-proof"), nil, fmt.Errorf("StoredKey in server (%q) is not equal to one calculated from proof (%q)", serverCache.storedKey, storedKey)
		}

		serverSignature := scramServerSignature(fn, serverCache.serverKey, authMessage)

		// Success!
		resp = []byte("v=" + base64.StdEncoding.EncodeToString(serverSignature))
		return false, resp, nil, nil
	default:
		return false, resp, nil, ErrInvalidState
	}
}

// SCRAMSaltPassword calculates salted version of the raw password using
// fn as hash function, salt and iter as number of iterations.
//
// See RFC 5802, section 3. SCRAM Algorithm Overview, for implementation details.
func SCRAMSaltPassword(fn func() hash.Hash, password []byte, salt []byte, iter int) []byte {
	return pbkdf2.Key(password, salt, iter, fn().Size(), fn)
}

// RFC 5802, section 3. SCRAM Algorithm Overview
// ClientKey       := HMAC(SaltedPassword, "Client Key")
func scramClientKey(fn func() hash.Hash, saltedPassword []byte) []byte {
	h := hmac.New(fn, saltedPassword)
	/* #nosec */
	_, _ = h.Write(clientKeyInput)
	return h.Sum(nil)
}

// RFC 5802, section 3. SCRAM Algorithm Overview
// StoredKey       := H(ClientKey)
func scramStoredKey(fn func() hash.Hash, clientKey []byte) []byte {
	h := fn()
	/* #nosec */
	_, _ = h.Write(clientKey)
	return h.Sum(nil)
}

// RFC 5802, section 3. SCRAM Algorithm Overview
// ClientSignature := HMAC(StoredKey, AuthMessage)
func scramClientSignature(fn func() hash.Hash, storedKey []byte, authMessage []byte) []byte {
	h := hmac.New(fn, storedKey)
	/* #nosec */
	_, _ = h.Write(authMessage)
	return h.Sum(nil)
}

// RFC 5802, section 3. SCRAM Algorithm Overview
// ServerKey       := HMAC(SaltedPassword, "Server Key")
func scramServerKey(fn func() hash.Hash, saltedPassword []byte) []byte {
	h := hmac.New(fn, saltedPassword)
	/* #nosec */
	_, _ = h.Write(serverKeyInput)
	return h.Sum(nil)
}

// RFC 5802, section 3. SCRAM Algorithm Overview
// ServerSignature := HMAC(ServerKey, AuthMessage)
func scramServerSignature(fn func() hash.Hash, serverKey []byte, authMessage []byte) []byte {
	h := hmac.New(fn, serverKey)
	/* #nosec */
	_, _ = h.Write(authMessage)
	return h.Sum(nil)
}

// Replace ',' and '=' with "=2C" and "=3D" respectively.
//
// escapeSaslname() does 1 allocation and have O(len(unescaped))
// time complexity.
// RFC 5802, section 5.1. SCRAM Attributes, n attribute
func escapeSaslname(unescaped []byte) []byte {
	n := bytes.Count(unescaped, []byte{'='}) + bytes.Count(unescaped, []byte{','})
	escaped := make([]byte, len(unescaped)+(n*2))
	w := 0
	start := 0
	for i := 0; i < n; i++ {
		j := start
		j += bytes.IndexAny(unescaped[start:], "=,")
		w += copy(escaped[w:], unescaped[start:j])
		switch unescaped[j] {
		case '=':
			w += copy(escaped[w:], "=3D")
		case ',':
			w += copy(escaped[w:], "=2C")
		}
		start = j + 1
	}
	copy(escaped[w:], unescaped[start:])
	return escaped
}

// Replace "=2C" and "=3D" with ',' and '=' respectively.
//
// unescapeSaslname() does not allocate memory and have O(len(escaped))
// time complexity.
// RFC 5802, section 5.1. SCRAM Attributes, n attribute
func unescapeSaslname(escaped []byte) []byte {
	j := 0
	i := 0
	for i < len(escaped) {
		if escaped[i] == '=' {
			if len(escaped)-i >= 3 {
				if escaped[i+1] == '2' && escaped[i+2] == 'C' {
					escaped[j] = ','
					i += 2
				} else if escaped[i+1] == '3' && escaped[i+2] == 'D' {
					escaped[j] = '='
					i += 2
				} else {
					escaped[j] = escaped[i]
				}
			} else {
				escaped[j] = escaped[i]
			}
		} else {
			escaped[j] = escaped[i]
		}
		j++
		i++
	}
	return escaped[:j]
}

type clientFirstMessage struct {
	gs2CbindFlag []byte
	authzID      []byte
	username     []byte
	nonce        []byte

	gs2Header []byte
	bare      []byte
}

// RFC 5802, section 7. Formal Syntax
// gs2-cbind-flag  	= ("p=" cb-name) / "n" / "y"
// gs2-header      	= gs2-cbind-flag "," [ authzid ] ","
// authzid         	= "a=" saslname
// reserved-mext  	= "m=" 1*(value-char)
// username        	= "n=" saslname
// nonce           	= "r=" c-nonce [s-nonce]
// extensions 		= attr-val *("," attr-val)
// client-first-message-bare 	= [reserved-mext ","] username "," nonce ["," extensions]
// client-first-message 		= gs2-header client-first-message-bare
// TODO: parse extensions
func parseClientFirstMessage(challenge []byte) (*clientFirstMessage, error) {
	var message clientFirstMessage
	minFields := 4
	fields := bytes.Split(challenge, []byte{','})
	if len(fields) < minFields {
		return nil, fmt.Errorf("expected at least %d fields, got %d", minFields, len(fields))
	}

	// gs2-cbind-flag
	if len(fields[0]) == 1 {
		if fields[0][0] != 'y' && fields[0][0] != 'n' {
			return nil, fmt.Errorf("%q is invalid gs2-cbind-flag", fields[0])
		}
	} else {
		if !bytes.HasPrefix(fields[0], []byte("p=")) {
			return nil, fmt.Errorf("%q is invalid gs2-cbind-flag", fields[0])
		}
	}
	message.gs2CbindFlag = fields[0]

	// authzid
	if len(fields[1]) > 0 {
		if !bytes.HasPrefix(fields[1], []byte("a=")) {
			return nil, fmt.Errorf("%q is invalid authzid", fields[1])
		}
		message.authzID = fields[1][2:]
	}

	// reserved-mext
	if bytes.HasPrefix(fields[2], []byte("m=")) {
		return nil, errors.New("SCRAM message extensions are not supported")
	}

	// username
	if !bytes.HasPrefix(fields[2], []byte("n=")) {
		return nil, fmt.Errorf("%q is invalid username", fields[2])
	}
	message.username = unescapeSaslname(fields[2][2:])

	// nonce
	if !bytes.HasPrefix(fields[3], []byte("r=")) {
		return nil, fmt.Errorf("%q is invalid nonce", fields[3])
	}
	message.nonce = fields[3][2:]

	switch {
	case len(message.username) == 0:
		return nil, fmt.Errorf("got empty username")
	case len(message.nonce) == 0:
		return nil, fmt.Errorf("got empty nonce")
	}

	message.gs2Header = bytes.Join(fields[:2], []byte{','})
	message.gs2Header = append(message.gs2Header, ',')
	message.bare = bytes.Join(fields[2:], []byte{','})
	return &message, nil
}

type clientFinalMessage struct {
	channelBinding []byte
	nonce          []byte
	proof          []byte

	messageWithoutProof []byte
}

// RFC 5802, section 7. Formal Syntax
// attr-val        = ALPHA "=" value
// channel-binding = "c=" base64
// proof           = "p=" base64
// nonce           = "r=" c-nonce [s-nonce]
// extensions 	   = attr-val *("," attr-val)
// client-final-message-without-proof = channel-binding "," nonce ["," extensions]
// client-final-message 			  = client-final-message-without-proof "," proof
func parseClientFinalMessage(challenge []byte) (*clientFinalMessage, error) {
	var message clientFinalMessage
	var err error

	remain := challenge
	for fieldIndex := 0; len(remain) > 0; fieldIndex++ {
		var field []byte
		field, remain = nextParam(remain)
		switch fieldIndex {
		case 0:
			if bytes.HasPrefix(field, []byte("c=")) {
				message.channelBinding, err = base64.StdEncoding.DecodeString(string(field[2:]))
				if err != nil {
					return nil, fmt.Errorf("cannot decode %q: %w", field, err)
				}
			} else {
				return nil, fmt.Errorf("expected channel-binding (c=...) as 1st field, got %q", field)
			}
		case 1:
			if bytes.HasPrefix(field, []byte("r=")) {
				message.nonce = field[2:]
			} else {
				return nil, fmt.Errorf("expected nonce (r=...) as 2nd field, got %q", field)
			}
		default:
			if bytes.HasPrefix(field, []byte("p=")) {
				if len(remain) > 0 {
					return nil, fmt.Errorf("expected proof (p=...) to be last field")
				}
				message.proof, err = base64.StdEncoding.DecodeString(string(field[2:]))
				if err != nil {
					return nil, fmt.Errorf("cannot decode %q: %w", field, err)
				}
			}
		}
	}
	switch {
	case len(message.channelBinding) == 0:
		return nil, fmt.Errorf("got empty channel-binding (c=...)")
	case len(message.nonce) == 0:
		return nil, fmt.Errorf("got empty nonce (r=...)")
	case len(message.proof) == 0:
		return nil, fmt.Errorf("got empty proof (p=...)")
	}
	message.messageWithoutProof = challenge[:bytes.LastIndexByte(challenge, ',')]
	return &message, nil
}

func nextParam(params []byte) ([]byte, []byte) {
	idx := bytes.IndexByte(params, ',')
	if idx == -1 {
		return params, nil
	}
	return params[:idx], params[idx+1:]
}

// Concatenate pieces
//
// concatenateByteSlices() does 1 allocation and have O(len(pieces)...)
// time complexity, where len(pieces)... is the sum of all the peaces lengths.
func concatenateByteSlices(pieces ...[]byte) []byte {
	var resultLen int
	for _, pice := range pieces {
		resultLen += len(pice)
	}

	var i int
	result := make([]byte, resultLen)
	for _, pice := range pieces {
		i += copy(result[i:], pice)
	}
	return result
}
