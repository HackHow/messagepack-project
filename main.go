package main

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/HackHow/messagepack-project/pkg/msgpack"
)

func main() {
	// 範例 JSON 字串
	inputJSON := `{
		"name": "Howard",
		"age": 26,
		"isDeveloper": true,
		"skills": ["Go", "JavaScript"],
		"profile": { "github": "howardshen", "level": 5 }
	}`

	// JSON → MessagePack
	msgPackData, err := msgpack.EncodeJSONToMsgPack([]byte(inputJSON))
	if err != nil {
		log.Fatal("Encode Error:", err)
	}
	fmt.Println("✅ MessagePack (hex):", hex.EncodeToString(msgPackData))

	// MessagePack → JSON
	outputJSON, err := msgpack.DecodeMsgPackToJSON(msgPackData)
	if err != nil {
		log.Fatal("Decode Error:", err)
	}
	fmt.Println("✅ Decoded JSON:", string(outputJSON))
}
