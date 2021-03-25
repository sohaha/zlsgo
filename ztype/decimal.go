package ztype

import (
	"math"
	"strconv"
	"strings"
)

var tenToAny = [...]string{0: "0", 1: "1", 2: "2", 3: "3", 4: "4", 5: "5", 6: "6", 7: "7", 8: "8", 9: "9", 10: "a", 11: "b", 12: "c", 13: "d", 14: "e", 15: "f", 16: "g", 17: "h", 18: "i", 19: "j", 20: "k", 21: "l", 22: "m", 23: "n", 24: "o", 25: "p", 26: "q", 27: "r", 28: "s", 29: "t", 30: "u", 31: "v", 32: "w", 33: "x", 34: "y", 35: "z", 36: "A", 37: "B", 38: "C", 39: "D", 40: "E", 41: "F", 42: "G", 43: "H", 44: "I", 45: "J", 46: "K", 47: "L", 48: "M", 49: "N", 50: "O", 51: "P", 52: "Q", 53: "R", 54: "S", 55: "T", 56: "U", 57: "V", 58: "W", 59: "X", 60: "Y", 61: "Z", 62: "_", 63: "-", 64: "|", 65: "<"}

// DecimalToAny Convert decimal to arbitrary decimal values
func DecimalToAny(value, base int) (newNumStr string) {
	if base < 2 {
		return
	}
	if base <= 32 {
		return strconv.FormatInt(int64(value), base)
	}
	var (
		remainder       int
		remainderString string
	)
	for value != 0 {
		remainder = value % base
		if remainder < 66 && remainder > 9 {
			remainderString = tenToAny[remainder]
		} else {
			remainderString = strconv.FormatInt(int64(remainder), 10)
		}
		newNumStr = remainderString + newNumStr
		value = value / base
	}
	return
}

// AnyToDecimal Convert arbitrary decimal values to decimal
func AnyToDecimal(value string, base int) (v int) {
	if base < 2 {
		return
	}
	if base <= 32 {
		n, _ := strconv.ParseInt(value, base, 64)
		v = int(n)
		return
	}
	n := 0.0
	nNum := len(strings.Split(value, "")) - 1
	for _, v := range strings.Split(value, "") {
		tmp := float64(findKey(v))
		if tmp != -1 {
			n = n + tmp*math.Pow(float64(base), float64(nNum))
			nNum = nNum - 1
		} else {
			break
		}
	}
	v = int(n)
	return
}

func findKey(in string) int {
	result := -1
	for k, v := range tenToAny {
		if in == v {
			result = k
		}
	}
	return result
}
