package inject

import (
	"fmt"
	"github.com/fdrolshagen/jetter/internal"
	"regexp"
)

func Inject(collection *internal.Collection, env internal.Environment) error {
	requests := &collection.Requests

	err := Auth(requests, env)
	if err != nil {
		return err
	}

	return nil
}

func replaceVars(input string, vars map[string]string) string {
	result := input
	for k, v := range vars {
		result = regexp.MustCompile(fmt.Sprintf(`\{\{\s*%s\s*\}\}`, k)).ReplaceAllString(result, v)
	}
	return result
}
