package common

import (
	"errors"
	"fmt"
	"log"
	"runtime/debug"

	"github.com/getsentry/sentry-go"
)

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

// Recover checks if there was a panic, logs the error
// to stdout and sentry (if enabled)
// it returns the error that caused the pani
// (nil if there was no panic)
func Recover(rec interface{}, hasSentry bool) error {
	if rec != nil {
		// recover from panic 😱
		var err error
		switch rec := rec.(type) {
		case string:
			err = errors.New(rec)
		case error:
			err = rec
		default:
			err = fmt.Errorf("%v - %#v", rec, rec)
		}

		// Log
		log.Printf("ERROR !!!\n%v\n%s", err, string(debug.Stack()))

		// Sentry logging
		if hasSentry {
			log.Printf("Reporting to sentry")
			sentry.CaptureException(err)
		}
		return err
	}
	return nil
}

// Some std errors
var (
	IllegalPayloadReportableError ReportableError = ReportableError{T: "Illegal payload"}
)
