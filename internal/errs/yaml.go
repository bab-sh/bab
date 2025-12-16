package errs

import (
	"regexp"
	"strconv"
	"strings"
)

var yamlLineRegex = regexp.MustCompile(`line (\d+):?`)

func ExtractYAMLLocation(err error) int {
	if err == nil {
		return 0
	}
	matches := yamlLineRegex.FindStringSubmatch(err.Error())
	if len(matches) >= 2 {
		if n, e := strconv.Atoi(matches[1]); e == nil {
			return n
		}
	}
	return 0
}

func CleanYAMLError(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	msg = strings.TrimPrefix(msg, "yaml: ")
	msg = strings.TrimPrefix(msg, "unmarshal errors:\n  ")
	msg = yamlLineRegex.ReplaceAllString(msg, "")
	msg = strings.TrimSpace(msg)
	return msg
}
