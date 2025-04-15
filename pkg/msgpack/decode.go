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

// decodeValue decodes a single MessagePack value from the buffer.
func decodeValue(buf *bytes.Buffer) (interface{}, error) {
	if buf.Len() == 0 {
		return nil, errors.New("empty buffer")
	}
	b, err := buf.ReadByte()
	if err != nil {
		return nil, err
	}
	// Handle fix types.
	switch {
	case b <= 0x7f:
		return int64(b), nil
	case b >= 0xe0:
		return int64(int8(b)), nil
	case b >= 0xa0 && b <= 0xbf:
		length := int(b & 0x1f)
		return readString(buf, length)
	case b >= 0x90 && b <= 0x9f:
		length := int(b & 0x0f)
		return readArray(buf, length)
	case b >= 0x80 && b <= 0x8f:
		length := int(b & 0x0f)
		return readMap(buf, length)
	}
	// Handle remaining codes.
	switch b {
	case 0xc0:
		return nil, nil
	case 0xc2:
		return false, nil
	case 0xc3:
		return true, nil
	case 0xcc:
		v, err := buf.ReadByte()
		if err != nil {
			return nil, err
		}
		return uint64(v), nil
	case 0xcd:
		bs, err := safeReadN(buf, 2)
		if err != nil {
			return nil, err
		}
		return uint64(bs[0])<<8 | uint64(bs[1]), nil
	case 0xce:
		bs, err := safeReadN(buf, 4)
		if err != nil {
			return nil, err
		}
		return uint64(bs[0])<<24 | uint64(bs[1])<<16 | uint64(bs[2])<<8 | uint64(bs[3]), nil
	case 0xcf:
		bs, err := safeReadN(buf, 8)
		if err != nil {
			return nil, err
		}
		var v uint64
		for i := 0; i < 8; i++ {
			v = (v << 8) | uint64(bs[i])
		}
		return v, nil
	case 0xd0:
		v, err := buf.ReadByte()
		if err != nil {
			return nil, err
		}
		return int8(v), nil
	case 0xd1:
		bs, err := safeReadN(buf, 2)
		if err != nil {
			return nil, err
		}
		return int16(int(bs[0])<<8 | int(bs[1])), nil
	case 0xd2:
		bs, err := safeReadN(buf, 4)
		if err != nil {
			return nil, err
		}
		return int32(int(bs[0])<<24 | int(bs[1])<<16 | int(bs[2])<<8 | int(bs[3])), nil
	case 0xd3:
		bs, err := safeReadN(buf, 8)
		if err != nil {
			return nil, err
		}
		var v int64
		for i := 0; i < 8; i++ {
			v = (v << 8) | int64(bs[i])
		}
		return v, nil
	case 0xca:
		bs, err := safeReadN(buf, 4)
		if err != nil {
			return nil, err
		}
		bits := uint32(bs[0])<<24 | uint32(bs[1])<<16 | uint32(bs[2])<<8 | uint32(bs[3])
		return math.Float32frombits(bits), nil
	case 0xcb:
		bs, err := safeReadN(buf, 8)
		if err != nil {
			return nil, err
		}
		return decodeFloat64(bs), nil
	case 0xd9:
		l, err := buf.ReadByte()
		if err != nil {
			return nil, err
		}
		return readString(buf, int(l))
	case 0xda:
		bs, err := safeReadN(buf, 2)
		if err != nil {
			return nil, err
		}
		length := int(bs[0])<<8 | int(bs[1])
		return readString(buf, length)
	case 0xdc:
		bs, err := safeReadN(buf, 2)
		if err != nil {
			return nil, err
		}
		length := int(bs[0])<<8 | int(bs[1])
		return readArray(buf, length)
	case 0xdd:
		bs, err := safeReadN(buf, 4)
		if err != nil {
			return nil, err
		}
		length := int(bs[0])<<24 | int(bs[1])<<16 | int(bs[2])<<8 | int(bs[3])
		return readArray(buf, length)
	case 0xde:
		bs, err := safeReadN(buf, 2)
		if err != nil {
			return nil, err
		}
		length := int(bs[0])<<8 | int(bs[1])
		return readMap(buf, length)
	case 0xdf:
		bs, err := safeReadN(buf, 4)
		if err != nil {
			return nil, err
		}
		length := int(bs[0])<<24 | int(bs[1])<<16 | int(bs[2])<<8 | int(bs[3])
		return readMap(buf, length)
	default:
		return nil, fmt.Errorf("unsupported byte: 0x%x", b)
	}
}

// safeReadN reads n bytes from the buffer.
func safeReadN(buf *bytes.Buffer, n int) ([]byte, error) {
	if buf.Len() < n {
		return nil, fmt.Errorf("unexpected EOF: need %d bytes, got %d", n, buf.Len())
	}
	return buf.Next(n), nil
}

// readString reads a string of the given length.
func readString(buf *bytes.Buffer, length int) (string, error) {
	bs, err := safeReadN(buf, length)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}

// readArray decodes an array from the buffer.
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

// readMap decodes a map (with string keys) from the buffer.
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

// decodeFloat64 converts byte slice to float64.
func decodeFloat64(bs []byte) float64 {
	v := uint64(0)
	for i := 0; i < 8; i++ {
		v = (v << 8) | uint64(bs[i])
	}
	return math.Float64frombits(v)
}
