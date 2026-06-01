package mathbinding

import (
	stdmath "math"

	core "dappco.re/go"
	"dappco.re/go/py/runtime"
)

// newMathInterpreter registers the math modules against a fresh bootstrap
// interpreter and returns the direct caller used to exercise the bindings.
//
//	caller := newMathInterpreter(t)
//	value, err := caller.Call("core.math", "mean", []float64{1, 2, 3})
func newMathInterpreter(t *core.T) runtime.DirectCaller {
	t.Helper()

	interpreter, err := runtime.New(runtime.Options{})
	if err != nil {
		t.Fatalf("create interpreter: %v", err)
	}
	t.Cleanup(func() { _ = interpreter.Close() })

	if err := Register(interpreter); err != nil {
		t.Fatalf("register math modules: %v", err)
	}

	caller, ok := interpreter.(runtime.DirectCaller)
	if !ok {
		t.Fatalf("interpreter does not expose direct calls: %T", interpreter)
	}
	return caller
}

// callFloat invokes a binding and asserts a float64 result.
func callFloat(t *core.T, caller runtime.DirectCaller, function string, arguments ...any) float64 {
	t.Helper()
	value, callErr := caller.Call("core.math", function, arguments...)
	if callErr != nil {
		t.Fatalf("%s: %v", function, callErr)
	}
	number, ok := value.(float64)
	if !ok {
		t.Fatalf("%s: expected float64, got %T", function, value)
	}
	return number
}

// TestMath_Statistics_Good computes mean, median, variance, and stdev.
func TestMath_Statistics_Good(t *core.T) {
	caller := newMathInterpreter(t)
	sample := []float64{2, 4, 4, 4, 5, 5, 7, 9}

	if got := callFloat(t, caller, "mean", sample); got != 5 {
		t.Fatalf("mean: %v", got)
	}
	if got := callFloat(t, caller, "median", sample); got != 4.5 {
		t.Fatalf("median: %v", got)
	}
	if got := callFloat(t, caller, "variance", sample); got != 4 {
		t.Fatalf("variance: %v", got)
	}
	if got := callFloat(t, caller, "stdev", sample); got != 2 {
		t.Fatalf("stdev: %v", got)
	}
}

// TestMath_Sort_Good sorts a heterogeneous numeric slice.
func TestMath_Sort_Good(t *core.T) {
	caller := newMathInterpreter(t)

	value, callErr := caller.Call("core.math", "sort", []int{3, 1, 2})
	if callErr != nil {
		t.Fatalf("sort: %v", callErr)
	}
	sorted, ok := value.([]any)
	if !ok || len(sorted) != 3 {
		t.Fatalf("sort: unexpected %#v", value)
	}
	if sorted[0] != 1 || sorted[2] != 3 {
		t.Fatalf("sort: not ordered %#v", sorted)
	}
}

// TestMath_BinarySearch_Good finds an element in a sorted slice.
func TestMath_BinarySearch_Good(t *core.T) {
	caller := newMathInterpreter(t)

	value, callErr := caller.Call("core.math", "binary_search", []int{1, 2, 3, 4, 5}, 4)
	if callErr != nil {
		t.Fatalf("binary_search: %v", callErr)
	}
	if value != 3 {
		t.Fatalf("binary_search: unexpected index %#v", value)
	}
}

// TestMath_EpsilonEqual_Good compares two close floats.
func TestMath_EpsilonEqual_Good(t *core.T) {
	caller := newMathInterpreter(t)

	value, callErr := caller.Call("core.math", "epsilon_equal", 0.1+0.2, 0.3)
	if callErr != nil {
		t.Fatalf("epsilon_equal: %v", callErr)
	}
	if value != true {
		t.Fatalf("epsilon_equal: expected near-equality, got %#v", value)
	}
}

// TestMath_Normalize_Good scales a slice to unit length.
func TestMath_Normalize_Good(t *core.T) {
	caller := newMathInterpreter(t)

	value, callErr := caller.Call("core.math", "normalize", []float64{3, 4})
	if callErr != nil {
		t.Fatalf("normalize: %v", callErr)
	}
	normalised, ok := value.([]float64)
	if !ok || len(normalised) != 2 {
		t.Fatalf("normalize: unexpected %#v", value)
	}
	magnitude := stdmath.Sqrt(normalised[0]*normalised[0] + normalised[1]*normalised[1])
	if stdmath.Abs(magnitude-1) > 1e-9 {
		t.Fatalf("normalize: expected unit length, got %v", magnitude)
	}
}

// TestMath_Rescale_Good rescales a slice into a new range.
func TestMath_Rescale_Good(t *core.T) {
	caller := newMathInterpreter(t)

	value, callErr := caller.Call("core.math", "rescale", []float64{0, 5, 10}, 0.0, 1.0)
	if callErr != nil {
		t.Fatalf("rescale: %v", callErr)
	}
	rescaled, ok := value.([]float64)
	if !ok || len(rescaled) != 3 {
		t.Fatalf("rescale: unexpected %#v", value)
	}
	if rescaled[0] != 0 || rescaled[2] != 1 {
		t.Fatalf("rescale: unexpected bounds %#v", rescaled)
	}
}

// TestMath_MovingAverage_Good computes a windowed moving average.
func TestMath_MovingAverage_Good(t *core.T) {
	caller := newMathInterpreter(t)

	value, callErr := caller.Call("core.math", "moving_average", []float64{1, 2, 3, 4}, 2)
	if callErr != nil {
		t.Fatalf("moving_average: %v", callErr)
	}
	result, ok := value.([]float64)
	if !ok || len(result) != 4 {
		t.Fatalf("moving_average: unexpected %#v", value)
	}
	if result[3] != 3.5 {
		t.Fatalf("moving_average: unexpected tail %v", result[3])
	}
}

// TestMath_Difference_Good computes lagged differences.
func TestMath_Difference_Good(t *core.T) {
	caller := newMathInterpreter(t)

	value, callErr := caller.Call("core.math", "difference", []float64{1, 3, 6, 10}, 1)
	if callErr != nil {
		t.Fatalf("difference: %v", callErr)
	}
	result, ok := value.([]float64)
	if !ok || len(result) != 3 {
		t.Fatalf("difference: unexpected %#v", value)
	}
	if result[0] != 2 || result[2] != 4 {
		t.Fatalf("difference: unexpected values %#v", result)
	}
}

// TestMath_KDTree_Good builds a KDTree and finds the nearest point.
func TestMath_KDTree_Good(t *core.T) {
	caller := newMathInterpreter(t)

	tree, callErr := caller.Call("core.math.kdtree", "build", [][]float64{{0, 0}, {1, 1}, {5, 5}})
	if callErr != nil {
		t.Fatalf("kdtree build: %v", callErr)
	}

	value, callErr := caller.Call("core.math.kdtree", "nearest", tree, []float64{0.9, 0.9}, 1)
	if callErr != nil {
		t.Fatalf("kdtree nearest: %v", callErr)
	}
	results, ok := value.([]map[string]any)
	if !ok || len(results) != 1 {
		t.Fatalf("kdtree nearest: unexpected %#v", value)
	}
	if results[0]["index"] != 1 {
		t.Fatalf("kdtree nearest: expected index 1, got %#v", results[0]["index"])
	}
}

// TestMath_KNN_Good searches for nearest neighbours.
func TestMath_KNN_Good(t *core.T) {
	caller := newMathInterpreter(t)

	value, callErr := caller.Call("core.math.knn", "search", [][]float64{{0, 0}, {1, 1}, {5, 5}}, []float64{4.8, 4.8}, 2)
	if callErr != nil {
		t.Fatalf("knn search: %v", callErr)
	}
	results, ok := value.([]map[string]any)
	if !ok || len(results) != 2 {
		t.Fatalf("knn search: unexpected %#v", value)
	}
	if results[0]["index"] != 2 {
		t.Fatalf("knn search: expected nearest index 2, got %#v", results[0]["index"])
	}
}

// TestMath_Mean_Bad reports an empty numeric slice.
func TestMath_Mean_Bad(t *core.T) {
	caller := newMathInterpreter(t)

	if _, callErr := caller.Call("core.math", "mean", []float64{}); callErr == nil {
		t.Fatal("expected error for empty slice")
	}
	if _, callErr := caller.Call("core.math", "mean"); callErr == nil {
		t.Fatal("expected error for missing argument")
	}
}

// TestMath_Mean_Ugly rejects a non-numeric-slice argument.
func TestMath_Mean_Ugly(t *core.T) {
	caller := newMathInterpreter(t)

	if _, callErr := caller.Call("core.math", "mean", "not-a-slice"); callErr == nil {
		t.Fatal("expected error for non-slice argument")
	}
}
