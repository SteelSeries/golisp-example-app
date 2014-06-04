package main

import (
	"fmt"
	"math/big"
	"strconv"
)

// BigratToInt converts a *big.Rat to an int64 (with truncation); it
// returns an error for integer overflows.
func BigratToInt(bigrat *big.Rat) (int64, error) {
	float_string := bigrat.FloatString(0)
	return strconv.ParseInt(float_string, 10, 64)
}

// BigratToInt converts a *big.Rat to a *big.Int (with truncation)
func BigratToBigint(bigrat *big.Rat) *big.Int {
	int_string := bigrat.FloatString(0)
	bigint := new(big.Int)
	// no error scenario could be imagined in testing, so discard err
	fmt.Sscan(int_string, bigint)
	return bigint
}

// BigratToFloat converts a *big.Rat to a float64 (with loss of
// precision).
func BigratToFloat(bigrat *big.Rat) float64 {
	float_string := bigrat.FloatString(10) // arbitrary largish precision
	// no error scenario could be imagined in testing, so discard err
	float, _ := strconv.ParseFloat(float_string, 64)
	return float
}

// FloatToBigrat converts a float64 to a *big.Rat.
func FloatToBigrat(float float64) *big.Rat {
	float_string := fmt.Sprintf("%g", float)
	bigrat := new(big.Rat)
	// no error scenario could be imagined in testing, so discard err
	fmt.Sscan(float_string, bigrat)
	return bigrat
}
