package sdkhttp

import (
	"net/http"
	"net/url"
)

type ctxKeyNamedArguments struct{}

// NamedArgsFromRequest is a helper function that extract url.Values that have
// been parsed using MuxMatcherPattern, url.Values should not be empty if
// parsing is successful and should be able to extract further following
// url.Values, same keys in the pattern result in new value added in url.Values.
func NamedArgsFromRequest(r *http.Request) url.Values {
	u, _ := Request.Get(r, ctxKeyNamedArguments{}).(url.Values)

	return u
}

type ctxKeyPanicRecovery struct{}

// PanicRecoveryFromRequest is a helper function that extract error value
// when panic occurred, the value is saved to *http.Request after recovery
// process and right before calling mux.PanicHandler.
func PanicRecoveryFromRequest(r *http.Request) any {
	return Request.Get(r, ctxKeyPanicRecovery{})
}
