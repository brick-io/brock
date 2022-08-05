package brock_test

import "testing"

func Test_Brock(t *testing.T) {
	t.Parallel()

	_ = t.Run("test_http", test_http)
	_ = t.Run("test_parser", test_parser)
}
