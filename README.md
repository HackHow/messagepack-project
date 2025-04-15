# 📦 MessagePack Project

A Go CLI tool that converts between **JSON** and **MessagePack** formats, implemented from scratch without third-party
MessagePack libraries like [vmihailenco/msgpack](https://github.com/vmihailenco/msgpack).

This project demonstrates how to serialize and deserialize structured data using
the [MessagePack specification](https://msgpack.org/index.html), with full unit test coverage and clean, readable
encoding logic.

---

## 🧾 Requirements

- **Go 1.24.1**
- Recommended tools:
    - [GoLand](https://www.jetbrains.com/go/) – full-featured Go IDE (built-in analysis engine, no need for `gopls`)
    - [VS Code](https://code.visualstudio.com/) + [Go extension](https://marketplace.visualstudio.com/items?itemName=golang.Go)
    - `gopls`: [Go Language Server](https://github.com/golang/tools/blob/master/gopls/README.md) used for IDE code
      intelligence

---

## 📥 Getting Started

```bash
# Clone the repository
git clone https://github.com/HackHow/messagepack-project.git
cd messagepack-project

# Run the CLI tool (see below for modes)
go run main.go --mode encode --input '{"your":"json"}'

# Run all tests
go test -v ./...
```

---

## 🚀 CLI Usage

### 🔄 Mode: Encode (`--mode encode`)

Converts a JSON string into a space-separated MessagePack hex string.

```bash
go run main.go --mode encode --input '{"deviceId":"C1234567","model":"AXIS-Q3515-LV","fps":30,"resolution":"1920x1080","enabled":true}'
```

✅ Output:

```
✅ MessagePack (hex): 85 A8 64 65 76 69 63 65 49 64 A8 43 31 32 33 34 35 36 37 A5 6D 6F 64 65 6C AD 41 58 49 53 2D 51 33 35 31 35 2D 4C 56 A3 66 70 73 1E AA 72 65 73 6F 6C 75 74 69 6F 6E A9 31 39 32 30 78 31 30 38 30 A7 65 6E 61 62 6C 65 64 C3
```

---

### 🔄 Mode: Decode (`--mode decode`)

Converts a MessagePack hex string back to readable JSON.

```bash
go run main.go --mode decode --input "85 A8 64 65 76 69 63 65 49 64 A8 43 31 32 33 34 35 36 37 ..."
```

✅ Output:

```json
{
  "deviceId": "C1234567",
  "model": "AXIS-Q3515-LV",
  "fps": 30,
  "resolution": "1920x1080",
  "enabled": true
}
```

💡 Input may contain or omit whitespace — the CLI will normalize it automatically:

- `85A864...` ✅ valid
- `85 A8 64...` ✅ also valid

---

## 🧪 Running Unit Tests

```bash
go test -v ./...
```

This runs:

- Round-trip tests across all `testdata/*.json` fixtures
- Encoder/decoder error cases:
    - malformed JSON
    - empty input
    - truncated binary
- HEX decoding from [msgpack-lite tool](https://kawanet.github.io/msgpack-lite/)

---

## 📂 Project Structure

```
.
├── main.go                      # CLI tool
├── pkg
│   └── msgpack
│       ├── encode.go            # JSON → MessagePack encoder
│       ├── decode.go            # MessagePack → JSON decoder
│       └── msgpack_codec_test.go # Unit tests
├── testdata/
│   ├── cam01_basic.json
│   ├── cam02_variation.json
│   └── cam03_advanced.json
├── go.mod / go.sum              # Go module files
```

---

## 📚 Why MessagePack?

[MessagePack](https://msgpack.org/index.html) is a binary serialization format that is:

- Compact (smaller than JSON)
- Fast to parse and generate
- Cross-language compatible

It's ideal for performance-critical systems, messaging protocols, or embedded applications.