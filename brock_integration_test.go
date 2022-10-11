//go:build integration

package brock_test

import "testing"

func TestIntegrationBrock(t *testing.T) {
	t.Parallel()

	_ = t.Run("amqp", testAMQP)
	_ = t.Run("dig", testDig)
}
