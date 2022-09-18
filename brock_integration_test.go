//go:build integration

package brock_test

import "testing"

func Test_Integration_Brock(t *testing.T) {
	t.Parallel()

	_ = t.Run("amqp", test_amqp)
	_ = t.Run("dig", test_dig)
}
