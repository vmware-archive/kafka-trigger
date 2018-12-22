package utils

import (
	"fmt"
	"github.com/cenkalti/backoff"
)

func RetryExp(o backoff.Operation) error {
	return backoff.Retry(o, backoff.NewExponentialBackOff())
}

func RetryExpTries(o backoff.Operation, tries int) error {
	if tries < 0 {
		return fmt.Errorf("Cannot retry with a negative number of tries: %d", tries)
	}
	b := backoff.WithMaxRetries(backoff.NewExponentialBackOff(), uint64(tries))
	return backoff.Retry(o, b)
}
