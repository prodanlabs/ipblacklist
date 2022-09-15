package ipblacklist

import (
	"context"
	"log"
	"net/http"
	"strings"
)

const (
	XOriginalForwardedFor = "X-Original-Forwarded-For"
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

	log.Println("------------>", r.realIPDepth)
	originalForwarded := strings.Split(req.Header.Get(XOriginalForwardedFor), ",")
	log.Printf("X-Original-Forwarded-For cap: %d v0: %s value: %s", len(originalForwarded), originalForwarded[0], originalForwarded)

	forwarded := strings.Split(req.Header.Get(xForwardedFor), ",")
	log.Printf("X-Forwarded-For cap: %d v0: %s value: %s", len(forwarded), forwarded[0], forwarded)

	realIP := strings.Split(req.Header.Get(xRealIP), ",")
	log.Printf("X-Real-Ip cap: %d v0: %s value: %s", len(realIP), realIP[0], realIP)
	r.next.ServeHTTP(rw, req)
}

func (r *ipBlackLister) RealIP(req *http.Request, depth int) {
	originalForwarded := strings.Split(req.Header.Get(XOriginalForwardedFor), ",")
	log.Printf("X-Original-Forwarded-For cap: %d v0: %s value: %s", len(originalForwarded), originalForwarded[0], originalForwarded)

	forwarded := strings.Split(req.Header.Get(xForwardedFor), ",")
	log.Printf("X-Forwarded-For cap: %d v0: %s value: %s", len(forwarded), forwarded[0], forwarded)

	realIP := strings.Split(req.Header.Get(xRealIP), ",")
	log.Printf("X-Real-Ip cap: %d v0: %s value: %s", len(realIP), realIP[0], realIP)

}
