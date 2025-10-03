package inject

import (
	"github.com/fdrolshagen/jetter/internal"
)

func Inject(collection *internal.Collection, env internal.Environment) error {
	requests := &collection.Requests
	collection.MergeEnvironmentVariables(env)
	err := Auth(requests, env)
	if err != nil {
		return err
	}
	return nil
}
