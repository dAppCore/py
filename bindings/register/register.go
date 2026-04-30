package register

import (
	actionbinding "dappco.re/go/py/bindings/action"
	"dappco.re/go/py/bindings/agent"
	"dappco.re/go/py/bindings/api"
	"dappco.re/go/py/bindings/array"
	"dappco.re/go/py/bindings/cache"
	"dappco.re/go/py/bindings/config"
	"dappco.re/go/py/bindings/container"
	cryptobinding "dappco.re/go/py/bindings/crypto"
	"dappco.re/go/py/bindings/data"
	dnsbinding "dappco.re/go/py/bindings/dns"
	"dappco.re/go/py/bindings/echo"
	entitlementbinding "dappco.re/go/py/bindings/entitlement"
	"dappco.re/go/py/bindings/err"
	"dappco.re/go/py/bindings/fs"
	i18nbinding "dappco.re/go/py/bindings/i18n"
	infobinding "dappco.re/go/py/bindings/info"
	"dappco.re/go/py/bindings/json"
	"dappco.re/go/py/bindings/log"
	mathbinding "dappco.re/go/py/bindings/math"
	"dappco.re/go/py/bindings/mcp"
	"dappco.re/go/py/bindings/medium"
	"dappco.re/go/py/bindings/options"
	pathbinding "dappco.re/go/py/bindings/path"
	"dappco.re/go/py/bindings/process"
	registrybinding "dappco.re/go/py/bindings/registry"
	scmbinding "dappco.re/go/py/bindings/scm"
	"dappco.re/go/py/bindings/service"
	stdlibbinding "dappco.re/go/py/bindings/stdlib"
	"dappco.re/go/py/bindings/store"
	stringsbinding "dappco.re/go/py/bindings/strings"
	taskbinding "dappco.re/go/py/bindings/task"
	"dappco.re/go/py/bindings/ws"
	"dappco.re/go/py/runtime"
)

// ModuleSpec describes one CorePy binding module and its registration hook.
type ModuleSpec struct {
	Name     string
	Register func(runtime.Interpreter) error
}

// DefaultModuleSpecs returns the binding registry used by Tier 1 backends.
func DefaultModuleSpecs() []ModuleSpec {
	return []ModuleSpec{
		{Name: "core.action", Register: actionbinding.Register},
		{Name: "core.agent", Register: agent.Register},
		{Name: "core.api", Register: api.Register},
		{Name: "core.array", Register: array.Register},
		{Name: "core.cache", Register: cache.Register},
		{Name: "core.container", Register: container.Register},
		{Name: "core.entitlement", Register: entitlementbinding.Register},
		{Name: "core.echo", Register: echo.Register},
		{Name: "core.fs", Register: fs.Register},
		{Name: "core.json", Register: json.Register},
		{Name: "core.medium", Register: medium.Register},
		{Name: "core.options", Register: options.Register},
		{Name: "core.path", Register: pathbinding.Register},
		{Name: "core.process", Register: process.Register},
		{Name: "core.config", Register: config.Register},
		{Name: "core.data", Register: data.Register},
		{Name: "core.i18n", Register: i18nbinding.Register},
		{Name: "core.info", Register: infobinding.Register},
		{Name: "core.service", Register: service.Register},
		{Name: "core.log", Register: log.Register},
		{Name: "core.err", Register: err.Register},
		{Name: "core.mcp", Register: mcp.Register},
		{Name: "core.crypto", Register: cryptobinding.Register},
		{Name: "core.dns", Register: dnsbinding.Register},
		{Name: "core.math", Register: mathbinding.Register},
		{Name: "core.registry", Register: registrybinding.Register},
		{Name: "core.scm", Register: scmbinding.Register},
		{Name: "core.store", Register: store.Register},
		{Name: "core.strings", Register: stringsbinding.Register},
		{Name: "core.task", Register: taskbinding.Register},
		{Name: "core.ws", Register: ws.Register},
	}
}

// DefaultModuleNames returns the canonical default binding names.
func DefaultModuleNames() []string {
	specs := DefaultModuleSpecs()
	names := make([]string, 0, len(specs))
	for _, spec := range specs {
		names = append(names, spec.Name)
	}
	return names
}

// DefaultShadowModuleSpecs returns Python stdlib-shaped aliases that Tier 1
// resolves to Core-backed primitives.
func DefaultShadowModuleSpecs() []ModuleSpec {
	stdlibSpecs := stdlibbinding.Specs()
	specs := make([]ModuleSpec, 0, len(stdlibSpecs))
	for _, spec := range stdlibSpecs {
		specs = append(specs, ModuleSpec{Name: spec.Name, Register: spec.Register})
	}
	return specs
}

// DefaultShadowModuleNames returns the canonical stdlib shadow names.
func DefaultShadowModuleNames() []string {
	specs := DefaultShadowModuleSpecs()
	names := make([]string, 0, len(specs))
	for _, spec := range specs {
		names = append(names, spec.Name)
	}
	return names
}

// DefaultModules registers the bootstrap CorePy module set.
//
//	register.DefaultModules(interpreter)
func DefaultModules(interpreter runtime.Interpreter) error {
	for _, spec := range DefaultModuleSpecs() {
		if err := spec.Register(interpreter); err != nil {
			return err
		}
	}
	for _, spec := range DefaultShadowModuleSpecs() {
		if err := spec.Register(interpreter); err != nil {
			return err
		}
	}
	return nil
}
