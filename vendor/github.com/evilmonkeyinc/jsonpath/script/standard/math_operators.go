package standard

import "math"

type plusOperator struct {
	arg1, arg2 interface{}
}

func (op *plusOperator) Evaluate(parameters map[string]interface{}) (interface{}, error) {
	first, err := getNumber(op.arg1, parameters)
	if err != nil {
		return nil, err
	}

	second, err := getNumber(op.arg2, parameters)
	if err != nil {
		return nil, err
	}

	return first + second, nil
}

type subtractOperator struct {
	arg1, arg2 interface{}
}

func (op *subtractOperator) Evaluate(parameters map[string]interface{}) (interface{}, error) {
	first, err := getNumber(op.arg1, parameters)
	if err != nil {
		return nil, err
	}

	second, err := getNumber(op.arg2, parameters)
	if err != nil {
		return nil, err
	}

	return first - second, nil
}

type multiplyOperator struct {
	arg1, arg2 interface{}
}

func (op *multiplyOperator) Evaluate(parameters map[string]interface{}) (interface{}, error) {
	first, err := getNumber(op.arg1, parameters)
	if err != nil {
		return nil, err
	}

	second, err := getNumber(op.arg2, parameters)
	if err != nil {
		return nil, err
	}

	return first * second, nil
}

type divideOperator struct {
	arg1, arg2 interface{}
}

func (op *divideOperator) Evaluate(parameters map[string]interface{}) (interface{}, error) {
	first, err := getNumber(op.arg1, parameters)
	if err != nil {
		return nil, err
	}

	second, err := getNumber(op.arg2, parameters)
	if err != nil {
		return nil, err
	}

	return first / second, nil
}

type modulusOperator struct {
	arg1, arg2 interface{}
}

func (op *modulusOperator) Evaluate(parameters map[string]interface{}) (interface{}, error) {
	first, err := getInteger(op.arg1, parameters)
	if err != nil {
		return nil, err
	}

	second, err := getInteger(op.arg2, parameters)
	if err != nil {
		return nil, err
	}

	return first % second, nil
}

type powerOfOperator struct {
	arg1, arg2 interface{}
}

func (op *powerOfOperator) Evaluate(parameters map[string]interface{}) (interface{}, error) {
	first, err := getNumber(op.arg1, parameters)
	if err != nil {
		return nil, err
	}

	second, err := getNumber(op.arg2, parameters)
	if err != nil {
		return nil, err
	}

	return math.Pow(first, second), nil
}
