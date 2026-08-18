package env

import (
	"encoding"
	"reflect"
	"strings"
	"time"
)

// TypeError represents a failed schema validation.
type TypeError struct {
	Path    string
	Message string
}

func (e *TypeError) Error() string {
	return "env: " + e.Path + ": " + e.Message
}

type field struct {
	segment  string
	required bool
	leaf     bool
	typ      reflect.Type
	index    int
	decoder  decoderFunc
}

type metadata struct {
	fields []field
	err    error
}

type cache struct {
	metadata map[reflect.Type]metadata
	decoders map[reflect.Type]decoderFunc
}

func (c *cache) loadFields(t reflect.Type) ([]field, error) {
	metadata, ok := c.metadata[t]
	if ok {
		return metadata.fields, metadata.err
	}

build:
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		ft := sf.Type
		for ft.Kind() == reflect.Pointer {
			ft = ft.Elem()
		}

		if !sf.IsExported() && (!sf.Anonymous || ft.Kind() != reflect.Struct) {
			continue
		}
		tag := sf.Tag.Get("env")
		if tag == "-" {
			continue
		}
		path := t.String() + "." + sf.Name

		decoder := c.loadDecoder(sf.Type)
		if decoder == nil {
			metadata.err = &TypeError{path, "unsupported type"}
			break
		}

		segment, opts, _ := strings.Cut(tag, ",")
		explicit := (segment != "")

		if explicit {
			if !isSegment(segment) {
				metadata.err = &TypeError{path, "invalid segment"}
				break
			}
		} else if !sf.Anonymous {
			if segment, ok = parseSegment(sf.Name); !ok {
				metadata.err = &TypeError{path, "invalid segment"}
				break
			}
		}

		required := false
		for opts != "" {
			var option string
			option, opts, _ = strings.Cut(opts, ",")
			switch option {
			case "inline":
				if explicit {
					metadata.err = &TypeError{path, "invalid inline option"}
					break build
				}
				segment = ""
			case "required":
				required = true
			default:
				metadata.err = &TypeError{path, "unknown option"}
				break build
			}
		}

		leaf := (ft.Kind() != reflect.Struct || isUnmarshaler(ft))
		if segment == "" && leaf {
			metadata.err = &TypeError{path, "invalid inline option"}
			break
		}

		metadata.fields = append(metadata.fields, field{
			segment:  segment,
			required: required,
			leaf:     leaf,
			typ:      ft,
			index:    i,
			decoder:  decoder,
		})
	}

	if metadata.err != nil {
		metadata.fields = nil
	}
	c.metadata[t] = metadata
	return metadata.fields, metadata.err
}

var (
	unmarshalerType = reflect.TypeFor[encoding.TextUnmarshaler]()
	durationType    = reflect.TypeFor[time.Duration]()
)

func (c *cache) loadDecoder(t reflect.Type) decoderFunc {
	decoder, ok := c.decoders[t]
	if ok {
		return decoder
	}

	if t.Kind() != reflect.Pointer && reflect.PointerTo(t).Implements(unmarshalerType) {
		decoder = addrUnmarshalerDecoder(t)
	} else if t.Implements(unmarshalerType) {
		decoder = unmarshalerDecoder(t)
	} else if t == durationType {
		decoder = durationDecoder
	} else {
		switch t.Kind() {
		case reflect.String:
			decoder = stringDecoder
		case reflect.Bool:
			decoder = boolDecoder
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			decoder = intDecoder(t.Bits())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			decoder = uintDecoder(t.Bits())
		case reflect.Float32:
			decoder = floatDecoder(32)
		case reflect.Float64:
			decoder = floatDecoder(64)
		case reflect.Struct:
			decoder = structDecoder(t)
		case reflect.Pointer:
			et := t.Elem()
			edecoder := c.loadDecoder(et)
			if edecoder == nil {
				break
			}
			decoder = pointerDecoder(et, edecoder)
		case reflect.Slice, reflect.Array, reflect.Map:
			et := t.Elem()
			if !isValue(et) {
				break
			}
			edecoder := c.loadDecoder(et)
			if edecoder == nil {
				break
			}

			if t.Kind() == reflect.Slice {
				decoder = sliceDecoder(t, et, edecoder)
				break
			}
			if t.Kind() == reflect.Array {
				decoder = arrayDecoder(t, et, edecoder)
				break
			}

			kt := t.Key()
			if !isValue(kt) {
				break
			}
			kdecoder := c.loadDecoder(kt)
			if kdecoder == nil {
				break
			}
			decoder = mapDecoder(t, kt, et, kdecoder, edecoder)
		}
	}

	c.decoders[t] = decoder
	return decoder
}

func isValue(t reflect.Type) bool {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.String:
		return true
	case reflect.Bool:
		return true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	case reflect.Float32, reflect.Float64:
		return true
	case reflect.Slice, reflect.Array, reflect.Map:
		return true
	case reflect.Struct:
		return isUnmarshaler(t)
	default:
		return false
	}
}

func isUnmarshaler(t reflect.Type) bool {
	return t.Implements(unmarshalerType) || reflect.PointerTo(t).Implements(unmarshalerType)
}
