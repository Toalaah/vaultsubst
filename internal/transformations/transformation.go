package transformations

import (
	"encoding/base64"
	"fmt"
	"strings"
)

// Apply applies and returns a given transformation from an input string. An
// empty string is returned if any errors occur and/or the transformation type
// is invalid, along with the corresponding error.
func Apply(transformation string, s string) (string, error) {
	switch transformation {
	case "base64":
		return base64.StdEncoding.EncodeToString([]byte(s)), nil
	case "base64d":
		b, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return "", err
		}
		return string(b), nil
	case "upper":
		return strings.ToUpper(s), nil
	case "lower":
		return strings.ToLower(s), nil
	case "trim":
		return strings.TrimSpace(s), nil
	default:
		return "", fmt.Errorf("Unknown transformation: %s", transformation)
	}
}
