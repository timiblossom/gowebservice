package constants

type ErrorBadParams struct {
	message string
}

func NewErrorBadParams(message string) *ErrorBadParams {
	return &ErrorBadParams{
		message: message,
	}
}
func (e *ErrorBadParams) Error() string {
	return e.message
}
