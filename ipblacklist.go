package ipblacklist

import (
	"context"
	"log"
	"net"
	"net/http"
	"strings"
)

const (
	xOriginalForwardedFor = "X-Original-Forwarded-For"
	xForwardedFor         = "X-Forwarded-For"
	xRealIP               = "X-Real-Ip"
)

type dynamicBlacklist struct {
	Enabled            bool `json:"enabled,omitempty" toml:"enabled,omitempty" yaml:"enabled,omitempty"`
	PeriodSeconds      int  `json:"periodseconds,omitempty" toml:"periodseconds,omitempty" yaml:"periodseconds,omitempty"`
	RateLimitThreshold int  `json:"ratelimitthreshold,omitempty" toml:"ratelimitthreshold,omitempty" yaml:"ratelimitthreshold,omitempty"`
}

type Config struct {
	StaticBlacklist  []string         `json:"staticblacklist,omitempty" toml:"staticblacklist,omitempty" yaml:"staticblacklist,omitempty"`
	DynamicBlacklist dynamicBlacklist `json:"dynamicblacklist,omitempty" toml:"dynamicblacklist,omitempty" yaml:"dynamicblacklist,omitempty"`
}

type ipBlackLister struct {
	next        http.Handler
	name        string
	realIPDepth int
}

func CreateConfig() *Config {
	return &Config{
		StaticBlacklist:  []string{},
		DynamicBlacklist: dynamicBlacklist{},
	}
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	ipBlacklists := &ipBlackLister{
		next:        next,
		name:        name,
		realIPDepth: 0,
	}

	return ipBlacklists, nil
}

func (r *ipBlackLister) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	log.Printf("RealIP : %s", r.RealIP(req, r.realIPDepth))
	r.next.ServeHTTP(rw, req)
}

func (r *ipBlackLister) RealIP(req *http.Request, depth int) net.IP {
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
