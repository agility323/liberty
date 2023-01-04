package lbtutil

import (
	"os"
)

func ReadVersionFile() string {
	fn := "build_version"
	if data, err := os.ReadFile(fn); err != nil { return string(data) }
	return ""
}
