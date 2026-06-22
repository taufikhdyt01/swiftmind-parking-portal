package gateway

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"swiftmind/pkg/domain"
	"swiftmind/pkg/httpx"
)

// proxyTo reverse-proxies a request to target, trimming stripPrefix from the
// path and forwarding the authenticated user's identity as trusted headers.
// Downstream services run on the internal network and trust these headers.
func (g *Gateway) proxyTo(target, stripPrefix string) http.Handler {
	base, err := url.Parse(target)
	if err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			httpx.Error(w, http.StatusInternalServerError, "bad downstream url")
		})
	}

	proxy := httputil.NewSingleHostReverseProxy(base)
	defaultDirector := proxy.Director
	proxy.Director = func(r *http.Request) {
		defaultDirector(r)
		r.URL.Path = base.Path + strings.TrimPrefix(r.URL.Path, stripPrefix)
		r.Host = base.Host

		// Strip any client-supplied identity headers, then set trusted ones.
		r.Header.Del(domain.HeaderUserID)
		r.Header.Del(domain.HeaderUserRole)
		r.Header.Del(domain.HeaderUserEmail)
		r.Header.Del(domain.HeaderUserName)
		if claims := claimsFromCtx(r.Context()); claims != nil {
			r.Header.Set(domain.HeaderUserID, claims.UserID)
			r.Header.Set(domain.HeaderUserRole, claims.Role)
			r.Header.Set(domain.HeaderUserEmail, claims.Email)
			r.Header.Set(domain.HeaderUserName, claims.Name)
		}
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, _ *http.Request, _ error) {
		httpx.Error(w, http.StatusBadGateway, "downstream service unavailable")
	}
	return proxy
}
