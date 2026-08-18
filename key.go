package env

import "errors"

var (
	ErrDuplicate = errors.New("duplicate key")
	ErrNotFound  = errors.New("variable not found")
)

// KeyError represents a failed variable lookup.
type KeyError struct {
	Key string
	Err error
}

func (e *KeyError) Error() string {
	return "env: " + e.Key + ": " + e.Err.Error()
}

func (e *KeyError) Unwrap() error {
	return e.Err
}

type key struct {
	sep []byte
	buf []byte
	idx []int
}

func (k *key) push(segment string) {
	k.idx = append(k.idx, len(k.buf))
	if segment == "" {
		return
	}
	if len(k.buf) > 0 {
		k.buf = append(k.buf, k.sep...)
	}
	k.buf = append(k.buf, segment...)
}

func (k *key) pop() {
	k.buf = k.buf[:k.idx[len(k.idx)-1]]
	k.idx = k.idx[:len(k.idx)-1]
}

func (k *key) string() string {
	return string(k.buf)
}

func isSegment(s string) bool {
	if s != "" && isDigit(s[0]) {
		return false
	}
	for i := 0; i < len(s); i++ {
		if !isUpper(s[i]) && !isDigit(s[i]) && s[i] != '_' {
			return false
		}
	}
	return true
}

func parseSegment(s string) (string, bool) {
	if s != "" && isDigit(s[0]) {
		return "", false
	}
	buf := make([]byte, 0, len(s))

	for i := 0; i < len(s); i++ {
		if isLower(s[i]) {
			buf = append(buf, s[i]-32)
			continue
		}

		if isUpper(s[i]) {
			if i == 0 || isDigit(s[i-1]) {
				buf = append(buf, s[i])
				continue
			}
			if isLower(s[i-1]) {
				buf = append(buf, '_', s[i])
				continue
			}
			if i == len(s)-1 || !isLower(s[i+1]) {
				buf = append(buf, s[i])
				continue
			}
			if s[i+1] != 'v' || i == len(s)-2 || !isDigit(s[i+2]) {
				buf = append(buf, '_')
			}
			buf = append(buf, s[i])
			continue
		}

		if isDigit(s[i]) {
			buf = append(buf, s[i])
			if i < len(s)-1 && !isDigit(s[i+1]) {
				buf = append(buf, '_')
			}
			continue
		}

		return "", false
	}
	return string(buf), true
}

func isDigit(c byte) bool {
	return '0' <= c && c <= '9'
}

func isLower(c byte) bool {
	return 'a' <= c && c <= 'z'
}

func isUpper(c byte) bool {
	return 'A' <= c && c <= 'Z'
}
