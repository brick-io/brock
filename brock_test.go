package brock_test

import "testing"

func TestBrock(t *testing.T) {
	t.Parallel()

	_ = t.Run("crypto", testCrypto)
	_ = t.Run("fsm", testFSM)
	_ = t.Run("http", testHTTP)
	_ = t.Run("parser", testParser)
	_ = t.Run("sql", testSQL)
}
