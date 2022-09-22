package sanitize

import (
	sanitize "github.com/mrz1836/go-sanitize"
)

func SanitizeStr(str string) string {
	return sanitize.Custom(str, `[^\p{Han}a-zA-Z0-9-._]+`)
}
