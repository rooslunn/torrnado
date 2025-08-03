package torrnado

import "fmt"

func Operror(op string, err error) error {
	return fmt.Errorf("%s: %s", op, err)
}