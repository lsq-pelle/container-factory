package main

import (
	"encoding/json"
	"io"
)

func toReader(f func(io.Writer) error) io.Reader {
	r, w := io.Pipe()
	go func() {
		w.CloseWithError(f(w))
	}()

	return r
}

func toWriter(f func(io.Reader) error) io.Writer {
	r, w := io.Pipe()

	go func() {
		r.CloseWithError(f(r))
	}()

	return w
}

func formatJSON(dst io.Writer, src io.Reader) (err error) {
	decoder := json.NewDecoder(src)
	encoder := json.NewEncoder(dst)

	for {
		var data interface{}

		err = decoder.Decode(&data)
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return
		}

		err = encoder.Encode(&data)
		if err != nil {
			return
		}
	}
}
