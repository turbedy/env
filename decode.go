// Package env maps environment variables to structs.
package env

import "reflect"

// OptionError represents a failed argument validation.
type OptionError struct {
	Option  string
	Message string
}

func (e *OptionError) Error() string {
	return "env: " + e.Option + ": " + e.Message
}

type Option func(s *state) error

// WithPrefix sets the initial segment of environment keys.
func WithPrefix(prefix string) Option {
	return func(s *state) error {
		if !isSegment(prefix) {
			return &OptionError{"prefix", "invalid segment"}
		}
		s.key.idx = []int{0}
		s.key.buf = []byte(prefix)
		return nil
	}
}

// WithSeparator sets the separator between segments of environment keys.
// The separator must match `^[A-Z0-9_]+$`.
func WithSeparator(sep string) Option {
	return func(s *state) error {
		if sep == "" {
			return &OptionError{"separator", "invalid separator"}
		}
		for i := 0; i < len(sep); i++ {
			if !isUpper(sep[i]) && !isDigit(sep[i]) && sep[i] != '_' {
				return &OptionError{"separator", "invalid separator"}
			}
		}
		s.key.sep = []byte(sep)
		return nil
	}
}

// Decode maps environment variables to the fields of v.
// Unexported fields are ignored, but exported fields of
// embedded structs are considered.
//
// Environment keys consist of segments joined by a separator.
// The default separator is `_`, and each segment must match `^[A-Z_][A-Z0-9_]*$`.
//
// Segments are generated from non-embedded field names as follows:
//   - A name must match `^[A-Z][a-zA-Z0-9]*$`.
//   - `_` is inserted between a digit and the following letter.
//   - `_` is inserted between a lowercase letter and the following uppercase letter.
//   - `_` is inserted before the last uppercase letter in a sequence of uppercase
//     letters followed by a lowercase letter, unless it is followed by v and a digit.
//   - Lowercase letters are converted to uppercase.
//
// Fields are only changed when environment variables are defined.
// If a field implements [encoding.TextUnmarshaler], uses
// [encoding.TextUnmarshaler.UnmarshalText] method as parser.
//
// Otherwise, uses the following type-dependent parsers:
//   - [time.ParseDuration] for [time.Duration].
//   - [strconv.ParseBool] for bools.
//   - [strconv.ParseInt] for integers.
//   - [strconv.ParseUint] for unsigned integers.
//   - [strconv.ParseFloat] for floating points.
//
// These parsers are also used in collections, which can be nested.
// Collection elements are separated by `,` and pairs by `:`.
// Use a backslash to escape separators, including backslashes before separators.
//
// The following struct tags are currently supported:
//
//	// Ignores the field.
//	`env:"-"`
//
//	// Specifies a custom segment for the field.
//	`env:"SEGMENT"`
//
//	// Marks the field as embedded.
//	`env:",inline"`
//
//	// Marks the field as required.
//	`env:",require"`
func Decode(v any, opts ...Option) error {
	s := &state{
		key:  key{sep: []byte{'_'}},
		isep: ",",
		psep: ":",

		cache: cache{
			metadata: make(map[reflect.Type]metadata),
			decoders: make(map[reflect.Type]decoderFunc),
		},
		seen: make(map[reflect.Type]struct{}),
		used: make(map[string]struct{}),
		req:  -1,
	}

	for _, option := range opts {
		if err := option(s); err != nil {
			return err
		}
	}

	rt := reflect.TypeOf(v)
	rv := reflect.ValueOf(v)

	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		if rt == nil {
			return &TypeError{"nil", "expected non-nil pointer to struct"}
		}
		return &TypeError{rt.String(), "expected non-nil pointer to struct"}
	}

	et := rt.Elem()
	for et.Kind() == reflect.Pointer {
		et = et.Elem()
	}
	if et.Kind() != reflect.Struct {
		return &TypeError{rt.String(), "expected non-nil pointer to struct"}
	}

	s.seen[et] = struct{}{}
	s.typs = append(s.typs, et)
	s.out = rv.Elem()
	s.cache.loadDecoder(s.out.Type())(s)
	return s.err
}
