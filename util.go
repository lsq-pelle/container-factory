package main

import "io"

func toReader(f func(io.Writer) error) io.Reader {
	r, w := io.Pipe()
	go func() {
		w.CloseWithError(f(w))
	}()

	return r
}
