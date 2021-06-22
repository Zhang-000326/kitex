package tpl

import "strings"

var (
	serverExtendTpl []string
	serverOptionTpl []string
)

// GetServerTpl returns the template for generating server.go.
func GetServerTpl() string {
	return `// Code generated by Kitex {{.Version}}. DO NOT EDIT.
package {{ToLower .ServiceName}}

import (
    {{- range $path, $alias := .Imports}}
    {{$alias }}"{{$path}}"
    {{- end}}
)

// NewServer creates a server.Server with the given handler and options.
func NewServer(handler {{call .ServiceTypeName}}, opts ...server.Option) server.Server {
    var options []server.Option
` + strings.Join(serverOptionTpl, "\n") + `
    options = append(options, opts...)

    svr := server.NewServer(options...)
    if err := svr.RegisterService(serviceInfo(), handler); err != nil {
            panic(err)
    }
    return svr
}
` + strings.Join(serverExtendTpl, "\n")
}