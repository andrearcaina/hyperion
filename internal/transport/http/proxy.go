package http

import (
	"errors"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/andrearcaina/hyperion/internal/store"
)

const forwardedRequestHeader = "X-Hyperion-Forwarded"

func (h *Handler) forwardToLeader(w http.ResponseWriter, r *http.Request, err error) bool {
	leader, ok := forwardableLeader(r, err)
	if !ok {
		return false
	}

	r.Header.Set(forwardedRequestHeader, "1")
	h.newLeaderProxy(leader).ServeHTTP(w, r)
	return true
}

func forwardableLeader(r *http.Request, err error) (*store.NotLeaderError, bool) {
	var leader *store.NotLeaderError

	ok := errors.As(err, &leader) && leader.LeaderHTTPAddress != "" && r.Header.Get(forwardedRequestHeader) == ""

	return leader, ok
}

func (h *Handler) newLeaderProxy(leader *store.NotLeaderError) *httputil.ReverseProxy {
	target := &url.URL{
		Scheme: "http",
		Host:   leader.LeaderHTTPAddress,
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = h.proxyTransport
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		h.logger.Error(r.Context(), "failed to forward request to leader", "leader", leader.LeaderID, "error", err)
		writeError(w, http.StatusBadGateway, errors.New("failed to reach Raft leader"))
	}

	return proxy
}
