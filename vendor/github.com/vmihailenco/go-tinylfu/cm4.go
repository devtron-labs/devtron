package tinylfu

// cm4 is a small conservative-update count-min sketch implementation with 4-bit counters
type cm4 struct {
	s    [depth]nvec
	mask uint32
}

const depth = 4

func newCM4(w int) *cm4 {
	if w < 1 {
		panic("cm4: bad width")
	}

	w32 := nextPowerOfTwo(uint32(w))
	c := cm4{
		mask: w32 - 1,
	}

	for i := 0; i < depth; i++ {
		c.s[i] = newNvec(int(w32))
	}

	return &c
}

func (c *cm4) add(keyh uint64) {
	h1, h2 := uint32(keyh), uint32(keyh>>32)

	for i := range c.s {
		pos := (h1 + uint32(i)*h2) & c.mask
		c.s[i].inc(pos)
	}
}

func (c *cm4) estimate(keyh uint64) byte {
	h1, h2 := uint32(keyh), uint32(keyh>>32)

	var min byte = 255
	for i := 0; i < depth; i++ {
		pos := (h1 + uint32(i)*h2) & c.mask
		v := c.s[i].get(pos)
		if v < min {
			min = v
		}
	}
	return min
}

func (c *cm4) reset() {
	for _, n := range c.s {
		n.reset()
	}
}

// nybble vector
type nvec []byte

func newNvec(w int) nvec {
	return make(nvec, w/2)
}

func (n nvec) get(i uint32) byte {
	// Ugly, but as a single expression so the compiler will inline it :/
	return byte(n[i/2]>>((i&1)*4)) & 0x0f
}

func (n nvec) inc(i uint32) {
	idx := i / 2
	shift := (i & 1) * 4
	v := (n[idx] >> shift) & 0x0f
	if v < 15 {
		n[idx] += 1 << shift
	}
}

func (n nvec) reset() {
	for i := range n {
		n[i] = (n[i] >> 1) & 0x77
	}
}
