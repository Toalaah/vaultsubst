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
