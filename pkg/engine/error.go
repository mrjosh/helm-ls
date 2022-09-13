package engine

type LintError struct {
	location int
	message  string
	errors   []error
}

func (err *LintError) Errors() []error {
	return err.errors
}

func (err *LintError) Error() string {
	message := ""
	for _, e := range err.errors {
		message += e.Error() + "\n"
	}
	return message
}
