//go:build go1.24 && cache

package govaluate

import (
	"sync"
	"weak"
)

var (
	paramMap = sync.Map{}

	constMap = sync.Map{}
)

func getParameterStage(name string) (*evaluationStage, error) {
	stage, ok := getParamFromMap(name)
	if ok {
		return stage, nil
	}

	operator := makeParameterStage(name)
	ret := &evaluationStage{
		operator: operator,
	}
	storeVal := weak.Make(ret)
	paramMap.Store(name, storeVal)
	return ret, nil
}

func getParamFromMap(name string) (*evaluationStage, bool) {
	stage, ok := paramMap.Load(name)
	if ok {
		ptr, ok := stage.(weak.Pointer[evaluationStage])
		if ok {
			ret := ptr.Value()
			if ret != nil {
				return ret, true
			}
			paramMap.Delete(name)
		}
	}
	return nil, false
}

func getConstantStage(value any) (*evaluationStage, error) {
	stage, ok := getConstantFromMap(value)
	if ok {
		return stage, nil
	}

	operator := makeLiteralStage(value)
	ret := &evaluationStage{
		symbol:   LITERAL,
		operator: operator,
	}
	storeVal := weak.Make(ret)
	constMap.Store(value, storeVal)
	return ret, nil
}

func getConstantFromMap(value any) (*evaluationStage, bool) {
	stage, ok := constMap.Load(value)
	if ok {
		ptr, ok := stage.(weak.Pointer[evaluationStage])
		if ok {
			ret := ptr.Value()
			if ret != nil {
				return ret, true
			}
			constMap.Delete(value)
		}
	}
	return nil, false
}
