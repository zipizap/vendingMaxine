package models

import (
	"fmt"
	"strings"
)

// convert_IDstring_2_IDuint converts a collection ID (ex: "ColID-1234") to a database collection ID (ex: 1234 uint)
func convert_IDstring_2_IDuint(IDstring string) (IDuint uint, err error) {
	// ID is of the form "<prefix>-1234"
	// ex: "colId-1234"
	parts := strings.Split(IDstring, "-")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid ID format %s", IDstring)
	}
	_, err = fmt.Sscanf(parts[1], "%d", &IDuint)
	if err != nil {
		return 0, err
	}
	return IDuint, nil
}

// convert_IDuint_2_IDstring converts a uint ID to a string ID with a prefix
// Ex: convert_IDuint_2_IDstring("ColID", 1234) returns "ColID-1234"
func convert_IDuint_2_IDstring(prefix string, IDuint uint) string {
	return fmt.Sprintf("%s-%d", prefix, IDuint)
}
