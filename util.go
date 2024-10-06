package main

import (
	"math"
	"strconv"
)

func string2l(s *string, len int, lval *int64) bool {
	llval, err := strconv.ParseInt(*s, 10, 64)
	if err != nil {
		return false
	}
	if llval < math.MinInt64 || llval > math.MaxInt64 {
		return false
	}

	*lval = llval
	return true
}
