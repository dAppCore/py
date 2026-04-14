package register

import (
	"dappco.re/go/py/bindings/config"
	"dappco.re/go/py/bindings/data"
	"dappco.re/go/py/bindings/echo"
	"dappco.re/go/py/bindings/err"
	"dappco.re/go/py/bindings/fs"
	"dappco.re/go/py/bindings/json"
	"dappco.re/go/py/bindings/log"
	"dappco.re/go/py/bindings/options"
	"dappco.re/go/py/bindings/service"
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
		options.Register,
		config.Register,
		data.Register,
		service.Register,
		log.Register,
		err.Register,
	} {
		if err := registerModule(interpreter); err != nil {
			return err
		}
	}
	return nil
}
