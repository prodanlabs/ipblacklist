package ipblacklist

import (
	"context"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/prodanlabs/ipblacklist/pkg/datasource"
)

const (
	xOriginalForwardedFor = "X-Original-Forwarded-For"
	xForwardedFor         = "X-Forwarded-For"
	xRealIP               = "X-Real-Ip"
)

type dynamicBlacklist struct {
	Enabled            bool   `json:"enabled,omitempty" toml:"enabled,omitempty" yaml:"enabled,omitempty"`
	PeriodSeconds      string `json:"periodseconds,omitempty" toml:"periodseconds,omitempty" yaml:"periodseconds,omitempty"`
	RateLimitThreshold int    `json:"ratelimitthreshold,omitempty" toml:"ratelimitthreshold,omitempty" yaml:"ratelimitthreshold,omitempty"`
}

type Config struct {
	StaticBlacklist  []string         `json:"staticblacklist,omitempty" toml:"staticblacklist,omitempty" yaml:"staticblacklist,omitempty"`
	DynamicBlacklist dynamicBlacklist `json:"dynamicblacklist,omitempty" toml:"dynamicblacklist,omitempty" yaml:"dynamicblacklist,omitempty"`
	RealIPDepth      int              `json:"realipdepth,omitempty" toml:"realipdepth,omitempty" yaml:"realipdepth,omitempty"`
	DBPath           string           `json:"dbpath,omitempty" toml:"dbpath,omitempty" yaml:"dbpath,omitempty"`
}

type ipBlackLister struct {
	next   http.Handler
	name   string
	db     *datasource.Sqlite
	config *Config
}

func CreateConfig() *Config {
	return &Config{
		StaticBlacklist:  []string{},
		DynamicBlacklist: dynamicBlacklist{},
	}
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	sqlite, err := datasource.NewSqlite(config.DBPath)
	if err != nil {
		return nil, err
	}

	ipBlacklists := &ipBlackLister{
		next:   next,
		name:   name,
		config: config,
		db:     sqlite,
	}

	return ipBlacklists, nil
}

func (r *ipBlackLister) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	userIP := realIP(req, r.config.RealIPDepth)

	if userIP != nil && inBlackList(userIP, r.config.StaticBlacklist) {
		rw.WriteHeader(http.StatusForbidden)
		return
	}

	if r.config.DynamicBlacklist.Enabled && r.db.InBlacklist(userIP.String()) {
		rw.WriteHeader(http.StatusForbidden)
		return
	}

	if r.config.DynamicBlacklist.Enabled {
		//go r.recordRequestLog(userIP.String(), req.RequestURI)
		go r.recordRequestLog(userIP.String(), req.RequestURI)
	}

	r.next.ServeHTTP(rw, req)
}

func (r *ipBlackLister) recordRequestLog(userIP, url string) {
	t := time.Now().UTC().Format("2006-01-02 15:04:05")
	rl := datasource.RequestLogs{
		CreateTime: t,
		ClientIP:   userIP,
		URL:        url,
		Counter:    0,
		LastTime:   t,
	}

	if err := r.db.InsertOrUpdateLogs(rl, r.config.DynamicBlacklist.PeriodSeconds); err != nil {
		log.Printf("Add %s request %s log failed. %v", userIP, url, err)
		return
	}

	if r.db.RateCount(userIP, url, r.config.DynamicBlacklist.PeriodSeconds) > r.config.DynamicBlacklist.RateLimitThreshold {
		if err := r.db.AddBlacklist(userIP); err != nil {
			log.Printf("Add %s/%s  blacklist failed. %v", userIP, url, err)
			return
		}
	}

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
