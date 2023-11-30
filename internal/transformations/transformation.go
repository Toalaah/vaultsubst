package transformations

import (
	"encoding/base64"
	"fmt"
	"strings"
)

func Apply(transformation string, value string) (string, error) {
	switch transformation {
	case "base64":
		return base64.StdEncoding.EncodeToString([]byte(value)), nil
	case "base64d":
		b, err := base64.StdEncoding.DecodeString(value)
		return string(b), err
	case "upper":
		return strings.ToUpper(value), nil
	case "lower":
		return strings.ToLower(value), nil
	case "trim":
		return strings.TrimSpace(value), nil
	default:
		return "", fmt.Errorf("Unknown transformation: %s", transformation)
	}
}
