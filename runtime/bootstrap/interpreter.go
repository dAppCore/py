// Package bootstrap hosts the Tier 1 CorePy bootstrap interpreter.
//
// This runtime implements the binding contract described in
// plans/code/core/py/RFC.md so CorePy can validate module registration,
// import shape, and round-trip execution before the gpython dependency lands.
// TODO(corepy-gpython): replace this subset interpreter with LetheanNetwork/gpython.
//
//	interpreter := bootstrap.New()
//	output, err := interpreter.Run(`
//	    from core import echo
//	    print(echo("hello"))
//	`)
package bootstrap

import (
	"bytes"
	"fmt" // AX-6-exception: bootstrap parser uses fmt error formatting until gpython owns exception construction.
	"reflect"
	"slices"
	"strconv"
	"strings" // AX-6-exception: bootstrap parser needs tokenizer helpers beyond the current core string wrapper set.

	"dappco.re/go/py/runtime/internal/contract"
)

type Module = contract.Module
type Function = contract.Function

type functionReference struct {
	moduleName   string
	functionName string
}

type callableReference struct {
	moduleName     string
	functionName   string
	boundArguments []any
}

type KeywordArguments = contract.KeywordArguments
type BoundMethod = contract.BoundMethod
type AttributeResolver = contract.AttributeResolver
type ModuleReference = contract.ModuleReference

// Interpreter executes a small Python subset against registered modules.
type Interpreter struct {
	modules map[string]*Module
	order   []string
	output  *bytes.Buffer
}

// New creates an empty bootstrap interpreter with a root `core` module.
//
//	interpreter := bootstrap.New()
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

// Close releases interpreter resources. The bootstrap backend is in-memory only.
func (interpreter *Interpreter) Close() error {
	return nil
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
// - `import core`
// - `import core.fs as filesystem`
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
		statements, err := splitTopLevel(rawLine, ';')
		if err != nil {
			return "", fmt.Errorf("runtime.Run line %d: %w", lineNumber+1, err)
		}
		for _, rawStatement := range statements {
			line := strings.TrimSpace(rawStatement)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			switch {
			case strings.HasPrefix(line, "import "):
				if err := interpreter.executeDirectImport(line, namespace); err != nil {
					return "", fmt.Errorf("runtime.Run line %d: %w", lineNumber+1, err)
				}
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
	}

	return interpreter.output.String(), nil
}

func (interpreter *Interpreter) executeDirectImport(line string, namespace map[string]any) error {
	body := strings.TrimSpace(strings.TrimPrefix(line, "import "))
	if body == "" {
		return fmt.Errorf("import module cannot be empty")
	}

	for _, rawTarget := range strings.Split(body, ",") {
		moduleName, bindingName, hasAlias, err := parseImportBinding(rawTarget)
		if err != nil {
			return err
		}
		if _, ok := interpreter.modules[moduleName]; !ok {
			return fmt.Errorf("module %q is not registered", moduleName)
		}

		if hasAlias {
			namespace[bindingName] = ModuleReference{Name: moduleName}
			continue
		}

		rootName := strings.Split(moduleName, ".")[0]
		namespace[rootName] = ModuleReference{Name: rootName}
	}
	return nil
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
		name, bindingName, _, err := parseImportBinding(rawName)
		if err != nil {
			return err
		}
		exported, err := interpreter.resolveImport(moduleName, name)
		if err != nil {
			return err
		}
		namespace[bindingName] = exported
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
		value, err := unquoteStringLiteral(expression)
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
	if strings.HasPrefix(expression, "[") && strings.HasSuffix(expression, "]") {
		return interpreter.evaluateListLiteral(expression[1:len(expression)-1], namespace)
	}
	if strings.HasPrefix(expression, "{") && strings.HasSuffix(expression, "}") {
		return interpreter.evaluateDictLiteral(expression[1:len(expression)-1], namespace)
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
		callArguments := append(append([]any(nil), callable.boundArguments...), arguments...)
		return interpreter.Call(callable.moduleName, callable.functionName, callArguments...)
	}

	if value, ok := namespace[expression]; ok {
		return value, nil
	}
	if strings.Contains(expression, ".") {
		return interpreter.resolveValue(expression, namespace)
	}
	return nil, fmt.Errorf("unknown identifier %q", expression)
}

func (interpreter *Interpreter) evaluateListLiteral(body string, namespace map[string]any) (any, error) {
	parts, err := splitTopLevel(body, ',')
	if err != nil {
		return nil, err
	}
	if len(parts) == 0 {
		return []any{}, nil
	}

	values := make([]any, 0, len(parts))
	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			continue
		}
		value, err := interpreter.evaluateExpression(part, namespace)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, nil
}

func (interpreter *Interpreter) evaluateDictLiteral(body string, namespace map[string]any) (any, error) {
	parts, err := splitTopLevel(body, ',')
	if err != nil {
		return nil, err
	}
	if len(parts) == 0 {
		return map[string]any{}, nil
	}

	values := map[string]any{}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		separatorIndex := topLevelIndex(part, ':')
		if separatorIndex == -1 {
			return nil, fmt.Errorf("invalid dict item %q", part)
		}

		keyValue, err := interpreter.evaluateExpression(part[:separatorIndex], namespace)
		if err != nil {
			return nil, err
		}
		key, ok := keyValue.(string)
		if !ok {
			return nil, fmt.Errorf("dict key %q must evaluate to string, got %T", part[:separatorIndex], keyValue)
		}

		value, err := interpreter.evaluateExpression(part[separatorIndex+1:], namespace)
		if err != nil {
			return nil, err
		}
		values[key] = value
	}
	return values, nil
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
	keywordArguments := KeywordArguments{}
	seenKeywordArguments := false
	for _, part := range parts {
		if name, valueExpression, ok := splitKeywordArgument(part); ok {
			if _, exists := keywordArguments[name]; exists {
				return nil, fmt.Errorf("duplicate keyword argument %q", name)
			}
			value, err := interpreter.evaluateExpression(valueExpression, namespace)
			if err != nil {
				return nil, err
			}
			keywordArguments[name] = value
			seenKeywordArguments = true
			continue
		}
		if seenKeywordArguments {
			return nil, fmt.Errorf("positional argument cannot follow keyword arguments")
		}
		value, err := interpreter.evaluateExpression(part, namespace)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	if len(keywordArguments) > 0 {
		values = append(values, keywordArguments)
	}
	return values, nil
}

func (interpreter *Interpreter) resolveCallable(expression string, namespace map[string]any) (callableReference, error) {
	value, err := interpreter.resolveValue(expression, namespace)
	if err != nil {
		return callableReference{}, err
	}

	switch typed := value.(type) {
	case functionReference:
		return callableReference{
			moduleName:   typed.moduleName,
			functionName: typed.functionName,
		}, nil
	case BoundMethod:
		return callableReference{
			moduleName:     typed.ModuleName,
			functionName:   typed.FunctionName,
			boundArguments: append([]any(nil), typed.Arguments...),
		}, nil
	default:
		return callableReference{}, fmt.Errorf("%q is not callable", expression)
	}
}

func (interpreter *Interpreter) resolveValue(expression string, namespace map[string]any) (any, error) {
	parts := strings.Split(strings.TrimSpace(expression), ".")
	if len(parts) == 0 || parts[0] == "" {
		return nil, fmt.Errorf("unknown identifier %q", expression)
	}

	value, ok := namespace[parts[0]]
	if !ok {
		return nil, fmt.Errorf("unknown identifier %q", expression)
	}
	if len(parts) == 1 {
		return value, nil
	}

	currentPath := parts[0]
	for _, segment := range parts[1:] {
		switch typed := value.(type) {
		case ModuleReference:
			nextValue, err := interpreter.resolveImport(typed.Name, segment)
			if err != nil {
				return nil, err
			}
			value = nextValue
		case AttributeResolver:
			nextValue, ok := typed.ResolveAttribute(segment)
			if !ok {
				return nil, fmt.Errorf("%q does not export %q", currentPath, segment)
			}
			value = nextValue
		default:
			return nil, fmt.Errorf("%q does not export %q", currentPath, segment)
		}
		currentPath += "." + segment
	}
	return value, nil
}

func moduleLineage(moduleName string) []string {
	parts := strings.Split(moduleName, ".")
	var names []string
	for index := range parts {
		names = append(names, strings.Join(parts[:index+1], "."))
	}
	return names
}

func parseImportBinding(raw string) (moduleName string, bindingName string, hasAlias bool, err error) {
	fields := strings.Fields(strings.TrimSpace(raw))
	switch len(fields) {
	case 0:
		return "", "", false, fmt.Errorf("import name cannot be empty")
	case 1:
		return fields[0], fields[0], false, nil
	case 3:
		if fields[1] != "as" {
			return "", "", false, fmt.Errorf("invalid import syntax: %q", raw)
		}
		if fields[0] == "" || fields[2] == "" {
			return "", "", false, fmt.Errorf("invalid import syntax: %q", raw)
		}
		return fields[0], fields[2], true, nil
	default:
		return "", "", false, fmt.Errorf("invalid import syntax: %q", raw)
	}
}

func splitArguments(argumentBody string) ([]string, error) {
	return splitTopLevel(argumentBody, ',')
}

// SplitKeywordArguments separates positional arguments from a trailing
// KeywordArguments payload.
//
//	positional, keywordArguments := runtime.SplitKeywordArguments(arguments)
func splitTopLevel(value string, separator rune) ([]string, error) {
	var (
		parts   []string
		builder strings.Builder
		stack   []rune
		quote   rune
		escaped bool
	)

	for _, character := range value {
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
		case isOpenGrouping(character):
			stack = append(stack, character)
			builder.WriteRune(character)
		case isCloseGrouping(character):
			if len(stack) == 0 || stack[len(stack)-1] != matchingOpenGrouping(character) {
				return nil, fmt.Errorf("unbalanced grouping in %q", value)
			}
			stack = stack[:len(stack)-1]
			builder.WriteRune(character)
		case character == separator && len(stack) == 0:
			parts = append(parts, strings.TrimSpace(builder.String()))
			builder.Reset()
		default:
			builder.WriteRune(character)
		}
	}

	if quote != 0 {
		return nil, fmt.Errorf("unterminated string literal in %q", value)
	}
	if len(stack) != 0 {
		return nil, fmt.Errorf("unbalanced grouping in %q", value)
	}

	last := strings.TrimSpace(builder.String())
	if last != "" || strings.TrimSpace(value) == "" {
		parts = append(parts, last)
	}
	return parts, nil
}

func topLevelIndex(value string, target rune) int {
	quote := rune(0)
	escaped := false
	var stack []rune

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
		case character == target && len(stack) == 0:
			return index
		case isOpenGrouping(character):
			stack = append(stack, character)
		case isCloseGrouping(character):
			if len(stack) > 0 && stack[len(stack)-1] == matchingOpenGrouping(character) {
				stack = stack[:len(stack)-1]
			}
		}
	}

	return -1
}

func splitKeywordArgument(part string) (name string, value string, ok bool) {
	index := topLevelIndex(part, '=')
	if index == -1 {
		return "", "", false
	}

	name = strings.TrimSpace(part[:index])
	if !isIdentifier(name) {
		return "", "", false
	}
	return name, strings.TrimSpace(part[index+1:]), true
}

func isQuoted(value string) bool {
	return len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\''))
}

func unquoteStringLiteral(value string) (string, error) {
	if len(value) >= 2 && value[0] == '"' {
		return strconv.Unquote(value)
	}
	if len(value) < 2 || value[0] != '\'' || value[len(value)-1] != '\'' {
		return "", fmt.Errorf("expected quoted string")
	}

	var builder strings.Builder
	escaped := false
	for _, character := range value[1 : len(value)-1] {
		if escaped {
			switch character {
			case '\\', '\'':
				builder.WriteRune(character)
			case 'n':
				builder.WriteByte('\n')
			case 'r':
				builder.WriteByte('\r')
			case 't':
				builder.WriteByte('\t')
			default:
				builder.WriteRune('\\')
				builder.WriteRune(character)
			}
			escaped = false
			continue
		}
		if character == '\\' {
			escaped = true
			continue
		}
		builder.WriteRune(character)
	}
	if escaped {
		return "", fmt.Errorf("unterminated escape sequence")
	}
	return builder.String(), nil
}

func isIdentifier(value string) bool {
	if value == "" {
		return false
	}

	for index, character := range value {
		if index == 0 {
			if (character < 'a' || character > 'z') && (character < 'A' || character > 'Z') && character != '_' {
				return false
			}
			continue
		}
		if (character < 'a' || character > 'z') && (character < 'A' || character > 'Z') && (character < '0' || character > '9') && character != '_' {
			return false
		}
	}
	return true
}

func isOpenGrouping(character rune) bool {
	return character == '(' || character == '[' || character == '{'
}

func isCloseGrouping(character rune) bool {
	return character == ')' || character == ']' || character == '}'
}

func matchingOpenGrouping(character rune) rune {
	switch character {
	case ')':
		return '('
	case ']':
		return '['
	case '}':
		return '{'
	default:
		return 0
	}
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
	case string:
		return typed
	default:
		return formatCompositeValue(typed, false)
	}
}

func formatCompositeValue(value any, nested bool) string {
	switch typed := value.(type) {
	case nil:
		return "None"
	case bool:
		if typed {
			return "True"
		}
		return "False"
	case string:
		if nested {
			return strconv.Quote(typed)
		}
		return typed
	}

	reflected := reflect.ValueOf(value)
	if !reflected.IsValid() {
		return "None"
	}

	switch reflected.Kind() {
	case reflect.Slice, reflect.Array:
		parts := make([]string, 0, reflected.Len())
		for index := 0; index < reflected.Len(); index++ {
			parts = append(parts, formatCompositeValue(reflected.Index(index).Interface(), true))
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case reflect.Map:
		if reflected.Type().Key().Kind() != reflect.String {
			return fmt.Sprint(value)
		}

		keys := make([]string, 0, reflected.Len())
		for _, keyValue := range reflected.MapKeys() {
			keys = append(keys, keyValue.String())
		}
		slices.Sort(keys)

		parts := make([]string, 0, len(keys))
		for _, key := range keys {
			part := strconv.Quote(key) + ": " + formatCompositeValue(reflected.MapIndex(reflect.ValueOf(key)).Interface(), true)
			parts = append(parts, part)
		}
		return "{" + strings.Join(parts, ", ") + "}"
	default:
		return fmt.Sprint(value)
	}
}
