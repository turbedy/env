package env

import (
	"encoding"
	"errors"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var ErrCollection = errors.New("invalid collection")

// ValueError represents a failed value conversion.
type ValueError struct {
	Key   string
	Value string
	Type  reflect.Type
	Err   error
}

func (e *ValueError) Error() string {
	return "env: " + e.Key + ": parsing " + strconv.Quote(e.Value) + " as " + e.Type.String() + ": " + e.Err.Error()
}

func (e *ValueError) Unwrap() error {
	return e.Err
}

type state struct {
	key  key
	isep string
	psep string

	cache cache
	typs  []reflect.Type
	seen  map[reflect.Type]struct{}
	used  map[string]struct{}
	req   int

	ok  bool
	in  string
	out reflect.Value
	err error
}

func (s *state) push(f field) {
	if _, ok := s.seen[f.typ]; ok {
		s.err = &TypeError{s.typs[0].String(), "cycle via " + f.typ.String()}
		return
	}
	s.seen[f.typ] = struct{}{}
	s.typs = append(s.typs, f.typ)

	if f.require && s.req < 0 {
		s.req = len(s.typs)
	}
	s.key.push(f.segment)
}

func (s *state) pop() {
	delete(s.seen, s.typs[len(s.typs)-1])
	s.typs = s.typs[:len(s.typs)-1]

	if len(s.typs) == s.req {
		s.req = -1
	}
	s.key.pop()
}

func (s *state) lookup() {
	key := s.key.string()
	if _, ok := s.used[key]; ok {
		s.err = &KeyError{key, ErrDuplicate}
		return
	}
	s.used[key] = struct{}{}

	s.in, s.ok = os.LookupEnv(key)
	if s.req >= 0 && !s.ok {
		s.err = &KeyError{key, ErrNotFound}
	}
}

type decoderFunc func(s *state) bool

func addrUnmarshalerDecoder(t reflect.Type) decoderFunc {
	return func(s *state) bool {
		v := reflect.New(t)
		s.err = v.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(s.in))
		if s.err != nil {
			return false
		}
		s.out.Set(v.Elem())
		return true
	}
}

func unmarshalerDecoder(t reflect.Type) decoderFunc {
	return func(s *state) bool {
		v := reflect.New(t).Elem()
		if v.Kind() == reflect.Pointer {
			v.Set(reflect.New(v.Type().Elem()))
		}
		s.err = v.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(s.in))
		if s.err != nil {
			return false
		}
		s.out.Set(v)
		return true
	}
}

func durationDecoder(s *state) bool {
	v, err := time.ParseDuration(s.in)
	if err != nil {
		s.err = err
		return false
	}
	s.out.SetInt(int64(v))
	return true
}

func stringDecoder(s *state) bool {
	s.out.SetString(s.in)
	return true
}

func boolDecoder(s *state) bool {
	v, err := strconv.ParseBool(s.in)
	if err != nil {
		s.err = err
		return false
	}
	s.out.SetBool(v)
	return true
}

func intDecoder(bits int) decoderFunc {
	return func(s *state) bool {
		v, err := strconv.ParseInt(s.in, 10, bits)
		if err != nil {
			s.err = err
			return false
		}
		s.out.SetInt(v)
		return true
	}
}

func uintDecoder(bits int) decoderFunc {
	return func(s *state) bool {
		v, err := strconv.ParseUint(s.in, 10, bits)
		if err != nil {
			s.err = err
			return false
		}
		s.out.SetUint(v)
		return true
	}
}

func floatDecoder(bits int) decoderFunc {
	return func(s *state) bool {
		v, err := strconv.ParseFloat(s.in, bits)
		if err != nil {
			s.err = err
			return false
		}
		s.out.SetFloat(v)
		return true
	}
}

func structDecoder(t reflect.Type) decoderFunc {
	return func(s *state) bool {
		defer func() {
			if s.err != nil {
				var keyErr *KeyError
				var typeErr *TypeError
				var valueErr *ValueError
				if !errors.As(s.err, &keyErr) && !errors.As(s.err, &typeErr) && !errors.As(s.err, &valueErr) {
					s.err = &ValueError{s.key.string(), s.in, s.out.Type(), s.err}
				}
			}
		}()

		fields, err := s.cache.loadFields(t)
		if err != nil {
			s.err = err
			return false
		}

		root := s.out
		found := false

		for _, f := range fields {
			s.push(f)
			if s.err != nil {
				return false
			}

			if f.leaf {
				s.lookup()
				if s.err != nil {
					return false
				}
				if !s.ok {
					s.pop()
					continue
				}
				found = true
			}

			s.out = root.Field(f.index)
			f.decoder(s)
			if s.err != nil {
				return false
			}
			s.pop()
		}
		return found
	}
}

func pointerDecoder(et reflect.Type, edecoder decoderFunc) decoderFunc {
	return func(s *state) bool {
		out := s.out
		defer func() {
			s.out = out
		}()

		var v reflect.Value

		if out.IsNil() {
			v = reflect.New(et)
			s.out = v.Elem()
		} else {
			s.out = out.Elem()
		}

		found := edecoder(s)
		if s.err != nil {
			return false
		}

		if out.IsNil() && found {
			out.Set(v)
		}
		return found
	}
}

func sliceDecoder(t, et reflect.Type, edecoder decoderFunc) decoderFunc {
	return func(s *state) bool {
		if s.in == "" {
			s.out.Set(reflect.MakeSlice(t, 0, 0))
			return true
		}

		in := s.in
		out := s.out
		defer func() {
			s.in = in
			s.out = out
		}()

		v := reflect.MakeSlice(t, 0, 2)
		elems := s.in

		for {
			var ok bool
			s.in, elems, ok = cut(elems, s.isep)
			s.out = reflect.New(et).Elem()

			if edecoder(s); s.err != nil {
				return false
			}
			v = reflect.Append(v, s.out)
			if !ok {
				break
			}
		}
		out.Set(v)
		return true
	}
}

func arrayDecoder(t, et reflect.Type, edecoder decoderFunc) decoderFunc {
	return func(s *state) bool {
		if s.in == "" {
			if t.Len() == 0 {
				s.out.Set(reflect.New(t).Elem())
				return true
			}
			s.err = ErrCollection
			return false
		}

		in := s.in
		out := s.out
		defer func() {
			s.in = in
			s.out = out
		}()

		v := reflect.New(t).Elem()
		elems := s.in

		for i := 0; i < v.Len(); i++ {
			var ok bool
			s.in, elems, ok = cut(elems, s.isep)
			s.out = reflect.New(et).Elem()

			if edecoder(s); s.err != nil {
				return false
			}
			v.Index(i).Set(s.out)
			if !ok && i != t.Len()-1 {
				s.err = ErrCollection
				return false
			}
		}
		if elems != "" {
			s.err = ErrCollection
			return false
		}
		out.Set(v)
		return true
	}
}

func mapDecoder(t, kt, et reflect.Type, kdecoder, edecoder decoderFunc) decoderFunc {
	return func(s *state) bool {
		if s.in == "" {
			s.out.Set(reflect.MakeMap(t))
			return true
		}

		in := s.in
		out := s.out
		defer func() {
			s.in = in
			s.out = out
		}()

		v := reflect.MakeMap(t)
		items := s.in

		for {
			var ok bool
			s.in, items, ok = cut(items, s.isep)
			key, elem, pair := cut(s.in, s.psep)
			if !pair {
				s.err = ErrCollection
				return false
			}

			k := reflect.New(kt).Elem()
			s.in = key
			s.out = k
			if kdecoder(s); s.err != nil {
				return false
			}

			e := reflect.New(et).Elem()
			s.in = elem
			s.out = e
			if edecoder(s); s.err != nil {
				return false
			}

			v.SetMapIndex(k, e)
			if !ok {
				break
			}
		}
		out.Set(v)
		return true
	}
}

func cut(s, sep string) (before, after string, found bool) {
	i := strings.Index(s, sep)
	if i < 0 {
		return s, "", false
	}
	if i == 0 {
		return "", s[i+len(sep):], true
	}

	var buf strings.Builder
	var off int

	for i >= 0 {
		var esc int
		for j := off + i - 1; j >= 0; j-- {
			if s[j] != '\\' {
				break
			}
			esc++
		}

		if esc%2 == 0 {
			buf.WriteString(s[off : off+i-(esc/2)])
			return buf.String(), s[off+i+len(sep):], true
		}

		buf.WriteString(s[off : off+i-(esc/2)-1])
		buf.WriteString(sep)

		off = off + i + len(sep)
		i = strings.Index(s[off:], sep)
	}
	buf.WriteString(s[off:])
	return buf.String(), "", false
}
