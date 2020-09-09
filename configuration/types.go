package configuration

import (
	"fmt"
	"time"
)

// Logging options in the configuration file.
type Logging struct {
	// Standard log level
	// Example:
	// 			panic,
	//			fatal,
	//			error,
	//			warn, warning,
	//			info,
	//			debug
	Level string
	File  *string
}

// DiscloseIP  forpublic IP if
// an error occurs.
type DiscloseIP bool

// EnableProfiling tells if profiling is activated.
type EnableProfiling bool

// EnableTracing tells if tracing is activated.
type EnableTracing bool

// RateLimit holds informations on rate limiting.
type RateLimit struct {
	Burst int
	Rate  float64
}

// Server specifies http based configuration for the underlying server.
type Server struct {
	Port         int
	Certificate  string
	PrivateKey   string
	Host         string
	Limit        *RateLimit
	ReadTimeout  *time.Duration
	WriteTimeout *time.Duration

	// Initial virtual allocation
	// to reduce GC pressure at runtime.
	WithBallast bool
}

// Addr for http interface.
func (s *Server) Addr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// HTTP for configuration on underlying server.
type HTTP struct {
	AllowedOrigins []string
	Request        Request
}

// Request holds configuration on request.
type Request struct {
	Timeout            time.Duration
	MaxRequestInFlight int
}
