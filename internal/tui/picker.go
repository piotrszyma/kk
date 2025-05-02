package tui

import (
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/pkg/errors"
)

type Option interface {
	Label() string
	Preview() string
}

func OptionPicker[T Option](opts []T) (T, error) {

	idx, err := fuzzyfinder.Find(
		opts,
		func(i int) string {
			return opts[i].Label()
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			return opts[i].Preview()
		}))

	if err != nil {
		var zero T
		return zero, errors.Wrap(err, "failed to select option")
	}

	return opts[idx], nil
}

func IsAbortError(err error) bool {
	return errors.Is(err, fuzzyfinder.ErrAbort)
}
