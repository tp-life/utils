package jsonx

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/bytedance/sonic"
)

// Marshal marshals v into json bytes.
func Marshal(v any) ([]byte, error) {
	return sonic.Marshal(v)
}

// MarshalToString marshals v into a string.
func MarshalToString(v any) (string, error) {
	data, err := Marshal(v)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// Unmarshal unmarshals data bytes into v.
func Unmarshal(data []byte, v any) error {

	decoder := sonic.ConfigDefault.NewDecoder(bytes.NewReader(data))
	if err := unmarshalUseNumber(decoder, v); err != nil {
		return formatError(string(data), err)
	}

	return nil
}

// UnmarshalFromString unmarshals v from str.
func UnmarshalFromString(str string, v any) error {
	decoder := sonic.ConfigDefault.NewDecoder(strings.NewReader(str))
	if err := unmarshalUseNumber(decoder, v); err != nil {
		return formatError(str, err)
	}

	return nil
}

// UnmarshalFromReader unmarshals v from reader.
func UnmarshalFromReader(reader io.Reader, v any) error {
	var buf strings.Builder
	teeReader := io.TeeReader(reader, &buf)
	decoder := sonic.ConfigDefault.NewDecoder(teeReader)
	if err := unmarshalUseNumber(decoder, v); err != nil {
		return formatError(buf.String(), err)
	}

	return nil
}

func unmarshalUseNumber(decoder sonic.Decoder, v any) error {
	decoder.UseNumber()
	return decoder.Decode(v)
}

func formatError(v string, err error) error {
	return fmt.Errorf("string: `%s`, error: `%w`", v, err)
}
