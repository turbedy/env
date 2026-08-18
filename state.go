package env

type state struct{}

type decoderFunc func(s *state) bool
