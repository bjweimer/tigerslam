package intmath

import (

)

func Abs(a int) int {
	if a >= 0 {
		return a
	}
	return -a
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Sign(a int) int {
	if a > 0 {
		return 1
	} else if a < 0 {
		return -1
	}
	
	return 0
}