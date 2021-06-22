package tpl

// ServiceTpl is the template for generating the servicename.go source.
var ServiceTpl = `// Code generated by Kitex {{.Version}}. DO NOT EDIT.

package {{ToLower .ServiceName}}

import (
	{{- range $path, $alias := .Imports}}
	{{$alias }}"{{$path}}"
	{{- end}}
)

{{- if gt (len .CombineServices) 0}}
type {{call .ServiceTypeName}} interface {
{{- range .CombineServices}}
	{{.PkgRefName}}.{{.ServiceName}}
{{- end}}
}
{{- end}}

func serviceInfo() *kitex.ServiceInfo {
	return {{LowerFirst .ServiceName}}ServiceInfo
 }

var {{LowerFirst .ServiceName}}ServiceInfo = newServiceInfo()

func newServiceInfo() *kitex.ServiceInfo {
	serviceName := "{{.RawServiceName}}"
	handlerType := (*{{call .ServiceTypeName}})(nil)
	methods := map[string]kitex.MethodInfo{
	{{- range .AllMethods}}
		"{{.RawName}}":
			kitex.NewMethodInfo({{LowerFirst .Name}}Handler, new{{.ArgStructName}}, {{if .Oneway}}nil{{else}}new{{.ResStructName}}{{end}}, {{if .Oneway}}true{{else}}false{{end}}),
	{{- end}}
	}
	extra := map[string]interface{}{
		"PackageName":	 "{{.PkgInfo.PkgRefName}}",
	}
	{{- if gt (len .CombineServices) 0}}
	extra["combine_service"] = true
	{{- end}}
	{{- if .HasStreaming}}
	extra["streaming"] = true
	{{- end}}
	svcInfo := &kitex.ServiceInfo{
		ServiceName: 	 serviceName,
		HandlerType: 	 handlerType,
		Methods:     	 methods,
		PayloadCodec:  	 kitex.{{.Codec | UpperFirst}},
		KiteXGenVersion: "{{.Version}}",
		Extra:           extra,
	}
	return svcInfo
}

{{range .AllMethods}}
{{- $isStreaming := or .ClientStreaming .ServerStreaming}}
{{- $unary := and (not .ServerStreaming) (not .ClientStreaming)}}
{{- $clientSide := and .ClientStreaming (not .ServerStreaming)}}
{{- $serverSide := and (not .ClientStreaming) .ServerStreaming}}
{{- $bidiSide := and .ClientStreaming .ServerStreaming}}
{{- $arg := "" }}
{{- if $.HasStreaming }}
{{- $arg = index .Args 0}}
{{- end}}

func {{LowerFirst .Name}}Handler(ctx context.Context, handler interface{}, arg, result interface{}) error {
	{{- if $.HasStreaming}}
	{{- /* streaming logic */}}
	st := arg.(*streaming.Args).Stream
	{{- if $isStreaming}}
	stream := &{{LowerFirst .ServiceName}}{{.RawName}}Server{st}
	{{- end}}
	{{- if or $unary $serverSide}}
	req := new({{NotPtr $arg.Type}})
	if err := st.RecvMsg(req); err != nil {
		return err
	}
	{{- end}}
	{{- if $isStreaming}}
	return handler.({{.PkgRefName}}.{{.ServiceName}}).{{.Name}}({{if $serverSide}}req, {{end}}stream)
	{{- else}}
	resp, err := handler.({{.PkgRefName}}.{{.ServiceName}}).{{.Name}}(ctx, req)	
	if err != nil {
		return err
	}
	if err := st.SendMsg(resp); err != nil {
		return err
	}
	return nil
	{{- end}}
	{{- else}}
	{{- /* unary logic */}}
	{{if .Args}}realArg := arg.(*{{if not .GenArgResultStruct}}{{.PkgRefName}}.{{end}}{{.ArgStructName}}){{end}}
	{{if or (not .Void) .Exceptions}}realResult := result.(*{{if not .GenArgResultStruct}}{{.PkgRefName}}.{{end}}{{.ResStructName}}){{end}}
	{{if .Void}}err := handler.({{.PkgRefName}}.{{.ServiceName}}).{{.Name}}(ctx{{range .Args}}, realArg.{{.Name}}{{end}})
	{{else}}success, err := handler.({{.PkgRefName}}.{{.ServiceName}}).{{.Name}}(ctx{{range .Args}}, realArg.{{.Name}}{{end}})
	{{end -}}
	if err != nil {
	{{if .Exceptions -}}
		switch v := err.(type) {
		{{range .Exceptions -}}
		case {{.Type}}:
			realResult.{{.Name}} = v
		{{end -}}
		default:
			return err
		}
	} else {
	{{else -}}
		return err
	}
	{{end -}}
	{{if not .Void}}realResult.Success = {{if .IsResponseNeedRedirect}}&{{end}}success{{end}}
	{{- if .Exceptions}}}{{end}}
	return nil
	{{- end}} {{/* streaming end */}}
}

{{- /* define streaming struct */}}
{{- if $isStreaming}}
type {{LowerFirst .ServiceName}}{{.RawName}}Client struct {
	streaming.Stream
}

{{- if or $clientSide $bidiSide}}
func (x *{{LowerFirst .ServiceName}}{{.RawName}}Client) Send(m {{$arg.Type}}) error {
	return x.Stream.SendMsg(m)
}
{{- end}}

{{- if $clientSide}}
func (x *{{LowerFirst .ServiceName}}{{.RawName}}Client) CloseAndRecv() ({{.Resp.Type}}, error) {
	if err := x.Stream.Close(); err != nil {
		return nil, err
	}
	m := new({{NotPtr .Resp.Type}})
	return m, x.Stream.RecvMsg(m)
}
{{- end}}

{{- if or $serverSide $bidiSide}}
func (x *{{LowerFirst .ServiceName}}{{.RawName}}Client) Recv() ({{.Resp.Type}}, error) {
	m := new({{NotPtr .Resp.Type}})
	return m, x.Stream.RecvMsg(m)
}
{{- end}}

type {{LowerFirst .ServiceName}}{{.RawName}}Server struct {
	streaming.Stream
}

{{if or $serverSide $bidiSide}}
func (x *{{LowerFirst .ServiceName}}{{.RawName}}Server) Send(m {{.Resp.Type}}) error {
	return x.Stream.SendMsg(m)
}
{{end}}

{{if $clientSide}}
func (x *{{LowerFirst .ServiceName}}{{.RawName}}Server) SendAndClose(m {{.Resp.Type}}) error {
	return x.Stream.SendMsg(m)
}
{{end}}

{{if or $clientSide $bidiSide}}
func (x *{{LowerFirst .ServiceName}}{{.RawName}}Server) Recv() ({{$arg.Type}}, error) {
	m := new({{NotPtr $arg.Type}})
	return m, x.Stream.RecvMsg(m)
}
{{end}}

{{- end}}
{{- /* define streaming struct end */}}
func new{{.ArgStructName}}() interface{} {
	return {{if not .GenArgResultStruct}}{{.PkgRefName}}.New{{.ArgStructName}}(){{else}}&{{.ArgStructName}}{}{{end}}
}
{{if not .Oneway}}
func new{{.ResStructName}}() interface{} {
	return {{if not .GenArgResultStruct}}{{.PkgRefName}}.New{{.ResStructName}}(){{else}}&{{.ResStructName}}{}{{end}}
}{{end}}

{{- if .GenArgResultStruct}}
{{$arg := index .Args 0}}
type {{.ArgStructName}} struct {
	Req {{$arg.Type}}
}

func (p *{{.ArgStructName}}) Marshal(out []byte) ([]byte, error) {
	if !p.IsSetReq() {
	   return out, fmt.Errorf("No req in {{.ArgStructName}}")
	}
	return proto.Marshal(p.Req)
}

func (p *{{.ArgStructName}}) Unmarshal(in []byte) error {
	msg := new({{NotPtr $arg.Type}})
	if err := proto.Unmarshal(in, msg); err != nil {
	   return err
	}
	p.Req = msg
	return nil
}

var {{.ArgStructName}}_Req_DEFAULT {{$arg.Type}}

func (p *{{.ArgStructName}}) GetReq() {{$arg.Type}} {
	if !p.IsSetReq() {
	   return {{.ArgStructName}}_Req_DEFAULT
	}
	return p.Req
}

func (p *{{.ArgStructName}}) IsSetReq() bool {
	return p.Req != nil
}

type {{.ResStructName}} struct {
	Success {{.Resp.Type}}
}

var {{.ResStructName}}_Success_DEFAULT {{.Resp.Type}}

func (p *{{.ResStructName}}) Marshal(out []byte) ([]byte, error) {
	if !p.IsSetSuccess() {
	   return out, fmt.Errorf("No req in {{.ResStructName}}")
	}
	return proto.Marshal(p.Success)
}

func (p *{{.ResStructName}}) Unmarshal(in []byte) error {
	msg := new({{NotPtr .Resp.Type}})
	if err := proto.Unmarshal(in, msg); err != nil {
	   return err
	}
	p.Success = msg
	return nil
}

func (p *{{.ResStructName}}) GetSuccess() {{.Resp.Type}} {
	if !p.IsSetSuccess() {
	   return {{.ResStructName}}_Success_DEFAULT
	}
	return p.Success
}

func (p *{{.ResStructName}}) SetSuccess(x interface{}) {
	p.Success = x.({{.Resp.Type}})
}

func (p *{{.ResStructName}}) IsSetSuccess() bool {
	return p.Success != nil
}
{{- end}}
{{end}}

type kClient struct {
	c client.Client
}

func newServiceClient(c client.Client) *kClient {
	return &kClient{
		c: c,
	}
}

{{range .AllMethods}}
{{- if or .ClientStreaming .ServerStreaming}}
{{- /* streaming logic */}}
func (p *kClient) {{.Name}}(ctx context.Context{{if not .ClientStreaming}}{{range .Args}}, {{LowerFirst .Name}} {{.Type}}{{end}}{{end}}) ({{.ServiceName}}_{{.RawName}}Client, error) {
	streamClient, ok := p.c.(client.Streaming)
	if !ok {
		return nil, fmt.Errorf("client not support streaming")
	}
	res := new(streaming.Result)
	err := streamClient.Stream(ctx, "{{.RawName}}", nil, res)
	if err != nil {
		return nil, err
	}
	stream := &{{LowerFirst .ServiceName}}{{.RawName}}Client{res.Stream}
	{{if not .ClientStreaming -}}
	if err := stream.Stream.SendMsg(req); err != nil {
		return nil, err
	}
	if err := stream.Stream.Close(); err != nil {
		return nil, err
	}
	{{end -}}
	return stream, nil
}
{{- else}}
func (p *kClient) {{.Name}}(ctx context.Context {{range .Args}}, {{.RawName}} {{.Type}}{{end}}) ({{if not .Void}}r {{.Resp.Type}}, {{end}}err error) {
	var _args {{if not .GenArgResultStruct}}{{.PkgRefName}}.{{end}}{{.ArgStructName}}
	{{range .Args -}}
	_args.{{.Name}} = {{.RawName}}
	{{end -}}
{{if .Void -}}
	{{if .Oneway -}}
	if err = p.c.Call(ctx, "{{.RawName}}", &_args, nil); err != nil {
		return
	}
	{{else -}}
	var _result {{if not .GenArgResultStruct}}{{.PkgRefName}}.{{end}}{{.ResStructName}}
	if err = p.c.Call(ctx, "{{.RawName}}", &_args, &_result); err != nil {
		return
	}
	{{end -}}
	return nil
{{else -}}
	var _result {{if not .GenArgResultStruct}}{{.PkgRefName}}.{{end}}{{.ResStructName}}
	if err = p.c.Call(ctx, "{{.RawName}}", &_args, &_result); err != nil {
		return
	}
	{{if .Exceptions -}}
	switch {
	{{range .Exceptions -}}
	case _result.{{.Name}} != nil:
		return r, _result.{{.Name}}
	{{end -}}
	}
	{{end -}}
	return _result.GetSuccess(), nil
{{end -}}
}
{{- end}}
{{end}}
`
