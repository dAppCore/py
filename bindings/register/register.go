package register

import (
	"dappco.re/go/py/bindings/config"
	"dappco.re/go/py/bindings/data"
	"dappco.re/go/py/bindings/echo"
	"dappco.re/go/py/bindings/err"
	"dappco.re/go/py/bindings/fs"
	"dappco.re/go/py/bindings/json"
	"dappco.re/go/py/bindings/log"
	mathbinding "dappco.re/go/py/bindings/math"
	"dappco.re/go/py/bindings/medium"
	"dappco.re/go/py/bindings/options"
	pathbinding "dappco.re/go/py/bindings/path"
	"dappco.re/go/py/bindings/process"
	"dappco.re/go/py/bindings/service"
	stringsbinding "dappco.re/go/py/bindings/strings"
	"dappco.re/go/py/runtime"
)

// DefaultModules registers the bootstrap CorePy module set.
//
//	register.DefaultModules(interpreter)
func DefaultModules(interpreter *runtime.Interpreter) error {
	for _, registerModule := range []func(*runtime.Interpreter) error{
		echo.Register,
		fs.Register,
		json.Register,
		medium.Register,
		options.Register,
		pathbinding.Register,
		process.Register,
		config.Register,
		data.Register,
		service.Register,
		log.Register,
		err.Register,
		mathbinding.Register,
		stringsbinding.Register,
	} {
		if err := registerModule(interpreter); err != nil {
			return err
		}
	}
	return nil
}
