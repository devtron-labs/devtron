package cache

import (
	"fmt"

	"github.com/segmentio/fasthash/fnv1a"
)

const (
	hashTypeInt uint64 = 1 << iota
	hashTypeSignedInt
	hashTypeBool
	hashTypeString
	hashTypeHashable
	hashTypeNil
)

type hasher struct {
	t uint64
	v interface{}
}

func (h *hasher) Hash() uint64 {
	return NewHash(h.t, h.v)
}

func NewHashable(t uint64, v interface{}) Hashable {
	return &hasher{t: t, v: v}
}

func InitHash(t uint64) uint64 {
	return fnv1a.AddUint64(fnv1a.Init64, t)
}

func NewHash(t uint64, in ...interface{}) uint64 {
	return AddToHash(InitHash(t), in...)
}

func AddToHash(h uint64, in ...interface{}) uint64 {
	for i := range in {
		if in[i] == nil {
			continue
		}
		h = addToHash(h, in[i])
	}
	return h
}

func addToHash(h uint64, in interface{}) uint64 {
	switch v := in.(type) {
	case uint64:
		return fnv1a.AddUint64(fnv1a.AddUint64(h, hashTypeInt), v)
	case uint32:
		return fnv1a.AddUint64(fnv1a.AddUint64(h, hashTypeInt), uint64(v))
	case uint16:
		return fnv1a.AddUint64(fnv1a.AddUint64(h, hashTypeInt), uint64(v))
	case uint8:
		return fnv1a.AddUint64(fnv1a.AddUint64(h, hashTypeInt), uint64(v))
	case uint:
		return fnv1a.AddUint64(fnv1a.AddUint64(h, hashTypeInt), uint64(v))
	case int64:
		if v < 0 {
			return fnv1a.AddUint64(fnv1a.AddUint64(h, hashTypeSignedInt), uint64(-v))
		} else {
			return fnv1a.AddUint64(fnv1a.AddUint64(h, hashTypeInt), uint64(v))
		}
	case int32:
		if v < 0 {
			return fnv1a.AddUint64(fnv1a.AddUint64(h, hashTypeSignedInt), uint64(-v))
		} else {
			return fnv1a.AddUint64(fnv1a.AddUint64(h, hashTypeInt), uint64(v))
		}
	case int16:
		if v < 0 {
			return fnv1a.AddUint64(fnv1a.AddUint64(h, hashTypeSignedInt), uint64(-v))
		} else {
			return fnv1a.AddUint64(fnv1a.AddUint64(h, hashTypeInt), uint64(v))
		}
	case int8:
		if v < 0 {
			return fnv1a.AddUint64(fnv1a.AddUint64(h, hashTypeSignedInt), uint64(-v))
		} else {
			return fnv1a.AddUint64(fnv1a.AddUint64(h, hashTypeInt), uint64(v))
		}
	case int:
		if v < 0 {
			return fnv1a.AddUint64(fnv1a.AddUint64(h, hashTypeSignedInt), uint64(-v))
		} else {
			return fnv1a.AddUint64(fnv1a.AddUint64(h, hashTypeInt), uint64(v))
		}
	case bool:
		if v {
			return fnv1a.AddUint64(fnv1a.AddUint64(h, hashTypeBool), 1)
		} else {
			return fnv1a.AddUint64(fnv1a.AddUint64(h, hashTypeBool), 2)
		}
	case string:
		return fnv1a.AddString64(fnv1a.AddUint64(h, hashTypeString), v)
	case Hashable:
		if in == nil {
			panic(fmt.Sprintf("could not hash nil element %T", in))
		}
		return fnv1a.AddUint64(fnv1a.AddUint64(h, hashTypeHashable), v.Hash())
	case nil:
		return fnv1a.AddUint64(fnv1a.AddUint64(h, hashTypeNil), 0)
	default:
		panic(fmt.Sprintf("unsupported value type %T", in))
	}
}
