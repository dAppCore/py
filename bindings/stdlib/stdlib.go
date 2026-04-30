// Package stdlib exposes small Python-standard-library-shaped modules backed
// by CorePy primitives.
//
// Tier 1 gpython does not carry CPython's C-backed stdlib modules. These
// shadows make common imports such as `import os`, `import json`, and
// `import subprocess` resolve to Core-backed helpers while keeping the broader
// module contract explicit.
package stdlib

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt" // AX-6-exception: stdlib shadow helpers report Python-shaped argument errors.
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	core "dappco.re/go/core"
	"dappco.re/go/py/bindings/typemap"
	"dappco.re/go/py/runtime"
)

// Spec describes one compatibility module.
type Spec struct {
	Name     string
	Register func(runtime.Interpreter) error
}

// Specs returns the Tier 1 stdlib shadow modules.
func Specs() []Spec {
	return []Spec{
		{Name: "base64", Register: registerBase64},
		{Name: "hashlib", Register: registerHashlib},
		{Name: "json", Register: registerJSON},
		{Name: "logging", Register: registerLogging},
		{Name: "os", Register: registerOS},
		{Name: "os.path", Register: registerOSPath},
		{Name: "socket", Register: registerSocket},
		{Name: "subprocess", Register: registerSubprocess},
	}
}

// Register registers all stdlib shadow modules.
func Register(interpreter runtime.Interpreter) error {
	for _, spec := range Specs() {
		if err := spec.Register(interpreter); err != nil {
			return err
		}
	}
	return nil
}

func registerJSON(interpreter runtime.Interpreter) error {
	return interpreter.RegisterModule(runtime.Module{
		Name:          "json",
		Documentation: "Tier 1 JSON stdlib shadow backed by core JSON helpers",
		Functions: map[string]runtime.Function{
			"dumps": jsonDumps,
			"loads": jsonLoads,
		},
	})
}

func jsonDumps(arguments ...any) (any, error) {
	if len(arguments) != 1 {
		return nil, core.E("json.dumps", "expected exactly one argument", nil)
	}
	return core.JSONMarshalString(arguments[0]), nil
}

func jsonLoads(arguments ...any) (any, error) {
	text, err := typemap.ExpectString(arguments, 0, "json.loads")
	if err != nil {
		return nil, err
	}
	var value any
	if _, err := typemap.ResultValue(core.JSONUnmarshalString(text, &value), "json.loads"); err != nil {
		return nil, err
	}
	return value, nil
}

func registerOS(interpreter runtime.Interpreter) error {
	return interpreter.RegisterModule(runtime.Module{
		Name:          "os",
		Documentation: "Tier 1 os stdlib shadow backed by core fs/process/path helpers",
		Functions: map[string]runtime.Function{
			"getcwd":   osGetcwd,
			"getenv":   osGetenv,
			"listdir":  osListdir,
			"makedirs": osMakedirs,
			"remove":   osRemove,
			"system":   osSystem,
		},
	})
}

func osGetcwd(arguments ...any) (any, error) {
	return os.Getwd()
}

func osGetenv(arguments ...any) (any, error) {
	key, err := typemap.ExpectString(arguments, 0, "os.getenv")
	if err != nil {
		return nil, err
	}
	if value, ok := os.LookupEnv(key); ok {
		return value, nil
	}
	if len(arguments) > 1 {
		return arguments[1], nil
	}
	return "", nil
}

func osListdir(arguments ...any) (any, error) {
	path, err := typemap.ExpectString(arguments, 0, "os.listdir")
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	slices.Sort(names)
	return names, nil
}

func osMakedirs(arguments ...any) (any, error) {
	path, err := typemap.ExpectString(arguments, 0, "os.makedirs")
	if err != nil {
		return nil, err
	}
	mode := 0755
	if len(arguments) > 1 {
		mode, err = typemap.ExpectInt(arguments, 1, "os.makedirs")
		if err != nil {
			return nil, err
		}
	}
	return nil, os.MkdirAll(path, os.FileMode(mode))
}

func osRemove(arguments ...any) (any, error) {
	path, err := typemap.ExpectString(arguments, 0, "os.remove")
	if err != nil {
		return nil, err
	}
	return nil, os.Remove(path)
}

func osSystem(arguments ...any) (any, error) {
	command, err := typemap.ExpectString(arguments, 0, "os.system")
	if err != nil {
		return nil, err
	}
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode(), nil
		}
		return 1, err
	}
	return 0, nil
}

func registerOSPath(interpreter runtime.Interpreter) error {
	return interpreter.RegisterModule(runtime.Module{
		Name:          "os.path",
		Documentation: "Tier 1 os.path stdlib shadow backed by core path helpers",
		Functions: map[string]runtime.Function{
			"abspath":  osPathAbsPath,
			"basename": osPathBasename,
			"dirname":  osPathDirname,
			"exists":   osPathExists,
			"isabs":    osPathIsAbs,
			"join":     osPathJoin,
		},
	})
}

func osPathAbsPath(arguments ...any) (any, error) {
	path, err := typemap.ExpectString(arguments, 0, "os.path.abspath")
	if err != nil {
		return nil, err
	}
	return filepath.Abs(path)
}

func osPathBasename(arguments ...any) (any, error) {
	path, err := typemap.ExpectString(arguments, 0, "os.path.basename")
	if err != nil {
		return nil, err
	}
	return filepath.Base(path), nil
}

func osPathDirname(arguments ...any) (any, error) {
	path, err := typemap.ExpectString(arguments, 0, "os.path.dirname")
	if err != nil {
		return nil, err
	}
	return filepath.Dir(path), nil
}

func osPathExists(arguments ...any) (any, error) {
	path, err := typemap.ExpectString(arguments, 0, "os.path.exists")
	if err != nil {
		return nil, err
	}
	_, statErr := os.Stat(path)
	if statErr == nil {
		return true, nil
	}
	if os.IsNotExist(statErr) {
		return false, nil
	}
	return nil, statErr
}

func osPathIsAbs(arguments ...any) (any, error) {
	path, err := typemap.ExpectString(arguments, 0, "os.path.isabs")
	if err != nil {
		return nil, err
	}
	return filepath.IsAbs(path), nil
}

func osPathJoin(arguments ...any) (any, error) {
	segments := make([]string, 0, len(arguments))
	for index := range arguments {
		segment, err := typemap.ExpectString(arguments, index, "os.path.join")
		if err != nil {
			return nil, err
		}
		segments = append(segments, segment)
	}
	return filepath.Join(segments...), nil
}

func registerSubprocess(interpreter runtime.Interpreter) error {
	return interpreter.RegisterModule(runtime.Module{
		Name:          "subprocess",
		Documentation: "Tier 1 subprocess stdlib shadow backed by core process helpers",
		Functions: map[string]runtime.Function{
			"check_output": subprocessCheckOutput,
			"getoutput":    subprocessGetOutput,
			"run":          subprocessRun,
		},
	})
}

func subprocessCheckOutput(arguments ...any) (any, error) {
	result, err := runCommand("subprocess.check_output", arguments, true)
	if err != nil {
		return nil, err
	}
	return result["stdout"], nil
}

func subprocessGetOutput(arguments ...any) (any, error) {
	command, err := typemap.ExpectString(arguments, 0, "subprocess.getoutput")
	if err != nil {
		return nil, err
	}
	completed := exec.Command("sh", "-c", command)
	output, err := completed.CombinedOutput()
	if err != nil {
		return string(output), nil
	}
	return string(output), nil
}

func subprocessRun(arguments ...any) (any, error) {
	return runCommand("subprocess.run", arguments, false)
}

func runCommand(functionName string, arguments []any, check bool) (map[string]any, error) {
	commandLine, err := commandLineArguments(arguments, functionName)
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(commandLine[0], commandLine[1:]...)
	output, err := cmd.Output()
	stderr := ""
	exitCode := 0
	if err != nil {
		exitCode = -1
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
			stderr = string(exitErr.Stderr)
		}
		if check {
			if stderr != "" {
				return nil, fmt.Errorf("%s exited with status %d: %s", functionName, exitCode, firstLine(stderr))
			}
			return nil, fmt.Errorf("%s exited with status %d: %w", functionName, exitCode, err)
		}
	}
	return map[string]any{
		"args":        commandLine,
		"stdout":      string(output),
		"stderr":      stderr,
		"returncode":  exitCode,
		"timed_out":   false,
		"ok":          exitCode == 0,
		"core_shadow": true,
	}, nil
}

func commandLineArguments(arguments []any, functionName string) ([]string, error) {
	if len(arguments) == 0 {
		return nil, fmt.Errorf("%s expected command arguments", functionName)
	}
	switch typed := arguments[0].(type) {
	case []any:
		values := make([]string, 0, len(typed))
		for index, value := range typed {
			text, ok := value.(string)
			if !ok {
				return nil, fmt.Errorf("%s expected command list item %d to be string, got %T", functionName, index, value)
			}
			values = append(values, text)
		}
		if len(values) == 0 {
			return nil, fmt.Errorf("%s expected non-empty command list", functionName)
		}
		return values, nil
	case []string:
		if len(typed) == 0 {
			return nil, fmt.Errorf("%s expected non-empty command list", functionName)
		}
		return append([]string(nil), typed...), nil
	case string:
		return append([]string{typed}, stringArguments(arguments[1:])...), nil
	default:
		return nil, fmt.Errorf("%s expected command string or list, got %T", functionName, arguments[0])
	}
}

func stringArguments(arguments []any) []string {
	values := make([]string, 0, len(arguments))
	for _, argument := range arguments {
		values = append(values, fmt.Sprint(argument))
	}
	return values
}

func registerLogging(interpreter runtime.Interpreter) error {
	return interpreter.RegisterModule(runtime.Module{
		Name:          "logging",
		Documentation: "Tier 1 logging stdlib shadow backed by core log helpers",
		Functions: map[string]runtime.Function{
			"basicConfig": loggingBasicConfig,
			"debug":       loggingDebug,
			"info":        loggingInfo,
			"warning":     loggingWarning,
			"error":       loggingError,
		},
	})
}

func loggingBasicConfig(arguments ...any) (any, error) {
	return nil, nil
}

func loggingDebug(arguments ...any) (any, error) {
	return logWith(core.Debug, "logging.debug", arguments...)
}

func loggingInfo(arguments ...any) (any, error) {
	return logWith(core.Info, "logging.info", arguments...)
}

func loggingWarning(arguments ...any) (any, error) {
	return logWith(core.Warn, "logging.warning", arguments...)
}

func loggingError(arguments ...any) (any, error) {
	return logWith(core.Error, "logging.error", arguments...)
}

func logWith(fn func(string, ...any), functionName string, arguments ...any) (any, error) {
	message, err := typemap.ExpectString(arguments, 0, functionName)
	if err != nil {
		return nil, err
	}
	fn(message, arguments[1:]...)
	return nil, nil
}

func registerHashlib(interpreter runtime.Interpreter) error {
	return interpreter.RegisterModule(runtime.Module{
		Name:          "hashlib",
		Documentation: "Tier 1 hashlib stdlib shadow backed by core crypto helpers",
		Functions: map[string]runtime.Function{
			"_hexdigest": hashlibHexDigest,
			"sha1":       hashlibSHA1,
			"sha256":     hashlibSHA256,
		},
	})
}

func hashlibSHA1(arguments ...any) (any, error) {
	value, err := typemap.ExpectBytes(arguments, 0, "hashlib.sha1")
	if err != nil {
		return nil, err
	}
	sum := sha1.Sum(value)
	return hexDigest{value: hex.EncodeToString(sum[:])}, nil
}

func hashlibSHA256(arguments ...any) (any, error) {
	value, err := typemap.ExpectBytes(arguments, 0, "hashlib.sha256")
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(value)
	return hexDigest{value: hex.EncodeToString(sum[:])}, nil
}

type hexDigest struct {
	value string
}

func (digest hexDigest) ResolveAttribute(name string) (any, bool) {
	switch name {
	case "hexdigest":
		return runtime.BoundMethod{ModuleName: "hashlib", FunctionName: "_hexdigest", Arguments: []any{digest}}, true
	default:
		return nil, false
	}
}

func hashlibHexDigest(arguments ...any) (any, error) {
	if len(arguments) == 0 {
		return nil, fmt.Errorf("hashlib._hexdigest expected digest handle")
	}
	digest, ok := arguments[0].(hexDigest)
	if !ok {
		return nil, fmt.Errorf("hashlib._hexdigest expected digest handle, got %T", arguments[0])
	}
	return digest.value, nil
}

func registerBase64(interpreter runtime.Interpreter) error {
	return interpreter.RegisterModule(runtime.Module{
		Name:          "base64",
		Documentation: "Tier 1 base64 stdlib shadow backed by core crypto helpers",
		Functions: map[string]runtime.Function{
			"b64decode": base64Decode,
			"b64encode": base64Encode,
		},
	})
}

func base64Encode(arguments ...any) (any, error) {
	value, err := typemap.ExpectBytes(arguments, 0, "base64.b64encode")
	if err != nil {
		return nil, err
	}
	return base64.StdEncoding.EncodeToString(value), nil
}

func base64Decode(arguments ...any) (any, error) {
	value, err := typemap.ExpectString(arguments, 0, "base64.b64decode")
	if err != nil {
		return nil, err
	}
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return nil, err
	}
	return decoded, nil
}

func registerSocket(interpreter runtime.Interpreter) error {
	return interpreter.RegisterModule(runtime.Module{
		Name:          "socket",
		Documentation: "Tier 1 socket DNS stdlib shadow backed by core DNS helpers",
		Functions: map[string]runtime.Function{
			"gethostbyname":    socketGetHostByName,
			"gethostbyname_ex": socketGetHostByNameEx,
			"getservbyname":    socketGetServByName,
		},
	})
}

func socketGetHostByName(arguments ...any) (any, error) {
	host, err := typemap.ExpectString(arguments, 0, "socket.gethostbyname")
	if err != nil {
		return nil, err
	}
	addresses, err := net.LookupHost(host)
	if err != nil {
		return nil, err
	}
	if len(addresses) == 0 {
		return nil, fmt.Errorf("socket.gethostbyname: no addresses for %s", host)
	}
	return addresses[0], nil
}

func socketGetHostByNameEx(arguments ...any) (any, error) {
	host, err := typemap.ExpectString(arguments, 0, "socket.gethostbyname_ex")
	if err != nil {
		return nil, err
	}
	addresses, err := net.LookupHost(host)
	if err != nil {
		return nil, err
	}
	slices.Sort(addresses)
	return []any{host, []string{}, addresses}, nil
}

func socketGetServByName(arguments ...any) (any, error) {
	service, err := typemap.ExpectString(arguments, 0, "socket.getservbyname")
	if err != nil {
		return nil, err
	}
	network := "tcp"
	if len(arguments) > 1 {
		network, err = typemap.ExpectString(arguments, 1, "socket.getservbyname")
		if err != nil {
			return nil, err
		}
	}
	return net.LookupPort(strings.ToLower(network), service)
}

func firstLine(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if index := strings.IndexByte(value, '\n'); index != -1 {
		return strings.TrimSpace(value[:index])
	}
	return value
}
