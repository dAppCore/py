package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"dappco.re/go/py/bindings/register"
	corepyruntime "dappco.re/go/py/runtime"
	"dappco.re/go/py/runtime/tier2"
)

const version = "0.4.0"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(arguments []string) error {
	return runWithIO(arguments, os.Stdout, os.Stderr)
}

func runWithIO(arguments []string, stdout io.Writer, stderr io.Writer) error {
	if len(arguments) == 0 {
		return usageError()
	}

	switch arguments[0] {
	case "run":
		return runScript(arguments[1:], stdout, stderr)
	case "modules":
		return listModules(arguments[1:], stdout, stderr)
	case "tier2":
		return runTier2(arguments[1:], stdout, stderr)
	case "version":
		return printVersion(arguments[1:], stdout)
	default:
		return usageError()
	}
}

func runScript(arguments []string, stdout io.Writer, stderr io.Writer) error {
	flags := flag.NewFlagSet("corepy run", flag.ContinueOnError)
	flags.SetOutput(stderr)
	backend := flags.String("backend", corepyruntime.BackendBootstrap, "runtime backend: bootstrap or gpython")
	expression := flags.String("e", "", "execute source string")
	tier2Fallback := flags.Bool("tier2-fallback", false, "retry unsupported Tier 1 imports with Tier 2 CPython")
	tier2Python := flags.String("python", "", "host Python executable for -tier2-fallback")
	tier2Timeout := flags.Duration("timeout", 0, "Tier 2 fallback timeout, for example 10s or 250ms")
	if err := flags.Parse(arguments); err != nil {
		return err
	}

	var (
		source   string
		filename string
	)
	switch {
	case strings.TrimSpace(*expression) != "":
		source = *expression
	case flags.NArg() == 1:
		filename = flags.Arg(0)
		content, err := os.ReadFile(flags.Arg(0))
		if err != nil {
			return fmt.Errorf("corepy run: read %s: %w", flags.Arg(0), err)
		}
		source = string(content)
	default:
		return fmt.Errorf("usage: corepy run [-backend bootstrap|gpython] [-tier2-fallback] [-python python3] [-timeout 10s] [-e source] [file.py]")
	}

	interpreter, err := newInterpreter(*backend)
	if err != nil {
		return err
	}
	defer interpreter.Close()

	output, err := interpreter.Run(source)
	if err != nil {
		if !*tier2Fallback || !corepyruntime.IsTier2FallbackCandidate(err) {
			return err
		}
		return runTier2Fallback(filename, source, *tier2Python, *tier2Timeout, stdout, stderr)
	}
	_, err = io.WriteString(stdout, output)
	return err
}

func runTier2Fallback(filename string, source string, python string, timeout time.Duration, stdout io.Writer, stderr io.Writer) error {
	runner := tier2.NewRunner(tier2.Options{
		Python:     python,
		PythonPath: localPythonPath(),
		Timeout:    timeout,
		Stdout:     stdout,
		Stderr:     stderr,
	})
	if filename != "" {
		_, err := runner.RunFile(context.Background(), filename)
		return err
	}
	_, err := runner.RunSource(context.Background(), source)
	return err
}

func listModules(arguments []string, stdout io.Writer, stderr io.Writer) error {
	flags := flag.NewFlagSet("corepy modules", flag.ContinueOnError)
	flags.SetOutput(stderr)
	backend := flags.String("backend", corepyruntime.BackendBootstrap, "runtime backend: bootstrap or gpython")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("usage: corepy modules [-backend bootstrap|gpython]")
	}

	interpreter, err := newInterpreter(*backend)
	if err != nil {
		return err
	}
	defer interpreter.Close()

	lister, ok := interpreter.(corepyruntime.ModuleLister)
	if !ok {
		return fmt.Errorf("corepy modules: backend does not expose module listing")
	}
	for _, name := range lister.Modules() {
		fmt.Fprintln(stdout, name)
	}
	return nil
}

func runTier2(arguments []string, stdout io.Writer, stderr io.Writer) error {
	if len(arguments) == 0 {
		return tier2UsageError()
	}

	switch arguments[0] {
	case "run":
		return runTier2Script(arguments[1:], stdout, stderr)
	case "which":
		return printTier2Python(arguments[1:], stdout, stderr)
	default:
		return tier2UsageError()
	}
}

func runTier2Script(arguments []string, stdout io.Writer, stderr io.Writer) error {
	flags := flag.NewFlagSet("corepy tier2 run", flag.ContinueOnError)
	flags.SetOutput(stderr)
	python := flags.String("python", "", "host Python executable")
	timeout := flags.Duration("timeout", 0, "timeout, for example 10s or 250ms")
	expression := flags.String("e", "", "execute source string")
	if err := flags.Parse(arguments); err != nil {
		return err
	}

	var (
		source   string
		filename string
	)
	switch {
	case strings.TrimSpace(*expression) != "":
		source = *expression
	case flags.NArg() == 1:
		filename = flags.Arg(0)
	default:
		return fmt.Errorf("usage: corepy tier2 run [-python python3] [-timeout 10s] [-e source] [file.py]")
	}

	runner := tier2.NewRunner(tier2.Options{
		Python:     *python,
		PythonPath: localPythonPath(),
		Timeout:    *timeout,
		Stdout:     stdout,
		Stderr:     stderr,
	})
	if source != "" {
		_, err := runner.RunSource(context.Background(), source)
		return err
	}
	_, err := runner.RunFile(context.Background(), filename)
	return err
}

func printTier2Python(arguments []string, stdout io.Writer, stderr io.Writer) error {
	flags := flag.NewFlagSet("corepy tier2 which", flag.ContinueOnError)
	flags.SetOutput(stderr)
	python := flags.String("python", "", "host Python executable")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("usage: corepy tier2 which [-python python3]")
	}

	path, err := tier2.ResolvePython(*python)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(stdout, path)
	return err
}

func printVersion(arguments []string, stdout io.Writer) error {
	if len(arguments) != 0 {
		return fmt.Errorf("usage: corepy version")
	}
	_, err := fmt.Fprintf(stdout, "corepy %s backend=%s tier2=cpython-subprocess\n", version, corepyruntime.BackendBootstrap)
	return err
}

func newInterpreter(backend string) (corepyruntime.Interpreter, error) {
	interpreter, err := corepyruntime.New(corepyruntime.Options{Backend: backend})
	if err != nil {
		return nil, err
	}
	if err := register.DefaultModules(interpreter); err != nil {
		_ = interpreter.Close()
		return nil, err
	}
	return interpreter, nil
}

func localPythonPath() []string {
	if path, ok := tier2.LocalPythonPath(""); ok {
		return []string{path}
	}
	return nil
}

func usageError() error {
	return fmt.Errorf("usage: corepy run [-backend bootstrap|gpython] [-tier2-fallback] [-python python3] [-timeout 10s] [-e source] [file.py] | corepy modules [-backend bootstrap|gpython] | corepy tier2 run|which | corepy version")
}

func tier2UsageError() error {
	return fmt.Errorf("usage: corepy tier2 run [-python python3] [-timeout %s] [-e source] [file.py] | corepy tier2 which [-python python3]", (10 * time.Second).String())
}
