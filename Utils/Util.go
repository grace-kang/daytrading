package Utils

import "strconv"

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
