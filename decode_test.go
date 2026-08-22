package env

import (
	"errors"
	"math"
	"reflect"
	"testing"
	"time"
)

type wrapper struct {
	Exported   string
	unexported string
}

type Nested struct {
	Wrapper wrapper
}

type Inlined struct {
	Wrapper wrapper `env:",inline"`
}

type Embedded struct {
	wrapper
}

func TestDecode_Parsing(t *testing.T) {
	type Scalars struct {
		TextUnmarshaler     *time.Time
		AddrTextUnmarshaler time.Time
		Duration            time.Duration
		String              string
		Bool                bool
		Int                 int
		Int8                int8
		Int16               int16
		Int32               int32
		Int64               int64
		Uint                uint
		Uint8               uint8
		Uint16              uint16
		Uint32              uint32
		Uint64              uint64
		Float32             float32
		Float64             float64
	}

	type Collections struct {
		Slice   []time.Time
		Array   [3]time.Duration
		Map     map[float32]bool
		Escapes map[string][]string
	}

	type Pointers struct {
		Wrapper *wrapper
		Slice   *[]string
		Array   *[0]string
		Map     *map[string]string
	}

	type Segments struct {
		Segment         string `env:"TAG"`
		GoogleAPISecret string
		DNSOverTLS      string
		IPv6Enabled     string
		OpenAuth2       string
	}

	tests := []struct {
		name  string
		key   string
		value string
		got   any
		want  any
	}{
		// Scalars
		{
			"encoding.TextUnmarshaler",
			"TEXT_UNMARSHALER", "2026-08-21T10:00:00Z", &Scalars{},
			&Scalars{TextUnmarshaler: func() *time.Time {
				date := time.Date(2026, 8, 21, 10, 0, 0, 0, time.UTC)
				return &date
			}()},
		},
		{
			"address is encoding.TextUnmarshaler",
			"ADDR_TEXT_UNMARSHALER", "2026-08-21T12:00:00Z", &Scalars{},
			&Scalars{AddrTextUnmarshaler: time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC)},
		},
		{"time.Duration", "DURATION", "16m", &Scalars{}, &Scalars{Duration: 16 * time.Minute}},
		{"string", "STRING", "Hello world!", &Scalars{}, &Scalars{String: "Hello world!"}},
		{"bool", "BOOL", "true", &Scalars{}, &Scalars{Bool: true}},
		{"int", "INT", "-127", &Scalars{}, &Scalars{Int: -127}},
		{"int8", "INT8", "127", &Scalars{}, &Scalars{Int8: 127}},
		{"int16", "INT16", "32767", &Scalars{}, &Scalars{Int16: 32767}},
		{"int32", "INT32", "2147483647", &Scalars{}, &Scalars{Int32: 2147483647}},
		{"int64", "INT64", "9223372036854775807", &Scalars{}, &Scalars{Int64: 9223372036854775807}},
		{"uint", "UINT", "4294967295", &Scalars{}, &Scalars{Uint: 4294967295}},
		{"uint8", "UINT8", "255", &Scalars{}, &Scalars{Uint8: 255}},
		{"uint16", "UINT16", "65535", &Scalars{}, &Scalars{Uint16: 65535}},
		{"uint32", "UINT32", "4294967295", &Scalars{}, &Scalars{Uint32: 4294967295}},
		{"uint64", "UINT64", "18446744073709551615", &Scalars{}, &Scalars{Uint64: 18446744073709551615}},
		{"float32", "FLOAT32", "3.40282346638528859811704183484516925440e+38", &Scalars{}, &Scalars{Float32: math.MaxFloat32}},
		{"float64", "FLOAT64", "1.79769313486231570814527423731704356798070e+308", &Scalars{}, &Scalars{Float64: math.MaxFloat64}},
		{"invalid encoding.TextUnmarshaler", "TEXT_UNMARSHALER", "2026-08-21T10:00:ZZZ", &Scalars{}, nil},
		{"invalid address is encoding.TextUnmarshaler", "ADDR_TEXT_UNMARSHALER", "2026-08-21T12:00:ZZZ", &Scalars{}, nil},
		{"invalid time.Duration", "DURATION", "8sss", &Scalars{}, nil},
		{"invalid bool", "BOOL", "bool", &Scalars{}, nil},
		{"invalid int", "INT", "int", &Scalars{}, nil},
		{"invalid int8", "INT8", "128", &Scalars{}, nil},
		{"invalid int16", "INT16", "32768", &Scalars{}, nil},
		{"invalid int32", "INT32", "2147483648", &Scalars{}, nil},
		{"invalid int64", "INT64", "9223372036854775808", &Scalars{}, nil},
		{"invalid uint", "UINT", "-255", &Scalars{}, nil},
		{"invalid uint8", "UINT8", "256", &Scalars{}, nil},
		{"invalid uint16", "UINT16", "65536", &Scalars{}, nil},
		{"invalid uint32", "UINT32", "4294967296", &Scalars{}, nil},
		{"invalid uint64", "UINT64", "18446744073709551616", &Scalars{}, nil},
		{"invalid float32", "FLOAT32", "3.40282346638528859811704183484516925440e+39", &Scalars{}, nil},
		{"invalid float64", "FLOAT64", "1.79769313486231570814527423731704356798070e+309", &Scalars{}, nil},
		// Collections
		{
			"slice",
			"SLICE", "2026-08-21T10:00:00Z,2026-08-21T12:00:00Z", &Collections{},
			&Collections{Slice: []time.Time{
				time.Date(2026, 8, 21, 10, 0, 0, 0, time.UTC),
				time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC),
			}},
		},
		{
			"array",
			"ARRAY", "8s,16m,32h", &Collections{},
			&Collections{Array: [3]time.Duration{8 * time.Second, 16 * time.Minute, 32 * time.Hour}},
		},
		{
			"map",
			"MAP", "6.28:false,3.14:true", &Collections{},
			&Collections{Map: map[float32]bool{6.28: false, 3.14: true}},
		},
		{
			"escapes",
			"ESCAPES", `:false,\\\:\\:true\,false,\,\\\\:C:\\Documents\\`, &Collections{},
			&Collections{Escapes: map[string][]string{
				``:    {"false"},
				`\:\`: {"true", "false"},
				`,\\`: {`C:\\Documents\\`},
			}},
		},
		{"invalid element of slice", "SLICE", "2026-08-21T12:00:ZZZ", &Collections{}, nil},
		{"invalid element of array", "ARRAY", "8s,16m,32hhh", &Collections{}, nil},
		{"upper bound of array", "ARRAY", "8s,16m,32h,64d", &Collections{}, nil},
		{"lower bound of array", "ARRAY", "8s,16m", &Collections{}, nil},
		{"empty value for array", "ARRAY", "", &Collections{}, nil},
		{"invalid pair of map", "MAP", "true", &Collections{}, nil},
		{"invalid key of map", "MAP", "false:true", &Collections{}, nil},
		{"invalid element of map", "MAP", "3.14:bool", &Collections{}, nil},
		// structs
		{"default field", "UNEXPORTED", "unexported", &wrapper{Exported: "default"}, &wrapper{Exported: "default"}},
		{"unexported field", "UNEXPORTED", "unexported", &wrapper{unexported: "default"}, &wrapper{unexported: "default"}},
		{"nested field", "WRAPPER_EXPORTED", "wrapper", &Nested{}, &Nested{Wrapper: wrapper{Exported: "wrapper"}}},
		{"inlined field", "EXPORTED", "wrapper", &Inlined{}, &Inlined{Wrapper: wrapper{Exported: "wrapper"}}},
		{"embedded field", "EXPORTED", "exported", &Embedded{}, &Embedded{wrapper: wrapper{Exported: "exported"}}},
		// pointers
		{
			"pointer to default field",
			"WRAPPER_EXPORTED", "pointer", &Pointers{Wrapper: &wrapper{unexported: "default"}},
			&Pointers{Wrapper: &wrapper{Exported: "pointer", unexported: "default"}},
		},
		{"pointer to nested field", "WRAPPER_EXPORTED", "pointer", &Pointers{}, &Pointers{Wrapper: &wrapper{Exported: "pointer"}}},
		{"empty value for pointer to slice", "SLICE", "", &Pointers{}, &Pointers{Slice: &[]string{}}},
		{"empty value for pointer to zero array", "ARRAY", "", &Pointers{}, &Pointers{Array: &[0]string{}}},
		{"empty value for pointer to map", "MAP", "", &Pointers{}, &Pointers{Map: &map[string]string{}}},
		{"invalid pointer to zero array", "ARRAY", "8s", &Pointers{}, nil},
		// segments
		{"segment from tag", "TAG", "content", &Segments{}, &Segments{Segment: "content"}},
		{"abbreviation", "GOOGLE_API_SECRET", "content", &Segments{}, &Segments{GoogleAPISecret: "content"}},
		{"abbreviation at boundaries", "DNS_OVER_TLS", "content", &Segments{}, &Segments{DNSOverTLS: "content"}},
		{"abbreviation with version", "IPV6_ENABLED", "content", &Segments{}, &Segments{IPv6Enabled: "content"}},
		{"digit at suffix", "OPEN_AUTH2", "content", &Segments{}, &Segments{OpenAuth2: "content"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(tt.key, tt.value)
			err := Decode(tt.got)
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
			if !reflect.DeepEqual(tt.got, tt.want) {
				t.Fatalf("Decode() = %+v, want %+v", tt.got, tt.want)
			}
		})
	}
}
