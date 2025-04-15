package msgpack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"sort"
)

// EncodeJSONToMsgPack converts JSON bytes to MessagePack format.
// For round-trip test, we don't enforce deterministic ordering.
func EncodeJSONToMsgPack(jsonData []byte) ([]byte, error) {
	var v interface{}
	if err := json.Unmarshal(jsonData, &v); err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	if err := encodeValue(buf, v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// encodeValue encodes a generic value into MessagePack format.
func encodeValue(buf *bytes.Buffer, v interface{}) error {
	switch val := v.(type) {
	case nil:
		buf.WriteByte(0xc0)
	case bool:
		if val {
			buf.WriteByte(0xc3)
		} else {
			buf.WriteByte(0xc2)
		}
	case float64:
		if val == float64(int64(val)) {
			return encodeSignedInt(buf, int64(val))
		} else {
			buf.WriteByte(0xcb)
			bits := math.Float64bits(val)
			for i := 7; i >= 0; i-- {
				buf.WriteByte(byte(bits >> (i * 8)))
			}
		}
	case float32:
		buf.WriteByte(0xca)
		bits := math.Float32bits(val)
		for i := 3; i >= 0; i-- {
			buf.WriteByte(byte(bits >> (i * 8)))
		}
	case string:
		strLen := len(val)
		if strLen <= 31 {
			buf.WriteByte(0xa0 | byte(strLen))
		} else if strLen <= 255 {
			buf.WriteByte(0xd9)
			buf.WriteByte(byte(strLen))
		} else {
			buf.WriteByte(0xda)
			buf.Write([]byte{byte(strLen >> 8), byte(strLen)})
		}
		buf.WriteString(val)
	case int, int8, int16, int32, int64:
		return encodeSignedInt(buf, reflect.ValueOf(val).Int())
	case uint, uint8, uint16, uint32, uint64:
		return encodeUnsignedInt(buf, reflect.ValueOf(val).Uint())
	case []interface{}:
		length := len(val)
		if length <= 15 {
			buf.WriteByte(0x90 | byte(length))
		} else if length <= 65535 {
			buf.WriteByte(0xdc)
			buf.Write([]byte{byte(length >> 8), byte(length)})
		} else {
			buf.WriteByte(0xdd)
			buf.Write([]byte{
				byte(length >> 24),
				byte(length >> 16),
				byte(length >> 8),
				byte(length),
			})
		}
		for _, elem := range val {
			if err := encodeValue(buf, elem); err != nil {
				return err
			}
		}
	case map[string]interface{}:
		// Use lexicographical order for consistent encoding even without ordering guarantee
		var keys []string
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		length := len(val)
		if length <= 15 {
			buf.WriteByte(0x80 | byte(length))
		} else if length <= 65535 {
			buf.WriteByte(0xde)
			buf.Write([]byte{byte(length >> 8), byte(length)})
		} else {
			buf.WriteByte(0xdf)
			buf.Write([]byte{
				byte(length >> 24),
				byte(length >> 16),
				byte(length >> 8),
				byte(length),
			})
		}
		for _, k := range keys {
			if err := encodeValue(buf, k); err != nil {
				return err
			}
			if err := encodeValue(buf, val[k]); err != nil {
				return err
			}
		}
	default:
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return encodeSignedInt(buf, rv.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return encodeUnsignedInt(buf, rv.Uint())
		default:
			return fmt.Errorf("unsupported type: %T", v)
		}
	}
	return nil
}

// encodeSignedInt encodes a signed integer.
func encodeSignedInt(buf *bytes.Buffer, n int64) error {
	if n >= 0 {
		if n <= 127 {
			buf.WriteByte(byte(n))
			return nil
		}
		if n <= 255 {
			buf.WriteByte(0xd0)
			buf.WriteByte(byte(n))
		} else if n <= 32767 {
			buf.WriteByte(0xd1)
			buf.Write([]byte{byte(n >> 8), byte(n)})
		} else if n <= 2147483647 {
			buf.WriteByte(0xd2)
			buf.Write([]byte{
				byte(n >> 24),
				byte(n >> 16),
				byte(n >> 8),
				byte(n),
			})
		} else {
			buf.WriteByte(0xd3)
			buf.Write([]byte{
				byte(n >> 56),
				byte(n >> 48),
				byte(n >> 40),
				byte(n >> 32),
				byte(n >> 24),
				byte(n >> 16),
				byte(n >> 8),
				byte(n),
			})
		}
	} else {
		if n >= -32 {
			buf.WriteByte(0xe0 | byte(n+32))
		} else if n >= -128 {
			buf.WriteByte(0xd0)
			buf.WriteByte(byte(n))
		} else if n >= -32768 {
			buf.WriteByte(0xd1)
			buf.Write([]byte{byte(n >> 8), byte(n)})
		} else if n >= -2147483648 {
			buf.WriteByte(0xd2)
			buf.Write([]byte{
				byte(n >> 24),
				byte(n >> 16),
				byte(n >> 8),
				byte(n),
			})
		} else {
			buf.WriteByte(0xd3)
			buf.Write([]byte{
				byte(n >> 56),
				byte(n >> 48),
				byte(n >> 40),
				byte(n >> 32),
				byte(n >> 24),
				byte(n >> 16),
				byte(n >> 8),
				byte(n),
			})
		}
	}
	return nil
}

// encodeUnsignedInt encodes an unsigned integer.
func encodeUnsignedInt(buf *bytes.Buffer, n uint64) error {
	if n <= 127 {
		buf.WriteByte(byte(n))
		return nil
	}
	if n <= 0xff {
		buf.WriteByte(0xcc)
		buf.WriteByte(byte(n))
	} else if n <= 0xffff {
		buf.WriteByte(0xcd)
		buf.Write([]byte{byte(n >> 8), byte(n)})
	} else if n <= 0xffffffff {
		buf.WriteByte(0xce)
		buf.Write([]byte{
			byte(n >> 24),
			byte(n >> 16),
			byte(n >> 8),
			byte(n),
		})
	} else {
		buf.WriteByte(0xcf)
		buf.Write([]byte{
			byte(n >> 56),
			byte(n >> 48),
			byte(n >> 40),
			byte(n >> 32),
			byte(n >> 24),
			byte(n >> 16),
			byte(n >> 8),
			byte(n),
		})
	}
	return nil
}
