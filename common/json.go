package common

import (
	"bytes"
	"encoding/json"
)

func DecodeJson(data []byte, v any) error {
	return json.NewDecoder(bytes.NewReader(data)).Decode(v)
}

func DecodeJsonStr(data string, v any) error {
	return DecodeJson(StringToByteSlice(data), v)
}

func EncodeJson(v any) ([]byte, error) {
	return json.Marshal(v)
}
