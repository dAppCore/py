package cache

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"dappco.re/go/py/bindings/typemap"
	"dappco.re/go/py/runtime"
)

const defaultTTL = time.Hour

type handle struct {
	baseDir string
	ttl     time.Duration
}

type entry struct {
	Data      any       `json:"data"`
	CachedAt  time.Time `json:"cached_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Register exposes file-backed cache helpers.
//
//	cache.Register(interpreter)
func Register(interpreter *runtime.Interpreter) error {
	return interpreter.RegisterModule(runtime.Module{
		Name:          "core.cache",
		Documentation: "JSON cache helpers for CorePy",
		Functions: map[string]runtime.Function{
			"new":          newCache,
			"path":         pathForKey,
			"set":          setValue,
			"set_with_ttl": setWithTTL,
			"get":          getValue,
			"has":          hasValue,
			"delete":       deleteValue,
			"clear":        clearValues,
			"keys":         keys,
		},
	})
}

func newCache(arguments ...any) (any, error) {
	baseDir := ""
	ttl := defaultTTL

	if len(arguments) > 0 && arguments[0] != nil {
		value, err := typemap.ExpectString(arguments, 0, "core.cache.new")
		if err != nil {
			return nil, err
		}
		baseDir = value
	}
	if len(arguments) > 1 {
		ttlSeconds, err := expectSeconds(arguments[1], "core.cache.new")
		if err != nil {
			return nil, err
		}
		ttl = ttlFromSeconds(ttlSeconds)
	}

	if baseDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("core.cache.new failed to resolve current directory: %w", err)
		}
		baseDir = filepath.Join(cwd, ".core", "cache")
	}
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("core.cache.new failed to create cache directory: %w", err)
	}
	return &handle{baseDir: baseDir, ttl: ttl}, nil
}

func pathForKey(arguments ...any) (any, error) {
	cacheHandle, err := expectHandle(arguments, 0, "core.cache.path")
	if err != nil {
		return nil, err
	}
	key, err := typemap.ExpectString(arguments, 1, "core.cache.path")
	if err != nil {
		return nil, err
	}
	return cacheHandle.path(key)
}

func setValue(arguments ...any) (any, error) {
	cacheHandle, err := expectHandle(arguments, 0, "core.cache.set")
	if err != nil {
		return nil, err
	}
	key, err := typemap.ExpectString(arguments, 1, "core.cache.set")
	if err != nil {
		return nil, err
	}
	if len(arguments) < 3 {
		return nil, fmt.Errorf("core.cache.set expected argument 2")
	}
	path, err := cacheHandle.set(key, arguments[2], cacheHandle.ttl)
	if err != nil {
		return nil, err
	}
	return path, nil
}

func setWithTTL(arguments ...any) (any, error) {
	cacheHandle, err := expectHandle(arguments, 0, "core.cache.set_with_ttl")
	if err != nil {
		return nil, err
	}
	key, err := typemap.ExpectString(arguments, 1, "core.cache.set_with_ttl")
	if err != nil {
		return nil, err
	}
	if len(arguments) < 4 {
		return nil, fmt.Errorf("core.cache.set_with_ttl expected argument 3")
	}
	ttlSeconds, err := expectSeconds(arguments[3], "core.cache.set_with_ttl")
	if err != nil {
		return nil, err
	}
	path, err := cacheHandle.set(key, arguments[2], ttlFromSeconds(ttlSeconds))
	if err != nil {
		return nil, err
	}
	return path, nil
}

func getValue(arguments ...any) (any, error) {
	cacheHandle, err := expectHandle(arguments, 0, "core.cache.get")
	if err != nil {
		return nil, err
	}
	key, err := typemap.ExpectString(arguments, 1, "core.cache.get")
	if err != nil {
		return nil, err
	}
	value, found, err := cacheHandle.get(key)
	if err != nil {
		return nil, err
	}
	if found {
		return value, nil
	}
	if len(arguments) > 2 {
		return arguments[2], nil
	}
	return nil, nil
}

func hasValue(arguments ...any) (any, error) {
	cacheHandle, err := expectHandle(arguments, 0, "core.cache.has")
	if err != nil {
		return nil, err
	}
	key, err := typemap.ExpectString(arguments, 1, "core.cache.has")
	if err != nil {
		return nil, err
	}
	_, found, err := cacheHandle.get(key)
	if err != nil {
		return nil, err
	}
	return found, nil
}

func deleteValue(arguments ...any) (any, error) {
	cacheHandle, err := expectHandle(arguments, 0, "core.cache.delete")
	if err != nil {
		return nil, err
	}
	key, err := typemap.ExpectString(arguments, 1, "core.cache.delete")
	if err != nil {
		return nil, err
	}
	path, err := cacheHandle.path(key)
	if err != nil {
		return nil, err
	}
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return nil, fmt.Errorf("core.cache.delete failed to remove %q: %w", path, err)
	}
	return true, nil
}

func clearValues(arguments ...any) (any, error) {
	cacheHandle, err := expectHandle(arguments, 0, "core.cache.clear")
	if err != nil {
		return nil, err
	}
	prefix := ""
	if len(arguments) > 1 {
		prefix, err = typemap.ExpectString(arguments, 1, "core.cache.clear")
		if err != nil {
			return nil, err
		}
	}
	keys, err := cacheHandle.keys(prefix)
	if err != nil {
		return nil, err
	}
	removed := 0
	for _, key := range keys {
		path, err := cacheHandle.path(key)
		if err != nil {
			return nil, err
		}
		if err := os.Remove(path); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("core.cache.clear failed to remove %q: %w", path, err)
		}
		removed++
	}
	return removed, nil
}

func keys(arguments ...any) (any, error) {
	cacheHandle, err := expectHandle(arguments, 0, "core.cache.keys")
	if err != nil {
		return nil, err
	}
	prefix := ""
	if len(arguments) > 1 {
		prefix, err = typemap.ExpectString(arguments, 1, "core.cache.keys")
		if err != nil {
			return nil, err
		}
	}
	return cacheHandle.keys(prefix)
}

func expectHandle(arguments []any, index int, functionName string) (*handle, error) {
	if index >= len(arguments) {
		return nil, fmt.Errorf("%s expected argument %d", functionName, index)
	}
	value, ok := arguments[index].(*handle)
	if !ok {
		return nil, fmt.Errorf("%s expected cache handle, got %T", functionName, arguments[index])
	}
	return value, nil
}

func expectSeconds(value any, functionName string) (int, error) {
	seconds, ok := value.(int)
	if !ok {
		return 0, fmt.Errorf("%s expected ttl seconds to be int, got %T", functionName, value)
	}
	if seconds < 0 {
		return 0, fmt.Errorf("%s expected ttl seconds to be zero or positive", functionName)
	}
	return seconds, nil
}

func ttlFromSeconds(seconds int) time.Duration {
	if seconds == 0 {
		return defaultTTL
	}
	return time.Duration(seconds) * time.Second
}

func (cacheHandle *handle) set(key string, value any, ttl time.Duration) (string, error) {
	path, err := cacheHandle.path(key)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", fmt.Errorf("core.cache.set failed to create directory: %w", err)
	}

	now := time.Now()
	content, err := json.MarshalIndent(entry{
		Data:      value,
		CachedAt:  now,
		ExpiresAt: now.Add(ttl),
	}, "", "  ")
	if err != nil {
		return "", fmt.Errorf("core.cache.set failed to marshal entry: %w", err)
	}
	if err := os.WriteFile(path, content, 0644); err != nil {
		return "", fmt.Errorf("core.cache.set failed to write entry: %w", err)
	}
	return path, nil
}

func (cacheHandle *handle) get(key string) (any, bool, error) {
	path, err := cacheHandle.path(key)
	if err != nil {
		return nil, false, err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("core.cache.get failed to read entry: %w", err)
	}

	var decoded entry
	decoder := json.NewDecoder(strings.NewReader(string(content)))
	decoder.UseNumber()
	if err := decoder.Decode(&decoded); err != nil {
		return nil, false, nil
	}
	if time.Now().After(decoded.ExpiresAt) {
		_ = os.Remove(path)
		return nil, false, nil
	}
	return normalizeJSONValue(decoded.Data), true, nil
}

func (cacheHandle *handle) path(key string) (string, error) {
	parts, err := normalizedParts(key, false, "cache key")
	if err != nil {
		return "", err
	}

	path := filepath.Join(append([]string{cacheHandle.baseDir}, parts...)...) + ".json"
	return path, nil
}

func (cacheHandle *handle) keys(prefix string) ([]string, error) {
	prefixText, err := normalizedPrefix(prefix)
	if err != nil {
		return nil, err
	}

	result := []string{}
	if err := filepath.WalkDir(cacheHandle.baseDir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}

		relative, err := filepath.Rel(cacheHandle.baseDir, path)
		if err != nil {
			return err
		}
		key := strings.TrimSuffix(filepath.ToSlash(relative), ".json")
		if prefixText != "" && key != prefixText && !strings.HasPrefix(key, prefixText+"/") {
			return nil
		}
		result = append(result, key)
		return nil
	}); err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("core.cache.keys failed to walk cache: %w", err)
	}
	slices.Sort(result)
	return result, nil
}

func normalizedPrefix(prefix string) (string, error) {
	parts, err := normalizedParts(prefix, true, "cache prefix")
	if err != nil {
		return "", err
	}
	return strings.Join(parts, "/"), nil
}

func normalizedParts(value string, allowEmpty bool, fieldName string) ([]string, error) {
	text := strings.TrimSpace(strings.ReplaceAll(value, "\\", "/"))
	if text == "" {
		if allowEmpty {
			return nil, nil
		}
		return nil, fmt.Errorf("%s must not be empty", fieldName)
	}
	if strings.HasPrefix(text, "/") {
		return nil, fmt.Errorf("%s must be relative", fieldName)
	}

	parts := []string{}
	for _, part := range strings.Split(text, "/") {
		if part == "" || part == "." {
			continue
		}
		if part == ".." {
			return nil, fmt.Errorf("%s must not contain '..'", fieldName)
		}
		parts = append(parts, part)
	}
	if len(parts) == 0 && !allowEmpty {
		return nil, fmt.Errorf("%s must not be empty", fieldName)
	}
	return parts, nil
}

func normalizeJSONValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, item := range typed {
			result[key] = normalizeJSONValue(item)
		}
		return result
	case []any:
		result := make([]any, 0, len(typed))
		for _, item := range typed {
			result = append(result, normalizeJSONValue(item))
		}
		return result
	case json.Number:
		text := typed.String()
		if !strings.ContainsAny(text, ".eE") {
			if integerValue, err := typed.Int64(); err == nil {
				return int(integerValue)
			}
		}
		floatValue, err := typed.Float64()
		if err != nil {
			return text
		}
		return floatValue
	default:
		return value
	}
}
