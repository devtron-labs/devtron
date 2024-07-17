package govaluate

// sanitizedParameters is a wrapper for Parameters that does sanitization as
// parameters are accessed.
type sanitizedParameters struct {
	orig Parameters
}

func (p sanitizedParameters) Get(key string) (interface{}, error) {
	value, err := p.orig.Get(key)
	if err != nil {
		return nil, err
	}

	return castToFloat64(value), nil
}

func castToFloat64(value interface{}) interface{} {
	switch value := value.(type) {
	case uint8:
		return float64(value)
	case uint16:
		return float64(value)
	case uint32:
		return float64(value)
	case uint64:
		return float64(value)
	case int8:
		return float64(value)
	case int16:
		return float64(value)
	case int32:
		return float64(value)
	case int64:
		return float64(value)
	case int:
		return float64(value)
	case float32:
		return float64(value)
	}

	return value
}
