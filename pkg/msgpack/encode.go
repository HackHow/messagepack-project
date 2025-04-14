package msgpack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
)

// EncodeJSONToMsgPack converts JSON bytes to MessagePack format.
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
		buf.WriteByte(0xcb)
		bits := math.Float64bits(val)
		for i := 7; i >= 0; i-- {
			buf.WriteByte(byte(bits >> (i * 8)))
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
	// 分別處理 signed 整數
	case int:
		return encodeSignedInt(buf, int64(val))
	case int8:
		return encodeSignedInt(buf, int64(val))
	case int16:
		return encodeSignedInt(buf, int64(val))
	case int32:
		return encodeSignedInt(buf, int64(val))
	case int64:
		return encodeSignedInt(buf, val)
	// 分別處理 unsigned 整數
	case uint:
		return encodeUnsignedInt(buf, uint64(val))
	case uint8:
		return encodeUnsignedInt(buf, uint64(val))
	case uint16:
		return encodeUnsignedInt(buf, uint64(val))
	case uint32:
		return encodeUnsignedInt(buf, uint64(val))
	case uint64:
		return encodeUnsignedInt(buf, val)
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
		for k, v := range val {
			if err := encodeValue(buf, k); err != nil {
				return err
			}
			if err := encodeValue(buf, v); err != nil {
				return err
			}
		}
	default:
		// 若型別不符合以上情形，嘗試使用 reflection
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

// encodeSignedInt encodes a signed integer according to MessagePack spec.
func encodeSignedInt(buf *bytes.Buffer, n int64) error {
	if n >= 0 {
		if n <= 127 {
			buf.WriteByte(byte(n))
			return nil
		}
		// 對於正數，若超出 positive fixint 範圍，可考慮仍使用 signed int 編碼
		if n <= 255 {
			buf.WriteByte(0xd0) // int8
			buf.WriteByte(byte(n))
		} else if n <= 32767 {
			buf.WriteByte(0xd1) // int16
			buf.Write([]byte{byte(n >> 8), byte(n)})
		} else if n <= 2147483647 {
			buf.WriteByte(0xd2) // int32
			buf.Write([]byte{
				byte(n >> 24),
				byte(n >> 16),
				byte(n >> 8),
				byte(n),
			})
		} else {
			buf.WriteByte(0xd3) // int64
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
		// 負數部分
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

// encodeUnsignedInt encodes an unsigned integer using MessagePack unsigned codes.
func encodeUnsignedInt(buf *bytes.Buffer, n uint64) error {
	if n <= 127 {
		buf.WriteByte(byte(n))
		return nil
	}
	if n <= 0xff {
		buf.WriteByte(0xcc) // uint8
		buf.WriteByte(byte(n))
	} else if n <= 0xffff {
		buf.WriteByte(0xcd) // uint16
		buf.Write([]byte{byte(n >> 8), byte(n)})
	} else if n <= 0xffffffff {
		buf.WriteByte(0xce) // uint32
		buf.Write([]byte{
			byte(n >> 24),
			byte(n >> 16),
			byte(n >> 8),
			byte(n),
		})
	} else {
		buf.WriteByte(0xcf) // uint64
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
