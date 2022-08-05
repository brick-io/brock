package brock_test

import "testing"

func test_http(t *testing.T) {
	t.Parallel()

	_ = t.Run("test_http", test_http_middleware)
}
