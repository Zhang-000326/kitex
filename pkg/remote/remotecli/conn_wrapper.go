package remotecli

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/cloudwego/kitex/pkg/kerrors"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/remote"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/stats"
)

var connWrapperPool sync.Pool

func init() {
	connWrapperPool.New = newConnWrapper
}

// ConnWrapper wraps a connection.
type ConnWrapper struct {
	connPool remote.ConnPool
	conn     net.Conn
	logger   klog.FormatLogger
}

// NewConnWrapper returns a new ConnWrapper using the given connPool and logger.
func NewConnWrapper(connPool remote.ConnPool, logger klog.FormatLogger) *ConnWrapper {
	cm := connWrapperPool.Get().(*ConnWrapper)
	cm.connPool = connPool
	cm.logger = logger
	return cm
}

// GetConn returns a connection using the given Dialer and RPCInfo.
func (cm *ConnWrapper) GetConn(ctx context.Context, d remote.Dialer, ri rpcinfo.RPCInfo) (net.Conn, error) {
	configs := ri.Config()
	timeout := configs.ConnectTimeout()
	if cm.connPool != nil {
		longConn, err := cm.getConnWithPool(ctx, cm.connPool, d, timeout, ri)
		if err != nil {
			return nil, err
		}
		cm.conn = longConn
		if raw, ok := longConn.(remote.RawConn); ok {
			rawConn := raw.RawConn()
			return rawConn, nil
		}
		return cm.conn, nil
	} else {
		var err error
		cm.conn, err = cm.getConnWithDialer(ctx, d, timeout, ri)
		return cm.conn, err
	}
}

// ReleaseConn should notice that ri may nil when oneway
// TODO duplicate release may cause problem?
func (cm *ConnWrapper) ReleaseConn(err error, ri rpcinfo.RPCInfo) {
	if cm.conn == nil {
		return
	}
	if cm.connPool != nil {
		if err == nil {
			if _, ok := ri.To().Tag(rpcinfo.ConnResetTag); ok {
				cm.connPool.Discard(cm.conn)
				cm.logger.Debugf("discarding the connection because peer will shutdown later")
			}
			cm.connPool.Put(cm.conn)
		} else {
			cm.connPool.Discard(cm.conn)
		}
	} else {
		cm.conn.Close()
	}

	cm.zero()
	connWrapperPool.Put(cm)
}

func newConnWrapper() interface{} {
	return &ConnWrapper{}
}

func (cm *ConnWrapper) zero() {
	cm.connPool = nil
	cm.conn = nil
}

func (cm *ConnWrapper) getConnWithPool(ctx context.Context, cp remote.ConnPool, d remote.Dialer,
	timeout time.Duration, ri rpcinfo.RPCInfo) (net.Conn, error) {
	addr := ri.To().Address()
	if addr == nil {
		return nil, kerrors.ErrNoDestAddress
	}
	opt := &remote.ConnOption{Dialer: d, ConnectTimeout: timeout}
	ri.Stats().Record(ctx, stats.ClientConnStart, stats.StatusInfo, "")
	longConn, err := cp.Get(ctx, addr.Network(), addr.String(), opt)
	if err != nil {
		ri.Stats().Record(ctx, stats.ClientConnFinish, stats.StatusError, err.Error())
		return nil, kerrors.ErrGetConnection.WithCause(err)
	}
	ri.Stats().Record(ctx, stats.ClientConnFinish, stats.StatusInfo, "")
	return longConn, nil
}

func (cm *ConnWrapper) getConnWithDialer(ctx context.Context, d remote.Dialer,
	timeout time.Duration, ri rpcinfo.RPCInfo) (net.Conn, error) {
	addr := ri.To().Address()
	if addr == nil {
		return nil, kerrors.ErrNoDestAddress
	}

	ri.Stats().Record(ctx, stats.ClientConnStart, stats.StatusInfo, "")
	conn, err := d.DialTimeout(addr.Network(), addr.String(), timeout)

	if err != nil {
		ri.Stats().Record(ctx, stats.ClientConnFinish, stats.StatusError, err.Error())
		return nil, kerrors.ErrGetConnection.WithCause(err)
	}
	ri.Stats().Record(ctx, stats.ClientConnFinish, stats.StatusInfo, "")
	return conn, nil
}