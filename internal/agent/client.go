package agent

import (
	"net/http"
	"slices"
	"time"

	"github.com/go-resty/resty/v2"
)

var retryCodes = []int{
	http.StatusTooManyRequests,
	http.StatusInternalServerError,
	http.StatusBadGateway,
	http.StatusServiceUnavailable,
	http.StatusGatewayTimeout,
}

var headers = map[string]string{
	"Accept-Encoding":  "",
	"Content-Encoding": "gzip",
	"Content-Type":     "application/json",
}

func NewRestyClient() *resty.Client {
	return resty.New().
		SetHeaders(headers).
		SetRetryCount(3).
		SetRetryAfter(func(c *resty.Client, r *resty.Response) (time.Duration, error) {
			switch r.Request.Attempt {
			case 1:
				return 1 * time.Second, nil
			case 2:
				return 3 * time.Second, nil
			case 3:
				return 5 * time.Second, nil
			default:
				return 0, nil
			}
		}).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			if r != nil {
				statusCode := r.StatusCode()
				if slices.Contains(retryCodes, statusCode) {
					return true
				}
			}
			if err != nil {
				return true
			}
			return false
		})
}
