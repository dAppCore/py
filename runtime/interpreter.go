// Package runtime hosts the Tier 1 CorePy bootstrap interpreter.
//
// This runtime implements the binding contract described in
// plans/code/core/py/RFC.md so CorePy can validate module registration,
// import shape, and round-trip execution before the gpython dependency lands.
//
//	interpreter := runtime.New()
//	output, err := interpreter.Run(`
//	    from core import echo
//	    print(echo("hello"))
//	`)
package runtime

import (
	"bytes"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// Function is a Python-callable binding exposed by a module.
//
//	module := runtime.Module{
//	    Name: "core",
//	    Functions: map[string]runtime.Function{
//	        "echo": func(arguments ...any) (any, error) { return arguments[0], nil },
//	    },
//	}
type Function func(arguments ...any) (any, error)

// Module defines a registered CorePy module.
//
//	runtime.Module{
//	    Name: "core.fs",
//	    Documentation: "Filesystem primitives",
//	    Functions: map[string]runtime.Function{"read_file": readFile},
//	}
type Module struct {
	Name          string
	Documentation string
	Functions     map[string]Function
}

type functionReference struct {
	moduleName   string
	functionName string
}

// ModuleReference is an imported module handle inside the bootstrap runtime.
//
//	from core import fs
//	print(fs.read_file("/tmp/demo.txt"))
type ModuleReference struct {
	Name string
}

// Interpreter executes a small Python subset against registered modules.
type Interpreter struct {
	modules map[string]*Module
	order   []string
	output  *bytes.Buffer
}

// New creates an empty interpreter with a root `core` module.
//
//	interpreter := runtime.New()
func New() *Interpreter {
	interpreter := &Interpreter{
		modules: map[string]*Module{},
		output:  &bytes.Buffer{},
	}
	_ = interpreter.RegisterModule(Module{
		Name:          "core",
		Documentation: "Root CorePy module",
	})
	return interpreter
}

// RegisterModule registers or extends a module by name.
//
//	interpreter.RegisterModule(runtime.Module{Name: "core", Functions: functions})
func (interpreter *Interpreter) RegisterModule(module Module) error {
	moduleName := strings.TrimSpace(module.Name)
	if moduleName == "" {
		return fmt.Errorf("runtime.RegisterModule: module name cannot be empty")
	}

	names := moduleLineage(moduleName)
	for _, name := range names {
		if _, ok := interpreter.modules[name]; ok {
			continue
		}
		interpreter.modules[name] = &Module{
			Name:      name,
			Functions: map[string]Function{},
		}
		interpreter.order = append(interpreter.order, name)
	}

	registered := interpreter.modules[moduleName]
	if module.Documentation != "" {
		registered.Documentation = module.Documentation
	}
	for functionName, function := range module.Functions {
		if strings.TrimSpace(functionName) == "" {
			return fmt.Errorf("runtime.RegisterModule(%s): function name cannot be empty", moduleName)
		}
		if function == nil {
			return fmt.Errorf("runtime.RegisterModule(%s): function %s is nil", moduleName, functionName)
		}
		registered.Functions[functionName] = function
	}
	return nil
}

// Modules returns registered module names in registration order.
//
//	names := interpreter.Modules()
func (interpreter *Interpreter) Modules() []string {
	return slices.Clone(interpreter.order)
}

// Call invokes a registered function directly.
//
//	value, err := interpreter.Call("core.fs", "read_file", "/tmp/demo.txt")
func (interpreter *Interpreter) Call(moduleName, functionName string, arguments ...any) (any, error) {
	module, ok := interpreter.modules[moduleName]
	if !ok {
		return nil, fmt.Errorf("runtime.Call: module %q is not registered", moduleName)
	}
	function, ok := module.Functions[functionName]
	if !ok {
		return nil, fmt.Errorf("runtime.Call: function %q is not registered in %q", functionName, moduleName)
	}
	return function(arguments...)
}

// Run executes a small Python subset used by the bootstrap integration tests.
//
// Supported statements:
// - `from core import echo, fs`
// - `name = expression`
// - `print(expression)`
//
//	output, err := interpreter.Run(`
//	    from core import echo
//	    print(echo("hello"))
//	`)
func (interpreter *Interpreter) Run(script string) (string, error) {
	interpreter.output.Reset()
	namespace := map[string]any{}

	for lineNumber, rawLine := range strings.Split(script, "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		switch {
		case strings.HasPrefix(line, "from "):
			if err := interpreter.executeImport(line, namespace); err != nil {
				return "", fmt.Errorf("runtime.Run line %d: %w", lineNumber+1, err)
			}
		case strings.HasPrefix(line, "print(") && strings.HasSuffix(line, ")"):
			expression := strings.TrimSuffix(strings.TrimPrefix(line, "print("), ")")
			value, err := interpreter.evaluateExpression(expression, namespace)
			if err != nil {
				return "", fmt.Errorf("runtime.Run line %d: %w", lineNumber+1, err)
			}
			if _, err := fmt.Fprintln(interpreter.output, formatValue(value)); err != nil {
				return "", fmt.Errorf("runtime.Run line %d: write output: %w", lineNumber+1, err)
			}
		default:
			index := topLevelIndex(line, '=')
			if index == -1 {
				if _, err := interpreter.evaluateExpression(line, namespace); err != nil {
					return "", fmt.Errorf("runtime.Run line %d: %w", lineNumber+1, err)
				}
				continue
			}

			name := strings.TrimSpace(line[:index])
			if name == "" {
				return "", fmt.Errorf("runtime.Run line %d: assignment target cannot be empty", lineNumber+1)
			}
			expression := strings.TrimSpace(line[index+1:])
			value, err := interpreter.evaluateExpression(expression, namespace)
			if err != nil {
				return "", fmt.Errorf("runtime.Run line %d: %w", lineNumber+1, err)
			}
			namespace[name] = value
		}
	}

	return interpreter.output.String(), nil
}

func (interpreter *Interpreter) executeImport(line string, namespace map[string]any) error {
	body := strings.TrimSpace(strings.TrimPrefix(line, "from "))
	parts := strings.SplitN(body, " import ", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid import syntax: %q", line)
	}

	moduleName := strings.TrimSpace(parts[0])
	if moduleName == "" {
		return fmt.Errorf("import module cannot be empty")
	}
	if _, ok := interpreter.modules[moduleName]; !ok {
		return fmt.Errorf("module %q is not registered", moduleName)
	}

	for _, rawName := range strings.Split(parts[1], ",") {
		name := strings.TrimSpace(rawName)
		if name == "" {
			return fmt.Errorf("import name cannot be empty")
		}
		exported, err := interpreter.resolveImport(moduleName, name)
		if err != nil {
			return err
		}
		namespace[name] = exported
	}
	return nil
}

func (interpreter *Interpreter) resolveImport(moduleName, name string) (any, error) {
	module := interpreter.modules[moduleName]
	if _, ok := module.Functions[name]; ok {
		return functionReference{moduleName: moduleName, functionName: name}, nil
	}

	childName := moduleName + "." + name
	if _, ok := interpreter.modules[childName]; ok {
		return ModuleReference{Name: childName}, nil
	}

	return nil, fmt.Errorf("module %q does not export %q", moduleName, name)
}

func (interpreter *Interpreter) evaluateExpression(expression string, namespace map[string]any) (any, error) {
	expression = strings.TrimSpace(expression)
	if expression == "" {
		return nil, fmt.Errorf("expression cannot be empty")
	}

	if isQuoted(expression) {
		value, err := strconv.Unquote(expression)
		if err != nil {
			return nil, fmt.Errorf("invalid string literal %q: %w", expression, err)
		}
		return value, nil
	}

	if expression == "True" {
		return true, nil
	}
	if expression == "False" {
		return false, nil
	}
	if expression == "None" {
		return nil, nil
	}
	if integerValue, err := strconv.Atoi(expression); err == nil {
		return integerValue, nil
	}
	if floatValue, err := strconv.ParseFloat(expression, 64); err == nil && strings.ContainsAny(expression, ".eE") {
		return floatValue, nil
	}

	if openIndex := topLevelIndex(expression, '('); openIndex != -1 && strings.HasSuffix(expression, ")") {
		callableExpression := strings.TrimSpace(expression[:openIndex])
		argumentBody := strings.TrimSpace(expression[openIndex+1 : len(expression)-1])
		arguments, err := interpreter.evaluateArguments(argumentBody, namespace)
		if err != nil {
			return nil, err
		}
		callable, err := interpreter.resolveCallable(callableExpression, namespace)
		if err != nil {
			return nil, err
		}
		return interpreter.Call(callable.moduleName, callable.functionName, arguments...)
	}

	value, ok := namespace[expression]
	if !ok {
		return nil, fmt.Errorf("unknown identifier %q", expression)
	}
	return value, nil
}

func (interpreter *Interpreter) evaluateArguments(argumentBody string, namespace map[string]any) ([]any, error) {
	if strings.TrimSpace(argumentBody) == "" {
		return nil, nil
	}

	parts, err := splitArguments(argumentBody)
	if err != nil {
		return nil, err
	}
	values := make([]any, 0, len(parts))
	for _, part := range parts {
		value, err := interpreter.evaluateExpression(part, namespace)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, nil
}

func (interpreter *Interpreter) resolveCallable(expression string, namespace map[string]any) (functionReference, error) {
	parts := strings.Split(expression, ".")
	if len(parts) == 0 {
		return functionReference{}, fmt.Errorf("call target cannot be empty")
	}

	value, ok := namespace[parts[0]]
	if !ok {
		return functionReference{}, fmt.Errorf("unknown callable %q", expression)
	}

	if len(parts) == 1 {
		callable, ok := value.(functionReference)
		if !ok {
			return functionReference{}, fmt.Errorf("%q is not callable", expression)
		}
		return callable, nil
	}

	moduleReference, ok := value.(ModuleReference)
	if !ok {
		return functionReference{}, fmt.Errorf("%q does not reference a module", parts[0])
	}

	moduleName := moduleReference.Name
	for _, segment := range parts[1 : len(parts)-1] {
		moduleName += "." + segment
		if _, ok := interpreter.modules[moduleName]; !ok {
			return functionReference{}, fmt.Errorf("module %q is not registered", moduleName)
		}
	}

	return functionReference{
		moduleName:   moduleName,
		functionName: parts[len(parts)-1],
	}, nil
}

func moduleLineage(moduleName string) []string {
	parts := strings.Split(moduleName, ".")
	var names []string
	for index := range parts {
		names = append(names, strings.Join(parts[:index+1], "."))
	}
	return names
}

func splitArguments(argumentBody string) ([]string, error) {
	var (
		arguments []string
		builder   strings.Builder
		depth     int
		quote     rune
		escaped   bool
	)

	for _, character := range argumentBody {
		switch {
		case quote != 0:
			builder.WriteRune(character)
			if escaped {
				escaped = false
				continue
			}
			if character == '\\' {
				escaped = true
				continue
			}
			if character == quote {
				quote = 0
			}
		case character == '"' || character == '\'':
			quote = character
			builder.WriteRune(character)
		case character == '(':
			depth++
			builder.WriteRune(character)
		case character == ')':
			depth--
			if depth < 0 {
				return nil, fmt.Errorf("unbalanced parentheses in %q", argumentBody)
			}
			builder.WriteRune(character)
		case character == ',' && depth == 0:
			arguments = append(arguments, strings.TrimSpace(builder.String()))
			builder.Reset()
		default:
			builder.WriteRune(character)
		}
	}

	if quote != 0 {
		return nil, fmt.Errorf("unterminated string literal in %q", argumentBody)
	}
	if depth != 0 {
		return nil, fmt.Errorf("unbalanced parentheses in %q", argumentBody)
	}

	last := strings.TrimSpace(builder.String())
	if last != "" {
		arguments = append(arguments, last)
	}
	return arguments, nil
}

func topLevelIndex(value string, target rune) int {
	depth := 0
	quote := rune(0)
	escaped := false

	for index, character := range value {
		switch {
		case quote != 0:
			if escaped {
				escaped = false
				continue
			}
			if character == '\\' {
				escaped = true
				continue
			}
			if character == quote {
				quote = 0
			}
		case character == '"' || character == '\'':
			quote = character
		case character == target && depth == 0:
			return index
		case character == '(':
			depth++
		case character == ')':
			if depth > 0 {
				depth--
			}
		}
	}

	return -1
}

func isQuoted(value string) bool {
	return len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\''))
}

func formatValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return "None"
	case bool:
		if typed {
			return "True"
		}
		return "False"
	default:
		return fmt.Sprint(typed)
	}
}
