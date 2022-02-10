package eventer

import "fmt"

type MissingOperatorClientParamError struct {
	Param string
}

func (err *MissingOperatorClientParamError) Error() string {
	return fmt.Sprintf("%s is required but was not assigned", err.Param)
}
