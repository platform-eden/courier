package error

type ErrorType int

const (
	CrashError ErrorType = iota
	LogError
	IgnoreError
)

type CourierError struct {
	ErrorType ErrorType
}

func (e ErrorType) String() string {
	types := []string{
		"CrashError",
		"LogError",
		"IgnoreError",
	}

	return types[int(e)]
}
