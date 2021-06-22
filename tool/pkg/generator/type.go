package generator

import (
	"bytes"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/cloudwego/kitex/tool/pkg/util"
)

// File .
type File struct {
	Name    string
	Content string
}

// PSM .
type PSM struct {
	Product   string
	Subsystem string
	Module    string
}

// PackageInfo contains informations to generate a package for a service.
type PackageInfo struct {
	Namespace    string            // a dot-seperated string for generating service package under kitex_gen
	Dependencies map[string]string // package name => import path, used for searching imports
	*ServiceInfo                   // the target service

	// the belowing fields will be filled and used by the gnerator
	Codec            string
	Version          string
	RealServiceName  string
	Imports          map[string]string // import path => alias
	ExternalKitexGen string
	PSM
	Features []feature
}

// AddImport .
func (p *PackageInfo) AddImport(pkg, path string) {
	if pkg != "" {
		if p.ExternalKitexGen != "" && strings.Contains(path, KitexGenPath) {
			parts := strings.Split(path, KitexGenPath)
			path = filepath.Join(p.ExternalKitexGen, parts[len(parts)-1])
		}
		if strings.HasSuffix(path, "/"+pkg) || path == pkg {
			p.Imports[path] = ""
		} else {
			p.Imports[path] = pkg
		}
	}
}

// AddImports .
func (p *PackageInfo) AddImports(pkgs ...string) {
	for _, pkg := range pkgs {
		if path, ok := p.Dependencies[pkg]; ok {
			p.AddImport(pkg, path)
		} else {
			p.AddImport(pkg, pkg)
		}
	}
}

// PkgInfo .
type PkgInfo struct {
	PkgRefName string
	ImportPath string
}

// ServiceInfo .
type ServiceInfo struct {
	PkgInfo
	ServiceName     string
	RawServiceName  string
	ServiceTypeName func() string
	Base            *ServiceInfo
	Methods         []*MethodInfo
	CombineServices []*ServiceInfo
	HasStreaming    bool
}

// AllMethods returns all methods that the service have.
func (s *ServiceInfo) AllMethods() (ms []*MethodInfo) {
	ms = s.Methods
	for base := s.Base; base != nil; base = base.Base {
		ms = append(base.Methods, ms...)
	}
	return ms
}

// MethodInfo .
type MethodInfo struct {
	PkgInfo
	ServiceName            string
	Name                   string
	RawName                string
	Oneway                 bool
	Void                   bool
	Args                   []*Parameter
	Resp                   *Parameter
	Exceptions             []*Parameter
	ArgStructName          string
	ResStructName          string
	IsResponseNeedRedirect bool // int -> int*
	GenArgResultStruct     bool
	ClientStreaming        bool
	ServerStreaming        bool
}

// Parameter .
type Parameter struct {
	Deps    []PkgInfo
	Name    string
	RawName string
	Type    string // *PkgA.StructB
}

var funcs = map[string]interface{}{
	"ToLower":    strings.ToLower,
	"LowerFirst": util.LowerFirst,
	"UpperFirst": util.UpperFirst,
	"NotPtr":     util.NotPtr,
	"HasFeature": HasFeature,
}

// Task .
type Task struct {
	Name string
	Path string
	Text string
	*template.Template
}

// Build .
func (t *Task) Build() error {
	x, err := template.New(t.Name).Funcs(funcs).Parse(t.Text)
	if err != nil {
		return err
	}
	t.Template = x
	return nil
}

// Render .
func (t *Task) Render(data interface{}) (*File, error) {
	if t.Template == nil {
		err := t.Build()
		if err != nil {
			return nil, err
		}
	}

	var buf bytes.Buffer
	err := t.ExecuteTemplate(&buf, t.Name, data)
	if err != nil {
		return nil, err
	}
	return &File{t.Path, buf.String()}, nil
}