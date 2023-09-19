package utils

// ToInterfaceArray converts an array of string to an array of interface{}
func ToInterfaceArray(arr []string) []interface{} {
	interfaceArr := make([]interface{}, len(arr))
	for i, v := range arr {
		interfaceArr[i] = v
	}
	return interfaceArr
}

// ToStringArray converts an array of interface{} back to an array of string
func ToStringArray(interfaceArr []interface{}) []string {
	stringArr := make([]string, len(interfaceArr))
	for i, v := range interfaceArr {
		stringArr[i] = v.(string)
	}
	return stringArr
}
