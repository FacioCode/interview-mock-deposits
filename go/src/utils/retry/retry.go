package retry

import "github.com/sirupsen/logrus"

type RetryableFunc func() error

func Retry(doFunc RetryableFunc, maxRetries int) error {
	err := doFunc()
	for retriesMade := 0; err != nil && retriesMade < maxRetries; err = doFunc() {
		logrus.WithField("error", err.Error()).Warn("retrying because of error")
		retriesMade += 1
	}
	return err
}
