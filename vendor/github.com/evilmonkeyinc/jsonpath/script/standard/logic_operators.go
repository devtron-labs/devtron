package standard

type andOperator struct {
	arg1, arg2 interface{}
}

func (op *andOperator) Evaluate(parameters map[string]interface{}) (interface{}, error) {
	first, err := getBoolean(op.arg1, parameters)
	if err != nil {
		return nil, err
	}

	second, err := getBoolean(op.arg2, parameters)
	if err != nil {
		return nil, err
	}

	return first && second, nil
}

type orOperator struct {
	arg1, arg2 interface{}
}

func (op *orOperator) Evaluate(parameters map[string]interface{}) (interface{}, error) {
	first, err := getBoolean(op.arg1, parameters)
	if err != nil {
		return nil, err
	}

	second, err := getBoolean(op.arg2, parameters)
	if err != nil {
		return nil, err
	}

	return first || second, nil
}

type lessThanOperator struct {
	arg1, arg2 interface{}
}

func (op *lessThanOperator) Evaluate(parameters map[string]interface{}) (interface{}, error) {
	first, err := getNumber(op.arg1, parameters)
	if err != nil {
		return nil, err
	}

	second, err := getNumber(op.arg2, parameters)
	if err != nil {
		return nil, err
	}

	return first < second, nil
}

type lessThanOrEqualOperator struct {
	arg1, arg2 interface{}
}

func (op *lessThanOrEqualOperator) Evaluate(parameters map[string]interface{}) (interface{}, error) {
	first, err := getNumber(op.arg1, parameters)
	if err != nil {
		return nil, err
	}

	second, err := getNumber(op.arg2, parameters)
	if err != nil {
		return nil, err
	}

	return first <= second, nil
}

type greaterThanOperator struct {
	arg1, arg2 interface{}
}

func (op *greaterThanOperator) Evaluate(parameters map[string]interface{}) (interface{}, error) {
	first, err := getNumber(op.arg1, parameters)
	if err != nil {
		return nil, err
	}

	second, err := getNumber(op.arg2, parameters)
	if err != nil {
		return nil, err
	}

	return first > second, nil
}

type greaterThanOrEqualOperator struct {
	arg1, arg2 interface{}
}

func (op *greaterThanOrEqualOperator) Evaluate(parameters map[string]interface{}) (interface{}, error) {
	first, err := getNumber(op.arg1, parameters)
	if err != nil {
		return nil, err
	}

	second, err := getNumber(op.arg2, parameters)
	if err != nil {
		return nil, err
	}

	return first >= second, nil
}

type equalsOperator struct {
	arg1, arg2 interface{}
}

func (op *equalsOperator) Evaluate(parameters map[string]interface{}) (interface{}, error) {

	if firstNumber, err := getNumber(op.arg1, parameters); err == nil {
		if secondNumber, err := getNumber(op.arg2, parameters); err == nil {
			return firstNumber == secondNumber, nil
		}
	}

	first, err := getString(op.arg1, parameters)
	if err != nil {
		return nil, err
	}

	second, err := getString(op.arg2, parameters)
	if err != nil {
		return nil, err
	}

	return first == second, nil
}

type notEqualsOperator struct {
	arg1, arg2 interface{}
}

func (op *notEqualsOperator) Evaluate(parameters map[string]interface{}) (interface{}, error) {

	if firstNumber, err := getNumber(op.arg1, parameters); err == nil {
		if secondNumber, err := getNumber(op.arg2, parameters); err == nil {
			return firstNumber != secondNumber, nil
		}
	}

	first, err := getString(op.arg1, parameters)
	if err != nil {
		return nil, err
	}

	second, err := getString(op.arg2, parameters)
	if err != nil {
		return nil, err
	}

	return first != second, nil
}

type notOperator struct {
	arg interface{}
}

func (op *notOperator) Evaluate(parameters map[string]interface{}) (interface{}, error) {

	result, err := getBoolean(op.arg, parameters)
	if err != nil {
		return nil, err
	}

	return !result, nil
}
