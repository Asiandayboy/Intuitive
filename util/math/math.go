/*
A simple math package for basic number comparisons
*/
package math

type Float interface {
	float32 | float64
}

type Integer interface {
	uint8 | uint16 | uint32 | uint64 | uint |
		int8 | int16 | int32 | int64 | int
}

type Number interface {
	Float | Integer
}

func Clamp[T Number](x T, min T, max T) T {
	if x > max {
		return max
	}
	if x < min {
		return min
	}
	return x
}

func Min[T Number](a T, b T) T {
	if a < b {
		return a
	}
	return b
}

func Max[T Number](a T, b T) T {
	if a > b {
		return a
	}
	return b
}

func Abs[T Number](x T) T {
	if x < 0 {
		return -x
	}
	return x
}
