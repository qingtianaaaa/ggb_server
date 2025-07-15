package converter

import (
	"github.com/jinzhu/copier"
)

type Converter interface{}

type Option func(*copier.Option)

func WithIgnoreEmpty() Option {
	return func(option *copier.Option) {
		option.IgnoreEmpty = true
	}
}

func WithDeepCopy() Option {
	return func(option *copier.Option) {
		option.DeepCopy = true
	}
}
