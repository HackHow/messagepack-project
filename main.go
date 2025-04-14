package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/HackHow/messagepack-project/pkg/msgpack"
)

func main() {
	mode := flag.String("mode", "encode", "Mode: encode or decode")
	input := flag.String("input", "", "JSON string for encode, or HEX string for decode")
	flag.Parse()

	if *input == "" {
		log.Fatal("❌ Please provide input via --input flag.")
	}

	switch *mode {
	case "encode":
		msgPackData, err := msgpack.EncodeJSONToMsgPack([]byte(*input))

		if err != nil {
			log.Fatalf("Encode Error: %v", err)
		}

		fmt.Println("✅ MessagePack (hex):", formatHex(msgPackData))

	case "decode":
		hexClean := strings.ReplaceAll(*input, " ", "")
		data, err := decodeHexString(hexClean)

		if err != nil {
			log.Fatalf("Invalid HEX input: %v", err)
		}

		jsonData, err := msgpack.DecodeMsgPackToJSON(data)

		if err != nil {
			if strings.Contains(err.Error(), "empty buffer") {
				log.Fatal("Decode Error: input appears to be truncated or malformed HEX data.")
			}
			log.Fatalf("Decode Error: %v", err)
		}

		fmt.Println("✅ Decoded JSON:")
		fmt.Println(string(jsonData))

	default:
		log.Fatalf("❌ Unsupported mode: %s. Use 'encode' or 'decode'", *mode)
	}
}

func formatHex(data []byte) string {
	var output strings.Builder

	for i, b := range data {
		if i > 0 {
			output.WriteString(" ")
		}
		output.WriteString(fmt.Sprintf("%02X", b))
	}

	return output.String()
}

func decodeHexString(s string) ([]byte, error) {
	if len(s)%2 != 0 {
		return nil, fmt.Errorf("hex string must be even length")
	}

	res := make([]byte, len(s)/2)
	for i := 0; i < len(res); i++ {
		_, err := fmt.Sscanf(s[2*i:2*i+2], "%02X", &res[i])
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}
