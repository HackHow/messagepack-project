package msgpack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
)

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
	case string:
		strLen := len(val)
		if strLen <= 31 {
			buf.WriteByte(0xa0 | byte(strLen))
		} else {
			buf.WriteByte(0xd9)
			buf.WriteByte(byte(strLen))
		}
		buf.WriteString(val)
	case float32:
		buf.WriteByte(0xca)
		// not used in json.Unmarshal, just in case
	case int, int8, int16, int32, int64:
		return encodeInt(buf, int64(val.(int))) // 強轉處理
	case uint, uint8, uint16, uint32, uint64:
		return encodeInt(buf, int64(val.(uint))) // 強轉處理
	case []interface{}:
		if len(val) <= 15 {
			buf.WriteByte(0x90 | byte(len(val)))
		} else {
			buf.Write([]byte{0xdc, byte(len(val) >> 8), byte(len(val))})
		}
		for _, elem := range val {
			if err := encodeValue(buf, elem); err != nil {
				return err
			}
		}
	case map[string]interface{}:
		if len(val) <= 15 {
			buf.WriteByte(0x80 | byte(len(val)))
		} else {
			buf.Write([]byte{0xde, byte(len(val) >> 8), byte(len(val))})
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
		return fmt.Errorf("unsupported type: %T", val)
	}
	return nil
}

func encodeInt(buf *bytes.Buffer, n int64) error {
	if n >= 0 && n <= 127 {
		buf.WriteByte(byte(n))
	} else if n >= -32 && n < 0 {
		buf.WriteByte(0xe0 | byte(n+32))
	} else if n >= -128 && n <= 127 {
		buf.WriteByte(0xd0)
		buf.WriteByte(byte(n))
	} else if n >= -32768 && n <= 32767 {
		buf.WriteByte(0xd1)
		buf.WriteByte(byte(n >> 8))
		buf.WriteByte(byte(n))
	} else if n >= -2147483648 && n <= 2147483647 {
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
	return nil
}
