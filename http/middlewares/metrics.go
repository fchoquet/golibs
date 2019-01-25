package middlewares

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fchoquet/golibs/http/ctx"
	"github.com/fchoquet/golibs/http/response"
	"github.com/fchoquet/golibs/metrics"
)

// Metrics wraps the passed handler with standard metrics about the request
func Metrics(client metrics.Client) Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

			start := time.Now()

			httpMethod := strings.ToLower(req.Method)

			m := client.WithTags([]string{
				"http_method:" + httpMethod,
			})

			if user, _, ok := req.BasicAuth(); ok {
				m = m.WithTag("user:" + user)
			}

			// just in case something goes wrong
			// All the tags are not available yet but at least we get something
			m.Incr("requests.count")

			// Process request
			wrappedWriter := response.WrapWriter(w)
			h.ServeHTTP(wrappedWriter, req)

			statusCode := wrappedWriter.StatusCode
			if statusCode == 0 {
				// If status is not explicitly set, then http.server sets it to 200
				statusCode = http.StatusOK
			}
			m = m.WithTag("status:" + strconv.Itoa(statusCode))

			var status string
			switch {
			case statusCode >= 200 && statusCode < 300:
				status = "ok"
			default:
				status = "failed"
			}

			routeName := ""
			if name, ok := ctx.RouteName(req.Context()); ok {
				routeName = name
				m = m.WithTag("route_name:" + routeName)
			}

			m.Incr("requests.attempted")
			m.Incr(fmt.Sprintf("requests.%s", status))

			if routeName != "" {
				m.Incr(fmt.Sprintf("routes.%s.attempted", routeName))
				m.Incr(fmt.Sprintf("routes.%s.%s", routeName, status))
			}

			if status == "ok" {
				m.Timing("duration", start)
			}
		})
	}
}
