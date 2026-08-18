package env

import (
	"encoding"
	"reflect"
	"strconv"
	"time"
)

type state struct {
	in  string
	out reflect.Value
	err error
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
