package main

import "strconv"

func GetKwds(kwds []interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for i := 0; i < len(kwds); i += 2 {
		result[kwds[i].(string)] = kwds[i+1]
	}

	return result
}

func FloatToString(input_num float64) string {
	// to convert a float number to a string
	return strconv.FormatFloat(input_num, 'f', 6, 64)
}

func IntToString(input_num int) string {
	// to convert a float number to a string
	return strconv.Itoa(input_num)
}

func ParseUint(s string, base int, bitSize int) uint64 {
	unit_, _ := strconv.ParseUint(s, base, bitSize)
	return unit_
}
