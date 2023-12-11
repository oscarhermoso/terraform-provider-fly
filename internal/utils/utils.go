package utils

func P[T any](x T) *T {
	return &x
}
