package register

import (
	actionbinding "github.com/dappcore/py/go/bindings/action"
	"github.com/dappcore/py/go/bindings/agent"
	"github.com/dappcore/py/go/bindings/api"
	"github.com/dappcore/py/go/bindings/array"
	"github.com/dappcore/py/go/bindings/cache"
	"github.com/dappcore/py/go/bindings/config"
	"github.com/dappcore/py/go/bindings/container"
	cryptobinding "github.com/dappcore/py/go/bindings/crypto"
	"github.com/dappcore/py/go/bindings/data"
	dnsbinding "github.com/dappcore/py/go/bindings/dns"
	"github.com/dappcore/py/go/bindings/echo"
	entitlementbinding "github.com/dappcore/py/go/bindings/entitlement"
	"github.com/dappcore/py/go/bindings/err"
	"github.com/dappcore/py/go/bindings/fs"
	i18nbinding "github.com/dappcore/py/go/bindings/i18n"
	infobinding "github.com/dappcore/py/go/bindings/info"
	"github.com/dappcore/py/go/bindings/json"
	"github.com/dappcore/py/go/bindings/log"
	mathbinding "github.com/dappcore/py/go/bindings/math"
	"github.com/dappcore/py/go/bindings/mcp"
	"github.com/dappcore/py/go/bindings/medium"
	"github.com/dappcore/py/go/bindings/options"
	pathbinding "github.com/dappcore/py/go/bindings/path"
	"github.com/dappcore/py/go/bindings/process"
	registrybinding "github.com/dappcore/py/go/bindings/registry"
	scmbinding "github.com/dappcore/py/go/bindings/scm"
	"github.com/dappcore/py/go/bindings/service"
	stdlibbinding "github.com/dappcore/py/go/bindings/stdlib"
	"github.com/dappcore/py/go/bindings/store"
	stringsbinding "github.com/dappcore/py/go/bindings/strings"
	taskbinding "github.com/dappcore/py/go/bindings/task"
	"github.com/dappcore/py/go/bindings/ws"
	"github.com/dappcore/py/go/runtime"
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
