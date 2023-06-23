package ctxslog

import (
	"log/slog"
	"net/http"
	"net/netip"
	"strings"
)

// RemoteAddrIP returns the ip parsed from r.RemoteAddr.
func RemoteAddrIP(r *http.Request) netip.Addr {
	parsed, err := netip.ParseAddrPort(r.RemoteAddr)
	if err != nil {
		slog.DebugCtx(
			r.Context(),
			"ctxslog.RemoteAddrIP: Cannot parse RemoteAddr",
			"err", err,
			"remoteAddr", r.RemoteAddr,
		)
		return netip.Addr{}
	}
	return parsed.Addr()
}

// GCPRealIP gets the real IP form an GCP request (cloud run or app engine).
//
// It picks the last non-local IP from X-Forwarded-For header,
// fallback to RemoteAddrIP if none found.
func GCPRealIP(r *http.Request) netip.Addr {
	xForwardedFor := r.Header.Get("x-forwarded-for")
	split := strings.Split(xForwardedFor, ",")
	for i := len(split) - 1; i >= 0; i-- {
		ip := strings.TrimSpace(split[i])
		addr, err := netip.ParseAddr(ip)
		if err != nil {
			slog.DebugCtx(
				r.Context(),
				"ctxslog.GCPRealIP: Wrong forwarded ip",
				"err", err,
				"x-forwarded-for", xForwardedFor,
				"ip", ip,
			)
			continue
		}
		if addr.IsPrivate() || addr.IsLoopback() {
			continue
		}
		return addr
	}

	return RemoteAddrIP(r)
}

// HTTPRequest returns a group value for some common HTTP request data.
//
// The ip lambda is used to determine the real ip of the request.
// If it's nil, RemoteAddrIP will be used.
func HTTPRequest(r *http.Request, ip func(*http.Request) netip.Addr) slog.Value {
	if ip == nil {
		ip = RemoteAddrIP
	}
	return slog.GroupValue(
		// ref: https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#HttpRequest
		slog.String("requestMethod", r.Method),
		slog.String("requestUrl", r.URL.String()),
		slog.String("userAgent", r.UserAgent()),
		slog.String("remoteIp", ip(r).String()),
		slog.String("referer", r.Referer()),
		slog.String("protocol", r.Proto),
	)
}
