package thrift

import (
	"context"

	"github.com/apache/thrift/lib/go/thrift"

	"github.com/cloudwego/kitex/pkg/generic/descriptor"
)

// NewWriteStruct ...
func NewWriteStruct(svc *descriptor.ServiceDescriptor, method string, isClient bool) (*WriteStruct, error) {
	fnDsc, err := svc.LookupFunctionByMethod(method)
	if err != nil {
		return nil, err
	}
	ty := fnDsc.Request
	if !isClient {
		ty = fnDsc.Response
	}
	ws := &WriteStruct{
		ty:             ty,
		hasRequestBase: fnDsc.HasRequestBase && isClient,
	}
	return ws, nil
}

// WriteStruct implement of MessageWriter
type WriteStruct struct {
	ty             *descriptor.TypeDescriptor
	hasRequestBase bool
}

var _ MessageWriter = (*WriteStruct)(nil)

// Write ...
func (m *WriteStruct) Write(ctx context.Context, out thrift.TProtocol, msg interface{}, requestBase *Base) error {
	if !m.hasRequestBase {
		requestBase = nil
	}
	return wrapStructWriter(ctx, msg, out, m.ty, &writerOption{requestBase: requestBase})
}

// NewReadStruct ...
func NewReadStruct(svc *descriptor.ServiceDescriptor, isClient bool) *ReadStruct {
	return &ReadStruct{
		svc:      svc,
		isClient: isClient,
	}
}

// ReadStruct implement of MessageReaderWithMethod
type ReadStruct struct {
	svc      *descriptor.ServiceDescriptor
	isClient bool
}

var _ MessageReader = (*ReadStruct)(nil)

// Read ...
func (m *ReadStruct) Read(ctx context.Context, method string, in thrift.TProtocol) (interface{}, error) {
	fnDsc, err := m.svc.LookupFunctionByMethod(method)
	if err != nil {
		return nil, err
	}
	fDsc := fnDsc.Response
	if !m.isClient {
		fDsc = fnDsc.Request
	}
	return skipStructReader(ctx, in, fDsc, &readerOption{throwException: true})
}
