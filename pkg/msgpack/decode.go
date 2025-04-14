package msgpack

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
)

// DecodeMsgPackToJSON decodes MessagePack binary data into JSON bytes.
func DecodeMsgPackToJSON(data []byte) ([]byte, error) {
	buf := bytes.NewBuffer(data)
	val, err := decodeValue(buf)
	if err != nil {
		return nil, err
	}
	return json.Marshal(val)
}

func decodeValue(buf *bytes.Buffer) (interface{}, error) {
	if buf.Len() == 0 {
		return nil, errors.New("empty buffer")
	}
	b, _ := buf.ReadByte()

	switch {
	case b <= 0x7f:
		return int64(b), nil // positive fixint — b in [0x00, 0x7f]
	case b >= 0x80 && b <= 0x8f:
		length := int(b & 0x0f)
		return readMap(buf, length) // fixmap — b in [0x80, 0x8f]
	case b >= 0x90 && b <= 0x9f:
		length := int(b & 0x0f)
		return readArray(buf, length) // fixarray — b in [0x90, 0x9f]
	case b >= 0xa0 && b <= 0xbf:
		length := int(b & 0x1f)
		return readString(buf, length) // fixstr — b in [0xa0, 0xbf]
	case b >= 0xe0:
		return int64(int8(b)), nil // negative fixint — b in [0xe0, 0xff]
	}

	switch b {
	case 0xc0:
		return nil, nil
	case 0xc2:
		return false, nil
	case 0xc3:
		return true, nil

	// Unsigned integers
	case 0xcc: // uint8
		v, _ := buf.ReadByte()
		return uint64(v), nil
	case 0xcd: // uint16
		bs, _ := safeReadN(buf, 2)
		return uint64(bs[0])<<8 | uint64(bs[1]), nil
	case 0xce: // uint32
		bs, _ := safeReadN(buf, 4)
		return uint64(bs[0])<<24 | uint64(bs[1])<<16 | uint64(bs[2])<<8 | uint64(bs[3]), nil
	case 0xcf: // uint64
		bs, _ := safeReadN(buf, 8)
		v := uint64(0)
		for i := 0; i < 8; i++ {
			v = (v << 8) | uint64(bs[i])
		}
		return v, nil

	// Signed integers
	case 0xd0: // int8
		v, _ := buf.ReadByte()
		return int8(v), nil
	case 0xd1: // int16
		bs, _ := safeReadN(buf, 2)
		return int16(int(bs[0])<<8 | int(bs[1])), nil
	case 0xd2: // int32
		bs, _ := safeReadN(buf, 4)
		return int32(int(bs[0])<<24 | int(bs[1])<<16 | int(bs[2])<<8 | int(bs[3])), nil
	case 0xd3: // int64
		bs, _ := safeReadN(buf, 8)
		v := int64(0)
		for i := 0; i < 8; i++ {
			v = (v << 8) | int64(bs[i])
		}
		return v, nil

	// Floats
	case 0xca: // float32
		bs, _ := safeReadN(buf, 4)
		bits := uint32(bs[0])<<24 | uint32(bs[1])<<16 | uint32(bs[2])<<8 | uint32(bs[3])
		return math.Float32frombits(bits), nil
	case 0xcb: // float64
		bs, _ := safeReadN(buf, 8)
		return decodeFloat64(bs), nil

	// Strings
	case 0xd9: // str8
		l, _ := buf.ReadByte()
		return readString(buf, int(l))
	case 0xda: // str16
		bs, _ := safeReadN(buf, 2)
		length := int(bs[0])<<8 | int(bs[1])
		return readString(buf, length)

	// Arrays
	case 0xdc: // array16
		bs, _ := safeReadN(buf, 2)
		length := int(bs[0])<<8 | int(bs[1])
		return readArray(buf, length)
	case 0xdd: // array32
		bs, _ := safeReadN(buf, 4)
		length := int(bs[0])<<24 | int(bs[1])<<16 | int(bs[2])<<8 | int(bs[3])
		return readArray(buf, length)

	// Maps
	case 0xde: // map16
		bs, _ := safeReadN(buf, 2)
		length := int(bs[0])<<8 | int(bs[1])
		return readMap(buf, length)
	case 0xdf: // map32
		bs, _ := safeReadN(buf, 4)
		length := int(bs[0])<<24 | int(bs[1])<<16 | int(bs[2])<<8 | int(bs[3])
		return readMap(buf, length)

	default:
		return nil, fmt.Errorf("unsupported byte: 0x%x", b)
	}
}

func safeReadN(buf *bytes.Buffer, n int) ([]byte, error) {
	if buf.Len() < n {
		return nil, fmt.Errorf("unexpected EOF: need %d bytes, got %d", n, buf.Len())
	}
	return buf.Next(n), nil
}

func readString(buf *bytes.Buffer, length int) (string, error) {
	bs, err := safeReadN(buf, length)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}

func readArray(buf *bytes.Buffer, length int) ([]interface{}, error) {
	arr := make([]interface{}, 0, length)
	for i := 0; i < length; i++ {
		val, err := decodeValue(buf)
		if err != nil {
			return nil, err
		}
		arr = append(arr, val)
	}
	return arr, nil
}

func readMap(buf *bytes.Buffer, length int) (map[string]interface{}, error) {
	m := make(map[string]interface{}, length)
	for i := 0; i < length; i++ {
		keyRaw, err := decodeValue(buf)
		if err != nil {
			return nil, err
		}
		key, ok := keyRaw.(string)
		if !ok {
			return nil, errors.New("non-string map key")
		}
		val, err := decodeValue(buf)
		if err != nil {
			return nil, err
		}
		m[key] = val
	}
	return m, nil
}

func decodeFloat64(bs []byte) float64 {
	v := uint64(0)
	for i := 0; i < 8; i++ {
		v = (v << 8) | uint64(bs[i])
	}
	return math.Float64frombits(v)
}
