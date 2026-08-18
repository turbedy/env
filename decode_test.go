package env

import (
	"errors"
	"math"
	"testing"
	"time"
)

func TestDecode_Scalars(t *testing.T) {
	type Scalars struct {
		TextUnmarshaler time.Time
		Duration        time.Duration
		String          string
		Bool            bool
		Int             int
		Int8            int8
		Int16           int16
		Int32           int32
		Int64           int64
		Uint            uint
		Uint8           uint8
		Uint16          uint16
		Uint32          uint32
		Uint64          uint64
		Float32         float32
		Float64         float64
	}

	tests := []struct {
		name  string
		key   string
		value string
		want  *Scalars
	}{
		{
			"encoding.TextUnmarshaler",
			"TEXT_UNMARSHALER", "2026-08-18T10:00:00Z",
			&Scalars{TextUnmarshaler: time.Date(2026, 8, 18, 10, 0, 0, 0, time.UTC)},
		},
		{"time.Duration", "DURATION", "32s", &Scalars{Duration: 32 * time.Second}},
		{"string", "STRING", "Hello world!", &Scalars{String: "Hello world!"}},
		{"bool", "BOOL", "true", &Scalars{Bool: true}},
		{"int", "INT", "-2147483648", &Scalars{Int: -2147483648}},
		{"int8", "INT8", "127", &Scalars{Int8: 127}},
		{"int16", "INT16", "32767", &Scalars{Int16: 32767}},
		{"int32", "INT32", "2147483647", &Scalars{Int32: 2147483647}},
		{"int64", "INT64", "9223372036854775807", &Scalars{Int64: 9223372036854775807}},
		{"uint", "UINT", "4294967295", &Scalars{Uint: 4294967295}},
		{"uint8", "UINT8", "255", &Scalars{Uint8: 255}},
		{"uint16", "UINT16", "65535", &Scalars{Uint16: 65535}},
		{"uint32", "UINT32", "4294967295", &Scalars{Uint32: 4294967295}},
		{"uint64", "UINT64", "18446744073709551615", &Scalars{Uint64: 18446744073709551615}},
		{"float32", "FLOAT32", "3.40282346638528859811704183484516925440e+38", &Scalars{Float32: math.MaxFloat32}},
		{"float64", "FLOAT64", "1.79769313486231570814527423731704356798070e+308", &Scalars{Float64: math.MaxFloat64}},
		{"invalid encoding.TextUnmarshaler", "TEXT_UNMARSHALER", "2026-08-18T10:00:ZZZ", nil},
		{"invalid time.Duration", "DURATION", "32sss", nil},
		{"invalid bool", "BOOL", "bool", nil},
		{"invalid int", "INT", "int", nil},
		{"invalid int8", "INT8", "128", nil},
		{"invalid int16", "INT16", "32768", nil},
		{"invalid int32", "INT32", "2147483648", nil},
		{"invalid int64", "INT64", "9223372036854775808", nil},
		{"invalid uint", "UINT", "-255", nil},
		{"invalid uint8", "UINT8", "256", nil},
		{"invalid uint16", "UINT16", "65536", nil},
		{"invalid uint32", "UINT32", "4294967296", nil},
		{"invalid uint64", "UINT64", "18446744073709551616", nil},
		{"invalid float32", "FLOAT32", "3.40282346638528859811704183484516925440e+39", nil},
		{"invalid float64", "FLOAT64", "1.79769313486231570814527423731704356798070e+309", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(tt.key, tt.value)
			var got Scalars
			err := Decode(&got)
			if tt.want == nil {
				var valueErr *ValueError
				if !errors.As(err, &valueErr) {
					t.Fatalf("Decode() error = %T, want *ValueError", err)
				}
				if valueErr.Key != tt.key {
					t.Fatalf("Decode() error = *ValueError{Key:%v}, want *ValueError{Key:%v}", valueErr.Key, tt.key)
				}
				return
			}
			if err != nil {
				t.Fatalf("Decode() error = %T, want nil", err)
			}
			if got != *tt.want {
				t.Fatalf("Decode() = %+v, want %+v", got, *tt.want)
			}
		})
	}
}
