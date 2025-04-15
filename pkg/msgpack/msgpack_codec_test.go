package msgpack_test

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/HackHow/messagepack-project/pkg/msgpack"
	"github.com/stretchr/testify/assert"
)

// getTestdataPath returns the path to a file in the testdata folder.
func getTestdataPath(filename string) string {
	return filepath.Join("..", "..", "testdata", filename)
}

// loadTestJSON loads JSON content from testdata folder.
func loadTestJSON(t *testing.T, filename string) []byte {
	data, err := os.ReadFile(getTestdataPath(filename))
	assert.NoError(t, err)
	return data
}

// TestEncodeAndDecodeAllJSONs performs round-trip testing for all JSON files in testdata/
func TestEncodeAndDecodeAllJSONs(t *testing.T) {
	entries, err := os.ReadDir(getTestdataPath(""))
	assert.NoError(t, err)

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		t.Run(entry.Name(), func(t *testing.T) {
			jsonBytes := loadTestJSON(t, entry.Name())

			msgpackData, err := msgpack.EncodeJSONToMsgPack(jsonBytes)
			assert.NoError(t, err)

			decodedJSON, err := msgpack.DecodeMsgPackToJSON(msgpackData)
			assert.NoError(t, err)

			var original, decoded interface{}
			assert.NoError(t, json.Unmarshal(jsonBytes, &original))
			assert.NoError(t, json.Unmarshal(decodedJSON, &decoded))

			assert.Equal(t, original, decoded)
		})
	}
}

func TestEncodeInvalidJSON(t *testing.T) {
	invalidJSON := `{
		"deviceId": "C1234567"
		"model": "AXIS-Q3515-LV"
	}` // Missing comma
	_, err := msgpack.EncodeJSONToMsgPack([]byte(invalidJSON))
	assert.Error(t, err)
}

func TestEncodeEmptyInput(t *testing.T) {
	_, err := msgpack.EncodeJSONToMsgPack([]byte(""))
	assert.Error(t, err)
}

func TestDecodeEmptyBuffer(t *testing.T) {
	_, err := msgpack.DecodeMsgPackToJSON([]byte{})
	assert.Error(t, err)
}

func TestDecodeTruncatedData(t *testing.T) {
	entries, err := os.ReadDir(getTestdataPath(""))
	assert.NoError(t, err)

	found := false
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		jsonBytes := loadTestJSON(t, entry.Name())
		msgpackData, err := msgpack.EncodeJSONToMsgPack(jsonBytes)
		if err != nil || len(msgpackData) < 6 {
			continue
		}

		t.Run("truncated_"+entry.Name(), func(t *testing.T) {
			truncated := msgpackData[:len(msgpackData)-5]
			_, err := msgpack.DecodeMsgPackToJSON(truncated)
			assert.Error(t, err)
		})

		found = true
		break
	}

	if !found {
		t.Skip("No suitable testdata found for truncation")
	}
}

func TestDecodeFromHexString(t *testing.T) {
	// MessagePack-encoded representation of `/testdata/cam01_basic.json`
	// Generated using https://kawanet.github.io/msgpack-lite/
	// Purpose: Test decoder compatibility with external MessagePack binaries.
	hexStr := `85 a8 64 65 76 69 63 65 49 64 a8 43 31 32 33 34 35 36 37 a5 6d 6f 64
	65 6c ad 41 58 49 53 2d 51 33 35 31 35 2d 4c 56 a3 66 70 73 1e a9 72
	65 73 6f 6c 75 74 69 6f 6e a9 31 39 32 30 78 31 30 38 30 a7 65 6e 61 
	62 6c 65 64 c3`
	replacer := strings.NewReplacer(" ", "", "\n", "", "\t", "")
	cleanHex := replacer.Replace(hexStr)
	data, err := hex.DecodeString(cleanHex)
	assert.NoError(t, err)

	decodedJSON, err := msgpack.DecodeMsgPackToJSON(data)
	assert.NoError(t, err)

	var decoded interface{}
	assert.NoError(t, json.Unmarshal(decodedJSON, &decoded))
}
