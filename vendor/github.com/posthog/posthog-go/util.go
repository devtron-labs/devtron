package posthog

type SizeLimitedMap struct {
	ids  map[string][]string
	size int
}

func newSizeLimitedMap(size int) *SizeLimitedMap {
	newMap := SizeLimitedMap{
		ids:  map[string][]string{},
		size: size,
	}

	return &newMap
}

func (sizeLimitedMap *SizeLimitedMap) add(key string, element string) {

	if len(sizeLimitedMap.ids) >= sizeLimitedMap.size {
		sizeLimitedMap.ids = map[string][]string{}
	}

	if val, ok := sizeLimitedMap.ids[key]; ok {
		sizeLimitedMap.ids[key] = append(val, element)
	} else {
		sizeLimitedMap.ids[key] = []string{element}
	}
}

func (sizeLimitedMap *SizeLimitedMap) contains(key string, element string) bool {
	if val, ok := sizeLimitedMap.ids[key]; ok {
		for _, v := range val {
			if v == element {
				return true
			}
		}
	}

	return false
}

func (sizeLimitedMap *SizeLimitedMap) count() int {
	return len(sizeLimitedMap.ids)
}
