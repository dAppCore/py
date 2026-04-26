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
	"dappco.re/go/py/bindings/store"
	stringsbinding "dappco.re/go/py/bindings/strings"
	taskbinding "dappco.re/go/py/bindings/task"
	"dappco.re/go/py/bindings/ws"
	"dappco.re/go/py/runtime"
)

// DefaultModules registers the bootstrap CorePy module set.
//
//	register.DefaultModules(interpreter)
func DefaultModules(interpreter runtime.Interpreter) error {
	for _, registerModule := range []func(runtime.Interpreter) error{
		actionbinding.Register,
		agent.Register,
		api.Register,
		array.Register,
		cache.Register,
		container.Register,
		entitlementbinding.Register,
		echo.Register,
		fs.Register,
		json.Register,
		medium.Register,
		options.Register,
		pathbinding.Register,
		process.Register,
		config.Register,
		data.Register,
		i18nbinding.Register,
		infobinding.Register,
		service.Register,
		log.Register,
		err.Register,
		mcp.Register,
		cryptobinding.Register,
		dnsbinding.Register,
		mathbinding.Register,
		registrybinding.Register,
		scmbinding.Register,
		store.Register,
		stringsbinding.Register,
		taskbinding.Register,
		ws.Register,
	} {
		if err := registerModule(interpreter); err != nil {
			return err
		}
	}
	return nil
}
