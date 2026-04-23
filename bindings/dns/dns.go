package dns

import (
	"net"
	"slices"

	"dappco.re/go/py/bindings/typemap"
	"dappco.re/go/py/runtime"
)

// Register exposes DNS resolution helpers.
//
//	dns.Register(interpreter)
func Register(interpreter *runtime.Interpreter) error {
	return interpreter.RegisterModule(runtime.Module{
		Name:          "core.dns",
		Documentation: "DNS helpers for CorePy",
		Functions: map[string]runtime.Function{
			"lookup_host":    lookupHost,
			"lookup_ip":      lookupIP,
			"reverse_lookup": reverseLookup,
			"lookup_port":    lookupPort,
		},
	})
}

func lookupHost(arguments ...any) (any, error) {
	name, err := typemap.ExpectString(arguments, 0, "core.dns.lookup_host")
	if err != nil {
		return nil, err
	}
	values, err := net.LookupHost(name)
	if err != nil {
		return nil, err
	}
	return uniqueSorted(values), nil
}

func lookupIP(arguments ...any) (any, error) {
	name, err := typemap.ExpectString(arguments, 0, "core.dns.lookup_ip")
	if err != nil {
		return nil, err
	}
	values, err := net.LookupIP(name)
	if err != nil {
		return nil, err
	}
	addresses := make([]string, 0, len(values))
	for _, value := range values {
		addresses = append(addresses, value.String())
	}
	return uniqueSorted(addresses), nil
}

func reverseLookup(arguments ...any) (any, error) {
	address, err := typemap.ExpectString(arguments, 0, "core.dns.reverse_lookup")
	if err != nil {
		return nil, err
	}
	values, err := net.LookupAddr(address)
	if err != nil {
		return nil, err
	}
	return uniqueSorted(values), nil
}

func lookupPort(arguments ...any) (any, error) {
	network, err := typemap.ExpectString(arguments, 0, "core.dns.lookup_port")
	if err != nil {
		return nil, err
	}
	service, err := typemap.ExpectString(arguments, 1, "core.dns.lookup_port")
	if err != nil {
		return nil, err
	}
	return net.LookupPort(network, service)
}

func uniqueSorted(values []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	slices.Sort(result)
	return result
}
