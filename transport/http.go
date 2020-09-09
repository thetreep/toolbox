package transport

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"time"

	"github.com/cockroachdb/errors"
	log "github.com/sirupsen/logrus"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"golang.org/x/time/rate"

	"github.com/thetreep/toolbox/binary"
	"github.com/thetreep/toolbox/configuration"
	"github.com/thetreep/toolbox/process"
)

// HTTPServer add graceful capacities to a http server.
type HTTPServer struct {
	run     func() error
	onClose func(context.Context) error
}

// Run for Server interface.
func (e *HTTPServer) Run() error {
	return e.run()
}

// Close for Server interface.
func (e *HTTPServer) Close(ctx context.Context) error {
	return e.onClose(ctx)
}

// NewHTTPServer runs a http server with graceful capacities.
func NewHTTPServer(ctx context.Context, conf *configuration.Server, handler http.Handler) (process.Process, error) {
	if conf == nil {
		return nil, errors.New("no configuration provided")
	}

	if conf.WithBallast {
		_ = make([]byte, 10<<30)
	}

	trace.ApplyConfig(trace.Config{DefaultSampler: trace.ProbabilitySampler(75.0)})
	view.SetReportingPeriod(10 * time.Second)

	tracing := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, span := trace.StartSpan(r.Context(), "child")
			defer span.End()

			next.ServeHTTP(w, r)
		})
	}

	handler = tracing(handler)

	if conf.Limit != nil {
		counter := rate.NewLimiter(rate.Limit(conf.Limit.Rate), conf.Limit.Burst)
		limiter := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !counter.Allow() {
					http.Error(
						w,
						http.StatusText(http.StatusTooManyRequests),
						http.StatusTooManyRequests)

					return
				}

				next.ServeHTTP(w, r)
			})
		}

		handler = limiter(handler)
	}

	nr := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			transactionName := routeName(r)

			if transactionName == "" || transactionName == "/info-commit" {
				if r != nil && r.Body != nil {
					defer func() {
						err := r.Body.Close()
						_ = err
					}()
				}

				err := json.NewEncoder(w).Encode(binary.BinaryMetadata())
				if err != nil {
					log.
						WithError(err).
						Error("could not json encode binary metadata")
				}

				return
			}

			next.ServeHTTP(w, r)
		})
	}

	handler = nr(handler)

	server := &http.Server{
		Addr:         conf.Addr(),
		Handler:      handler,
		ReadTimeout:  2 * time.Minute,
		WriteTimeout: 2 * time.Minute,
	}

	log.Infof("running server at http://%s", conf.Addr())

	run := func() error {
		if conf.Certificate != "" && conf.PrivateKey != "" {
			server.TLSConfig = TLSConfiguration()

			err := server.ListenAndServeTLS(
				conf.Certificate,
				conf.PrivateKey,
			)
			if errors.Is(err, http.ErrServerClosed) {
				return errors.Wrap(err, "HTTP server ListenAndServeTLS")
			}

			return nil
		}

		if err := server.ListenAndServe(); errors.Is(err, http.ErrServerClosed) {
			return errors.Wrap(err, "HTTP server ListenAndServe")
		}

		return nil
	}

	onClose := func(ctx context.Context) error {
		err := server.Shutdown(ctx)
		if err != nil {
			return errors.Wrap(err, "shutdown error")
		}

		return nil
	}

	return &HTTPServer{
		run,
		onClose,
	}, nil
}

func routeName(r *http.Request) string {
	if r == nil {
		return "/"
	}

	return r.URL.Path
}

// TLSConfiguration holds sane defaults for https server. like accepted ciphers.
func TLSConfiguration() *tls.Config {
	return &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		},
	}
}
