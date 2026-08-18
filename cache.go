package env

import (
	"encoding"
	"reflect"
	"strings"
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

var unmarshalerType = reflect.TypeFor[encoding.TextUnmarshaler]()

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

func (c *cache) loadDecoder(t reflect.Type) decoderFunc {
	decoder, ok := c.decoders[t]
	if ok {
		return decoder
	}

	c.decoders[t] = decoder
	return decoder
}

func isUnmarshaler(t reflect.Type) bool {
	return t.Implements(unmarshalerType) || reflect.PointerTo(t).Implements(unmarshalerType)
}
