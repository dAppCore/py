package mathbinding

import (
	"fmt"
	stdmath "math"
	"sort"
	"strings"

	"dappco.re/go/py/runtime"
)

// Register exposes math helpers backed by pure Go algorithms.
//
//	mathbinding.Register(interpreter)
func Register(interpreter *runtime.Interpreter) error {
	for _, module := range []runtime.Module{
		{
			Name:          "core.math",
			Documentation: "Statistics, sorting, and scaling helpers for CorePy",
			Functions: map[string]runtime.Function{
				"mean":          mean,
				"median":        median,
				"variance":      variance,
				"stdev":         stdev,
				"sort":          sortValues,
				"binary_search": binarySearch,
				"epsilon_equal": epsilonEqual,
				"normalize":     normalize,
				"rescale":       rescale,
			},
		},
		{
			Name:          "core.math.kdtree",
			Documentation: "KDTree-style nearest-neighbour helpers for CorePy",
			Functions: map[string]runtime.Function{
				"build":   buildKDTree,
				"nearest": nearestKDTree,
			},
		},
		{
			Name:          "core.math.knn",
			Documentation: "KNN helpers for CorePy",
			Functions: map[string]runtime.Function{
				"search": searchKNN,
			},
		},
	} {
		if err := interpreter.RegisterModule(module); err != nil {
			return err
		}
	}
	return nil
}

type kdTreeHandle struct {
	points [][]float64
	metric string
}

// ResolveAttribute exposes the RFC KDTree object surface inside the bootstrap runtime.
//
//	method, ok := tree.ResolveAttribute("nearest")
func (tree *kdTreeHandle) ResolveAttribute(name string) (any, bool) {
	switch name {
	case "nearest":
		return runtime.BoundMethod{
			ModuleName:   "core.math.kdtree",
			FunctionName: "nearest",
			Arguments:    []any{tree},
		}, true
	case "metric":
		return tree.metric, true
	case "points":
		points := make([][]float64, 0, len(tree.points))
		for _, point := range tree.points {
			points = append(points, append([]float64(nil), point...))
		}
		return points, true
	default:
		return nil, false
	}
}

type neighbor struct {
	Index    int
	Distance float64
	Point    []float64
}

func mean(arguments ...any) (any, error) {
	values, err := expectNumericSlice(arguments, 0, "core.math.mean")
	if err != nil {
		return nil, err
	}
	return average(values), nil
}

func median(arguments ...any) (any, error) {
	values, err := expectNumericSlice(arguments, 0, "core.math.median")
	if err != nil {
		return nil, err
	}
	return medianValue(values), nil
}

func variance(arguments ...any) (any, error) {
	values, err := expectNumericSlice(arguments, 0, "core.math.variance")
	if err != nil {
		return nil, err
	}
	return varianceValue(values), nil
}

func stdev(arguments ...any) (any, error) {
	values, err := expectNumericSlice(arguments, 0, "core.math.stdev")
	if err != nil {
		return nil, err
	}
	return stdmath.Sqrt(varianceValue(values)), nil
}

func sortValues(arguments ...any) (any, error) {
	values, err := expectSortableSlice(arguments, 0, "core.math.sort")
	if err != nil {
		return nil, err
	}

	sorted := append([]any(nil), values...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return compareSortable(sorted[i], sorted[j]) < 0
	})
	return sorted, nil
}

func binarySearch(arguments ...any) (any, error) {
	values, err := expectSortableSlice(arguments, 0, "core.math.binary_search")
	if err != nil {
		return nil, err
	}
	if len(arguments) < 2 {
		return nil, fmt.Errorf("core.math.binary_search expected argument 1")
	}
	target := arguments[1]

	low := 0
	high := len(values) - 1
	for low <= high {
		mid := (low + high) / 2
		comparison := compareSortable(values[mid], target)
		switch {
		case comparison == 0:
			return mid, nil
		case comparison < 0:
			low = mid + 1
		default:
			high = mid - 1
		}
	}
	return -1, nil
}

func epsilonEqual(arguments ...any) (any, error) {
	left, err := expectFloat(arguments, 0, "core.math.epsilon_equal")
	if err != nil {
		return nil, err
	}
	right, err := expectFloat(arguments, 1, "core.math.epsilon_equal")
	if err != nil {
		return nil, err
	}

	epsilon := 1e-9
	if len(arguments) > 2 {
		epsilon, err = expectFloat(arguments, 2, "core.math.epsilon_equal")
		if err != nil {
			return nil, err
		}
	}
	return stdmath.Abs(left-right) <= epsilon, nil
}

func normalize(arguments ...any) (any, error) {
	if len(arguments) == 0 {
		return nil, fmt.Errorf("core.math.normalize expected argument 0")
	}
	values, err := numericSliceFromValue(arguments[0], "core.math.normalize")
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return []float64{}, nil
	}

	minimum, maximum := minMax(values)
	if minimum == maximum {
		return make([]float64, len(values)), nil
	}

	result := make([]float64, 0, len(values))
	scale := maximum - minimum
	for _, value := range values {
		result = append(result, (value-minimum)/scale)
	}
	return result, nil
}

func rescale(arguments ...any) (any, error) {
	if len(arguments) == 0 {
		return nil, fmt.Errorf("core.math.rescale expected argument 0")
	}
	values, err := numericSliceFromValue(arguments[0], "core.math.rescale")
	if err != nil {
		return nil, err
	}
	newMinimum, err := expectFloat(arguments, 1, "core.math.rescale")
	if err != nil {
		return nil, err
	}
	newMaximum, err := expectFloat(arguments, 2, "core.math.rescale")
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return []float64{}, nil
	}

	minimum, maximum := minMax(values)
	if minimum == maximum {
		result := make([]float64, len(values))
		for index := range result {
			result[index] = newMinimum
		}
		return result, nil
	}

	result := make([]float64, 0, len(values))
	inputScale := maximum - minimum
	outputScale := newMaximum - newMinimum
	for _, value := range values {
		normalized := (value - minimum) / inputScale
		result = append(result, newMinimum+(normalized*outputScale))
	}
	return result, nil
}

func buildKDTree(arguments ...any) (any, error) {
	positional, keywordArguments := runtime.SplitKeywordArguments(arguments)
	points, err := expectPointSet(positional, 0, "core.math.kdtree.build")
	if err != nil {
		return nil, err
	}
	if err := validateKeywordArguments("core.math.kdtree.build", keywordArguments, "metric"); err != nil {
		return nil, err
	}
	metric := "euclidean"
	if len(positional) > 1 {
		metric, err = expectMetric(positional[1], "core.math.kdtree.build")
		if err != nil {
			return nil, err
		}
	}
	metric, err = keywordMetric("core.math.kdtree.build", metric, keywordArguments, len(positional) > 1)
	if err != nil {
		return nil, err
	}
	return &kdTreeHandle{
		points: points,
		metric: metric,
	}, nil
}

func nearestKDTree(arguments ...any) (any, error) {
	positional, keywordArguments := runtime.SplitKeywordArguments(arguments)
	if len(positional) == 0 {
		return nil, fmt.Errorf("core.math.kdtree.nearest expected argument 0")
	}
	if err := validateKeywordArguments("core.math.kdtree.nearest", keywordArguments, "k"); err != nil {
		return nil, err
	}

	tree, ok := positional[0].(*kdTreeHandle)
	if !ok {
		return nil, fmt.Errorf("core.math.kdtree.nearest expected KDTree handle, got %T", positional[0])
	}
	query, err := expectPoint(positional, 1, "core.math.kdtree.nearest")
	if err != nil {
		return nil, err
	}
	k := 1
	if len(positional) > 2 {
		k, err = expectPositiveInt(positional, 2, "core.math.kdtree.nearest")
		if err != nil {
			return nil, err
		}
	}
	k, err = keywordPositiveInt("core.math.kdtree.nearest", "k", k, keywordArguments, len(positional) > 2)
	if err != nil {
		return nil, err
	}

	return searchPoints(tree.points, query, k, tree.metric)
}

func searchKNN(arguments ...any) (any, error) {
	positional, keywordArguments := runtime.SplitKeywordArguments(arguments)
	points, err := expectPointSet(positional, 0, "core.math.knn.search")
	if err != nil {
		return nil, err
	}
	if err := validateKeywordArguments("core.math.knn.search", keywordArguments, "k", "metric"); err != nil {
		return nil, err
	}
	query, err := expectPoint(positional, 1, "core.math.knn.search")
	if err != nil {
		return nil, err
	}
	k := 1
	if len(positional) > 2 {
		k, err = expectPositiveInt(positional, 2, "core.math.knn.search")
		if err != nil {
			return nil, err
		}
	}
	k, err = keywordPositiveInt("core.math.knn.search", "k", k, keywordArguments, len(positional) > 2)
	if err != nil {
		return nil, err
	}

	metric := "euclidean"
	if len(positional) > 3 {
		metric, err = expectMetric(positional[3], "core.math.knn.search")
		if err != nil {
			return nil, err
		}
	}
	metric, err = keywordMetric("core.math.knn.search", metric, keywordArguments, len(positional) > 3)
	if err != nil {
		return nil, err
	}
	return searchPoints(points, query, k, metric)
}

func searchPoints(points [][]float64, query []float64, k int, metric string) ([]map[string]any, error) {
	if k <= 0 {
		return nil, fmt.Errorf("k must be positive")
	}

	neighbors := make([]neighbor, 0, len(points))
	for index, point := range points {
		distance, err := pointDistance(metric, point, query)
		if err != nil {
			return nil, err
		}
		neighbors = append(neighbors, neighbor{
			Index:    index,
			Distance: distance,
			Point:    append([]float64(nil), point...),
		})
	}

	sort.SliceStable(neighbors, func(i, j int) bool {
		if neighbors[i].Distance == neighbors[j].Distance {
			return neighbors[i].Index < neighbors[j].Index
		}
		return neighbors[i].Distance < neighbors[j].Distance
	})
	if k > len(neighbors) {
		k = len(neighbors)
	}

	results := make([]map[string]any, 0, k)
	for _, item := range neighbors[:k] {
		results = append(results, map[string]any{
			"index":    item.Index,
			"distance": item.Distance,
			"point":    append([]float64(nil), item.Point...),
		})
	}
	return results, nil
}

func pointDistance(metric string, left, right []float64) (float64, error) {
	if len(left) != len(right) {
		return 0, fmt.Errorf("point dimension mismatch: %d != %d", len(left), len(right))
	}

	switch metric {
	case "euclidean":
		var total float64
		for index := range left {
			delta := left[index] - right[index]
			total += delta * delta
		}
		return stdmath.Sqrt(total), nil
	case "manhattan":
		var total float64
		for index := range left {
			total += stdmath.Abs(left[index] - right[index])
		}
		return total, nil
	case "chebyshev":
		var maximum float64
		for index := range left {
			delta := stdmath.Abs(left[index] - right[index])
			if delta > maximum {
				maximum = delta
			}
		}
		return maximum, nil
	case "cosine":
		var dotProduct float64
		var leftNorm float64
		var rightNorm float64
		for index := range left {
			dotProduct += left[index] * right[index]
			leftNorm += left[index] * left[index]
			rightNorm += right[index] * right[index]
		}
		if leftNorm == 0 && rightNorm == 0 {
			return 0, nil
		}
		if leftNorm == 0 || rightNorm == 0 {
			return 1, nil
		}
		return 1 - (dotProduct / (stdmath.Sqrt(leftNorm) * stdmath.Sqrt(rightNorm))), nil
	default:
		return 0, fmt.Errorf("unknown metric %q", metric)
	}
}

func average(values []float64) float64 {
	var total float64
	for _, value := range values {
		total += value
	}
	return total / float64(len(values))
}

func medianValue(values []float64) float64 {
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	middle := len(sorted) / 2
	if len(sorted)%2 == 1 {
		return sorted[middle]
	}
	return (sorted[middle-1] + sorted[middle]) / 2
}

func varianceValue(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	meanValue := average(values)
	var total float64
	for _, value := range values {
		delta := value - meanValue
		total += delta * delta
	}
	return total / float64(len(values))
}

func minMax(values []float64) (float64, float64) {
	minimum := values[0]
	maximum := values[0]
	for _, value := range values[1:] {
		if value < minimum {
			minimum = value
		}
		if value > maximum {
			maximum = value
		}
	}
	return minimum, maximum
}

func expectNumericSlice(arguments []any, index int, functionName string) ([]float64, error) {
	if index >= len(arguments) {
		return nil, fmt.Errorf("%s expected argument %d", functionName, index)
	}
	values, err := numericSliceFromValue(arguments[index], functionName)
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, fmt.Errorf("%s expected at least one numeric value", functionName)
	}
	return values, nil
}

func expectPoint(arguments []any, index int, functionName string) ([]float64, error) {
	if index >= len(arguments) {
		return nil, fmt.Errorf("%s expected argument %d", functionName, index)
	}
	values, err := numericSliceFromValue(arguments[index], functionName)
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, fmt.Errorf("%s expected point with at least one dimension", functionName)
	}
	return values, nil
}

func expectPointSet(arguments []any, index int, functionName string) ([][]float64, error) {
	if index >= len(arguments) {
		return nil, fmt.Errorf("%s expected argument %d", functionName, index)
	}
	points, err := pointSetFromValue(arguments[index], functionName)
	if err != nil {
		return nil, err
	}
	if len(points) == 0 {
		return nil, fmt.Errorf("%s expected at least one point", functionName)
	}
	return points, nil
}

func numericSliceFromValue(value any, functionName string) ([]float64, error) {
	switch typed := value.(type) {
	case []float64:
		return append([]float64(nil), typed...), nil
	case []int:
		result := make([]float64, 0, len(typed))
		for _, item := range typed {
			result = append(result, float64(item))
		}
		return result, nil
	case []any:
		result := make([]float64, 0, len(typed))
		for _, item := range typed {
			number, err := floatFromValue(item, functionName)
			if err != nil {
				return nil, err
			}
			result = append(result, number)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("%s expected numeric slice, got %T", functionName, value)
	}
}

func pointSetFromValue(value any, functionName string) ([][]float64, error) {
	switch typed := value.(type) {
	case [][]float64:
		result := make([][]float64, 0, len(typed))
		for _, point := range typed {
			result = append(result, append([]float64(nil), point...))
		}
		return result, nil
	case []any:
		result := make([][]float64, 0, len(typed))
		for _, item := range typed {
			point, err := numericSliceFromValue(item, functionName)
			if err != nil {
				return nil, err
			}
			result = append(result, point)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("%s expected point list, got %T", functionName, value)
	}
}

func expectSortableSlice(arguments []any, index int, functionName string) ([]any, error) {
	if index >= len(arguments) {
		return nil, fmt.Errorf("%s expected argument %d", functionName, index)
	}

	switch typed := arguments[index].(type) {
	case []string:
		result := make([]any, 0, len(typed))
		for _, item := range typed {
			result = append(result, item)
		}
		return result, nil
	case []int:
		result := make([]any, 0, len(typed))
		for _, item := range typed {
			result = append(result, item)
		}
		return result, nil
	case []float64:
		result := make([]any, 0, len(typed))
		for _, item := range typed {
			result = append(result, item)
		}
		return result, nil
	case []any:
		if len(typed) == 0 {
			return []any{}, nil
		}
		firstKind := sortableKind(typed[0])
		if firstKind == "" {
			return nil, fmt.Errorf("%s expected sortable values, got %T", functionName, typed[0])
		}
		result := make([]any, 0, len(typed))
		for _, item := range typed {
			if sortableKind(item) != firstKind {
				return nil, fmt.Errorf("%s expected homogenous sortable values, got %T", functionName, item)
			}
			result = append(result, item)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("%s expected sortable slice, got %T", functionName, arguments[index])
	}
}

func compareSortable(left, right any) int {
	if leftText, ok := left.(string); ok {
		rightText, ok := right.(string)
		if !ok {
			return strings.Compare(fmt.Sprintf("%T", left), fmt.Sprintf("%T", right))
		}
		return strings.Compare(leftText, rightText)
	}

	leftNumber, leftOK := maybeFloat(left)
	rightNumber, rightOK := maybeFloat(right)
	if leftOK && rightOK {
		switch {
		case leftNumber < rightNumber:
			return -1
		case leftNumber > rightNumber:
			return 1
		default:
			return 0
		}
	}

	return strings.Compare(fmt.Sprintf("%v", left), fmt.Sprintf("%v", right))
}

func sortableKind(value any) string {
	if _, ok := value.(string); ok {
		return "string"
	}
	if _, ok := maybeFloat(value); ok {
		return "number"
	}
	return ""
}

func expectFloat(arguments []any, index int, functionName string) (float64, error) {
	if index >= len(arguments) {
		return 0, fmt.Errorf("%s expected argument %d", functionName, index)
	}
	return floatFromValue(arguments[index], functionName)
}

func expectPositiveInt(arguments []any, index int, functionName string) (int, error) {
	if index >= len(arguments) {
		return 0, fmt.Errorf("%s expected argument %d", functionName, index)
	}
	switch typed := arguments[index].(type) {
	case int:
		if typed <= 0 {
			return 0, fmt.Errorf("%s expected positive integer, got %d", functionName, typed)
		}
		return typed, nil
	default:
		return 0, fmt.Errorf("%s expected positive integer, got %T", functionName, arguments[index])
	}
}

func expectMetric(value any, functionName string) (string, error) {
	text, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("%s expected metric string, got %T", functionName, value)
	}
	text = strings.ToLower(text)
	switch text {
	case "euclidean", "manhattan", "chebyshev", "cosine":
		return text, nil
	default:
		return "", fmt.Errorf("%s unknown metric %q", functionName, text)
	}
}

func floatFromValue(value any, functionName string) (float64, error) {
	if number, ok := maybeFloat(value); ok {
		return number, nil
	}
	return 0, fmt.Errorf("%s expected number, got %T", functionName, value)
}

func maybeFloat(value any) (float64, bool) {
	switch typed := value.(type) {
	case int:
		return float64(typed), true
	case float64:
		return typed, true
	default:
		return 0, false
	}
}

func keywordMetric(functionName string, current string, keywordArguments runtime.KeywordArguments, alreadySet bool) (string, error) {
	if len(keywordArguments) == 0 {
		return current, nil
	}

	metricValue, ok := keywordArguments["metric"]
	if !ok {
		return current, nil
	}
	if alreadySet {
		return "", fmt.Errorf("%s received multiple values for metric", functionName)
	}
	return expectMetric(metricValue, functionName)
}

func keywordPositiveInt(functionName, name string, current int, keywordArguments runtime.KeywordArguments, alreadySet bool) (int, error) {
	if len(keywordArguments) == 0 {
		return current, nil
	}

	value, ok := keywordArguments[name]
	if !ok {
		return current, nil
	}
	if alreadySet {
		return 0, fmt.Errorf("%s received multiple values for %s", functionName, name)
	}
	switch typed := value.(type) {
	case int:
		if typed <= 0 {
			return 0, fmt.Errorf("%s expected positive integer, got %d", functionName, typed)
		}
		return typed, nil
	default:
		return 0, fmt.Errorf("%s expected positive integer, got %T", functionName, value)
	}
}

func validateKeywordArguments(functionName string, keywordArguments runtime.KeywordArguments, allowed ...string) error {
	if len(keywordArguments) == 0 {
		return nil
	}

	allowedSet := make(map[string]struct{}, len(allowed))
	for _, name := range allowed {
		allowedSet[name] = struct{}{}
	}

	var unexpected []string
	for name := range keywordArguments {
		if _, ok := allowedSet[name]; ok {
			continue
		}
		unexpected = append(unexpected, name)
	}
	if len(unexpected) == 0 {
		return nil
	}

	sort.Strings(unexpected)
	if len(unexpected) == 1 {
		return fmt.Errorf("%s got unexpected keyword argument %q", functionName, unexpected[0])
	}
	return fmt.Errorf("%s got unexpected keyword arguments %s", functionName, strings.Join(unexpected, ", "))
}
