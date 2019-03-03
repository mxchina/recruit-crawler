package utils

import (
	"fmt"
	"math"
	"strconv"
)

func StrToFloat64(str string, len int) float64 {
	lenstr := "%." + strconv.Itoa(len) + "f"
	value,_ := strconv.ParseFloat(str,64)
	nstr := fmt.Sprintf(lenstr,value)
	val,_ := strconv.ParseFloat(nstr,64)
	return val
}

func StrToFloat64Round(str string, prec int, round bool) float64 {
	f, _ := strconv.ParseFloat(str, 64)
	return Precision(f, prec, round)
}

func Precision(f float64, prec int, round bool) float64 {
	pow10_n := math.Pow10(prec)
	if round {
		return math.Trunc(f+0.5/pow10_n) * pow10_n / pow10_n
	}
	return math.Trunc((f)*pow10_n) / pow10_n
}
