package inject

import (
	"github.com/fdrolshagen/jetter/internal"
)

func Inject(requests *[]internal.Request, env internal.Environment) error {

	err := Auth(requests, env)
	if err != nil {
		return err
	}

	return nil
}
