package ipblacklist

import (
	"context"
	"net"
	"net/http"
	"strings"
)

const (
	xOriginalForwardedFor = "X-Original-Forwarded-For"
	xForwardedFor         = "X-Forwarded-For"
	xRealIP               = "X-Real-Ip"
)

type Config struct {
	StaticBlacklist []string `json:"staticblacklist,omitempty" toml:"staticblacklist,omitempty" yaml:"staticblacklist,omitempty"`
	RealIPDepth     int      `json:"realipdepth,omitempty" toml:"realipdepth,omitempty" yaml:"realipdepth,omitempty"`
}

type ipBlackLister struct {
	next   http.Handler
	name   string
	config *Config
}

func CreateConfig() *Config {
	return &Config{
		StaticBlacklist: []string{},
	}
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	ipBlacklists := &ipBlackLister{
		next:   next,
		name:   name,
		config: config,
	}

	return ipBlacklists, nil
}

func (r *ipBlackLister) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	userIP := realIP(req, r.config.RealIPDepth)

	if userIP != nil && inBlackList(userIP, r.config.StaticBlacklist) {
		rw.WriteHeader(http.StatusForbidden)
		return
	}

	r.next.ServeHTTP(rw, req)
}

func inBlackList(ip net.IP, blacklist []string) bool {
	for i := range blacklist {
		if ip.String() == blacklist[i] {
			return true
		}
	}

	return false
}

func realIP(req *http.Request, depth int) net.IP {
	ipFromXRealIP := strings.Split(req.Header.Get(xRealIP), ",")
	if ip := net.ParseIP(ipFromXRealIP[depth]); ip != nil {
		return ip
	}

	ipFromXForwardedFor := strings.Split(req.Header.Get(xForwardedFor), ",")
	if ip := net.ParseIP(ipFromXForwardedFor[depth]); ip != nil {
		return ip
	}

	ipFromXOriginalForwardedFor := strings.Split(req.Header.Get(xOriginalForwardedFor), ",")
	if ip := net.ParseIP(ipFromXOriginalForwardedFor[depth]); ip != nil {
		return ip
	}

	return nil
}
