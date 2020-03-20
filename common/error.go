package common

// ReportableError represents an error that will be reported
// to the user as a callback response or as a message,
// depending on the ctx
type ReportableError struct {
	T string
}

func (e ReportableError) Error() string {
	return e.T
}

// Report reports the error to the user
func (e ReportableError) Report(c *Ctx) {
	c.Report(e.T)
}
