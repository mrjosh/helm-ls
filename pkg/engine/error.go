package engine

type LintError struct {
	message string
	errors  []error
}

func NewLintError(msg string, errs ...error) *LintError {
	return &LintError{
		errors:  errs,
		message: msg,
	}
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
