package testing

import "io"

type Interface interface {
	Fprintf(dst io.Reader, format string, a ...interface{}) (int, error)
	Printf(string, ...interface{}) (int, error)
	Just(string, int) string
}
