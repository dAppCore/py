package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"dappco.re/go/py/bindings/register"
	corepyruntime "dappco.re/go/py/runtime"
)

const version = "0.2.0"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(arguments []string) error {
	if len(arguments) == 0 {
		return usageError()
	}

	switch arguments[0] {
	case "run":
		return runScript(arguments[1:])
	case "modules":
		return listModules(arguments[1:])
	case "version":
		return printVersion(arguments[1:])
	default:
		return usageError()
	}
}

func runScript(arguments []string) error {
	flags := flag.NewFlagSet("corepy run", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	expression := flags.String("e", "", "execute source string")
	if err := flags.Parse(arguments); err != nil {
		return err
	}

	var source string
	switch {
	case strings.TrimSpace(*expression) != "":
		source = *expression
	case flags.NArg() == 1:
		content, err := os.ReadFile(flags.Arg(0))
		if err != nil {
			return fmt.Errorf("corepy run: read %s: %w", flags.Arg(0), err)
		}
		source = string(content)
	default:
		return fmt.Errorf("usage: corepy run [-e source] [file.py]")
	}

	interpreter, err := newInterpreter()
	if err != nil {
		return err
	}
	defer interpreter.Close()

	output, err := interpreter.Run(source)
	if err != nil {
		return err
	}
	_, err = os.Stdout.WriteString(output)
	return err
}

func listModules(arguments []string) error {
	if len(arguments) != 0 {
		return fmt.Errorf("usage: corepy modules")
	}

	interpreter, err := newInterpreter()
	if err != nil {
		return err
	}
	defer interpreter.Close()

	lister, ok := interpreter.(corepyruntime.ModuleLister)
	if !ok {
		return fmt.Errorf("corepy modules: backend does not expose module listing")
	}
	for _, name := range lister.Modules() {
		fmt.Println(name)
	}
	return nil
}

func printVersion(arguments []string) error {
	if len(arguments) != 0 {
		return fmt.Errorf("usage: corepy version")
	}
	fmt.Printf("corepy %s backend=%s\n", version, corepyruntime.BackendBootstrap)
	return nil
}

func newInterpreter() (corepyruntime.Interpreter, error) {
	interpreter, err := corepyruntime.New(corepyruntime.Options{Backend: corepyruntime.BackendBootstrap})
	if err != nil {
		return nil, err
	}
	if err := register.DefaultModules(interpreter); err != nil {
		_ = interpreter.Close()
		return nil, err
	}
	return interpreter, nil
}

func usageError() error {
	return fmt.Errorf("usage: corepy run [-e source] [file.py] | corepy modules | corepy version")
}
