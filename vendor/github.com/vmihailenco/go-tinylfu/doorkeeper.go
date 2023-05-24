package tinylfu

import (
	"math"
)

// doorkeeper is a small bloom-filter-based cache admission policy
type doorkeeper struct {
	m      uint32    // size of bit vector in bits
	k      uint32    // distinct hash functions needed
	filter bitvector // our filter bit vector
}

func newDoorkeeper(capacity int, falsePositiveRate float64) *doorkeeper {
	bits := float64(capacity) * -math.Log(falsePositiveRate) / (math.Log(2.0) * math.Log(2.0)) // in bits
	m := nextPowerOfTwo(uint32(bits))

	if m < 1024 {
		m = 1024
	}

	k := uint32(0.7 * float64(m) / float64(capacity))
	if k < 2 {
		k = 2
	}

	return &doorkeeper{
		m:      m,
		filter: newbv(m),
		k:      k,
	}
}

func (d *doorkeeper) allow(keyh uint64) bool {
	if d == nil {
		return true
	}
	alreadyPresent := d.insert(keyh)
	return alreadyPresent
}

// insert inserts the byte array b into the bloom filter.  Returns true if the value
// was already considered to be in the bloom filter.
func (d *doorkeeper) insert(h uint64) bool {
	h1, h2 := uint32(h), uint32(h>>32)
	var o uint = 1
	for i := uint32(0); i < d.k; i++ {
		o &= d.filter.getset((h1 + (i * h2)) & (d.m - 1))
	}
	return o == 1
}

// Reset clears the bloom filter
func (d *doorkeeper) reset() {
	if d == nil {
		return
	}
	for i := range d.filter {
		d.filter[i] = 0
	}
}

// Internal routines for the bit vector
type bitvector []uint64

func newbv(size uint32) bitvector {
	return make([]uint64, uint(size+63)/64)
}

// set bit 'bit' in the bitvector d and return previous value
func (b bitvector) getset(bit uint32) uint {
	shift := bit % 64
	idx := bit / 64
	bb := b[idx]
	m := uint64(1) << shift
	b[idx] |= m
	return uint((bb & m) >> shift)
}

// return the integer >= i which is a power of two
func nextPowerOfTwo(i uint32) uint32 {
	n := i - 1
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n++
	return n
}
