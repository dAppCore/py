package runtime_test

import (
	"reflect"
	"strings"
	"testing"
)

func TestRFCStubModules_Available_Good(t *testing.T) {
	interpreter := newTestInterpreter(t)

	output, err := interpreter.Run(`
from core import agent, api, container, mcp, store, ws
print(agent.available())
print(api.available())
print(container.available())
print(mcp.available())
print(store.available())
print(ws.available())
`)
	if err != nil {
		t.Fatalf("run RFC stub module imports: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	expected := []string{"False", "False", "False", "False", "False", "False"}
	if !reflect.DeepEqual(lines, expected) {
		t.Fatalf("unexpected availability output %#v", lines)
	}
}
