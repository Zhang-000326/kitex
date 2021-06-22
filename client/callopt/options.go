// Package callopt contains options that control the behavior of client on request level.
package callopt

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/kitex/internal/configutil"
	"github.com/cloudwego/kitex/internal/utils"
	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/pkg/http"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	remoteinfo2 "github.com/cloudwego/kitex/pkg/rpcinfo/remoteinfo"
)

var (
	callOptionsPool = sync.Pool{
		New: newOptions,
	}
)

type callOptions struct {
	configs      rpcinfo.RPCConfig
	svr          remoteinfo2.RemoteInfo
	locks        *configutil.ConfigLocks
	httpResolver http.Resolver
}

func newOptions() interface{} {
	return &callOptions{
		locks: &configutil.ConfigLocks{
			Tags: make(map[string]struct{}),
		},
	}
}

func (co *callOptions) zero() {
	co.configs = nil
	co.svr = nil
	co.locks.Zero()
}

// Recycle zeros the call option and put it to the pool.
func (co *callOptions) Recycle() {
	co.zero()
	callOptionsPool.Put(co)
}

// Option is a series of options used at the beginning of a RPC call.
type Option struct {
	f func(o *callOptions, di *strings.Builder)
}

// WithHostPort specifies the target address for a RPC call.
// The given address will overwrite the result from Resolver.
func WithHostPort(hostport string) Option {
	return Option{func(o *callOptions, di *strings.Builder) {
		di.WriteString("WithHostPort(")
		di.WriteString(hostport)
		di.WriteString("),")

		if err := setInstance(o.svr, hostport); err != nil {
			panic(fmt.Errorf("WithHostPort: %w", err))
		}
	}}
}

func setInstance(svr remoteinfo2.RemoteInfo, hostport string) error {
	if _, err := net.ResolveTCPAddr("tcp", hostport); err == nil {
		svr.SetInstance(discovery.NewInstance("tcp", hostport, discovery.DefaultWeight, nil))
	} else if _, err := net.ResolveUnixAddr("unix", hostport); err == nil {
		svr.SetInstance(discovery.NewInstance("unix", hostport, discovery.DefaultWeight, nil))
	} else {
		return fmt.Errorf("invalid '%s'", hostport)
	}
	return nil
}

// WithURL specifies the target for a RPC call with url.
// The given url will be resolved to hostport and overwrites the result from Resolver.
func WithURL(url string) Option {
	return Option{func(o *callOptions, di *strings.Builder) {
		di.WriteString("WithURL(")
		di.WriteString(url)
		di.WriteString("),")

		o.svr.SetTag(rpcinfo.HTTPURL, url)
		hostport, err := o.httpResolver.Resolve(url)
		if err != nil {
			panic(fmt.Errorf("http resolve failed: %w", err))
		}

		if err := setInstance(o.svr, hostport); err != nil {
			panic(fmt.Errorf("WithURL: %w", err))
		}
	}}
}

// WithHTTPHost specifies host in http header(work when RPC over http).
func WithHTTPHost(host string) Option {
	return Option{func(o *callOptions, di *strings.Builder) {
		o.svr.SetTag(rpcinfo.HTTPHost, host)
	}}
}

// WithRPCTimeout specifies the RPC timeout for a RPC call.
func WithRPCTimeout(d time.Duration) Option {
	return Option{func(o *callOptions, di *strings.Builder) {
		di.WriteString("WithRPCTimeout(")
		utils.WriteInt64ToStringBuilder(di, d.Milliseconds())
		di.WriteString("ms)")

		cfg := rpcinfo.AsMutableRPCConfig(o.configs)
		cfg.SetRPCTimeout(d)
		o.locks.Bits |= rpcinfo.BitRPCTimeout

		// TODO SetReadWriteTimeout 是否考虑删除
		cfg.SetReadWriteTimeout(d)
		o.locks.Bits |= rpcinfo.BitReadWriteTimeout
	}}
}

// WithConnectTimeout specifies the connection timeout for a RPC call.
func WithConnectTimeout(d time.Duration) Option {
	return Option{func(o *callOptions, di *strings.Builder) {
		di.WriteString("WithConnectTimeout(")
		utils.WriteInt64ToStringBuilder(di, d.Milliseconds())
		di.WriteString("ms)")

		cfg := rpcinfo.AsMutableRPCConfig(o.configs)
		cfg.SetConnectTimeout(d)
		o.locks.Bits |= rpcinfo.BitConnectTimeout
	}}
}

// WithCluster sets the target cluster for service discovery for a RPC call.
func WithCluster(cluster string) Option {
	return lockTargetTag("WithCluster", "cluster", cluster)
}

// WithIDC sets the target IDC for service discovery for a RPC call.
func WithIDC(idc string) Option {
	return lockTargetTag("WithIDC", "idc", idc)
}

func lockTargetTag(opt, key, val string) Option {
	return Option{f: func(o *callOptions, di *strings.Builder) {
		di.WriteString(opt)
		di.WriteByte('(')
		di.WriteString(val)
		di.WriteByte(')')

		o.svr.SetTag(key, val)
		o.locks.Tags[key] = struct{}{}
	}}
}

// Apply applies call options to the rpcinfo.RPCConfig and internal.RemoteInfo of kitex client.
// The return value records the name and arguments of each option.
// This function is for internal purpose only.
func Apply(cos []Option, cfg rpcinfo.RPCConfig, svr remoteinfo2.RemoteInfo, locks *configutil.ConfigLocks, httpResolver http.Resolver) string {
	var buf strings.Builder

	co := callOptionsPool.Get().(*callOptions)
	co.configs = cfg
	co.svr = svr
	co.locks.Merge(locks)
	co.httpResolver = httpResolver

	buf.WriteByte('[')
	for i := range cos {
		if i > 0 {
			buf.WriteByte(',')
		}
		cos[i].f(co, &buf)
	}
	buf.WriteByte(']')

	co.locks.ApplyLocks(cfg, svr)
	co.Recycle()
	return buf.String()
}