package validations

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	e "web_server/domain/entities"
	env "web_server/environments/processors"
	u "web_server/utils"
)

// Validation Input/Output Structures
type ValidateEnvInput struct {
	Data       interface{}
	ParentPath []string
}

type ValidateEnvOutput struct {
	IsValid  bool
	Errors   []error
	ErrorMap map[string][]error
}

var (
	allowedValueKinds = map[reflect.Kind]bool{
		reflect.String:  true,
		reflect.Int:     true,
		reflect.Float64: true,
		reflect.Bool:    true,
		reflect.Map:     true,
		reflect.Slice:   true,
		reflect.Struct:  true,
	}
)

type ValidationContext struct {
	Path []string
}

// Main Validation Entry Points
func ValidateEnvData(data e.EnvData[any]) ValidateEnvOutput {
	ctx := &ValidationContext{
		Path: []string{"EnvData"},
	}
	return ValidateEnvDataWithContext(data, ctx)
}

func ValidateInternalEnvMapData[K comparable, V any](data e.EnvMapData[K, V]) ValidateEnvOutput {
	ctx := &ValidationContext{
		Path: []string{"EnvMapData"},
	}
	return validateEnvMapData(reflect.ValueOf(data), ctx)
}

// Core Validation Logic
func ValidateEnvDataWithContext(data interface{}, ctx *ValidationContext) ValidateEnvOutput {
	val := reflect.ValueOf(data)
	result := ValidateEnvOutput{
		ErrorMap: make(map[string][]error),
	}

	if val.Kind() == reflect.Struct {
		if hasField(val, "Value") { // Check for EnvData
			result = validateEnvDataStruct(val, ctx)
		} else if hasField(val, "EnvInfos") { // Check for EnvMapData
			result = validateEnvMapData(val, ctx)
		} else {
			if err := deepValidate(val, ctx); err != nil {
				result.appendError(err, ctx)
			}
		}
	} else {
		if err := deepValidate(val, ctx); err != nil {
			result.appendError(err, ctx)
		}
	}

	result.IsValid = len(result.Errors) == 0
	return result
}

// Helper to check if a struct has a specific field
func hasField(val reflect.Value, fieldName string) bool {
	if val.Kind() != reflect.Struct {
		return false
	}
	_, ok := val.Type().FieldByName(fieldName)
	return ok
}

// ... (keep the rest of the functions as before, but update any references to envMapDataType)

func validateEnvInfosMap(val reflect.Value, ctx *ValidationContext) error {
	if val.Kind() != reflect.Map {
		return env.GetFuncError(env.InvalidMapType, nil, ctx.Path)
	}

	var wg sync.WaitGroup
	errChan := make(chan error, val.Len())
	mapKeys := val.MapKeys()

	for _, key := range mapKeys {
		wg.Add(1)
		go func(k reflect.Value) {
			defer wg.Done()

			keyCtx := withContext(ctx, fmt.Sprintf("[%v]", k.Interface()))
			mapVal := val.MapIndex(k)

			// Check if map value is EnvData by structure
			if !isEnvData(mapVal) {
				errChan <- env.GetFuncError(env.InvalidMapValueType, nil, keyCtx.Path)
				return
			}

			if err := deepValidate(mapVal, keyCtx); err != nil {
				errChan <- err
			}
		}(key)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	for err := range errChan {
		if err != nil {
			return err
		}
	}
	return nil
}

// Check if a value is EnvData by field presence
func isEnvData(val reflect.Value) bool {
	return hasField(val, "Value") && hasField(val, "StatusInfo")
}

func validateEnvDataStruct(val reflect.Value, ctx *ValidationContext) ValidateEnvOutput {
	result := ValidateEnvOutput{ErrorMap: make(map[string][]error)}

	// Validate Value field
	valueField := val.FieldByName("Value")
	valueCtx := withContext(ctx, "Value")
	if err := validateValueField(valueField, valueCtx); err != nil {
		result.appendError(err, valueCtx)
	}

	// Validate StatusInfo
	statusCtx := withContext(ctx, "StatusInfo")
	if err := validateStatusInfo(val, statusCtx); err != nil {
		result.appendError(err, statusCtx)
	}

	result.IsValid = len(result.Errors) == 0
	return result
}

func validateEnvMapData(val reflect.Value, ctx *ValidationContext) ValidateEnvOutput {
	result := ValidateEnvOutput{ErrorMap: make(map[string][]error)}

	envInfos := val.FieldByName("EnvInfos")
	envInfosCtx := withContext(ctx, "EnvInfos")
	if err := validateEnvInfosMap(envInfos, envInfosCtx); err != nil {
		result.appendError(err, envInfosCtx)
	}

	statusCtx := withContext(ctx, "StatusInfos")
	if err := validateStatusInfo(val, statusCtx); err != nil {
		result.appendError(err, statusCtx)
	}

	result.IsValid = len(result.Errors) == 0
	return result
}

// Field-specific Validation
func validateValueField(val reflect.Value, ctx *ValidationContext) error {
	if !val.IsValid() || (val.Kind() == reflect.Ptr && val.IsNil()) {
		return env.GetFuncError(env.NilValueNotAllowed, nil, ctx.Path)
	}

	if !allowedValueKinds[val.Kind()] {
		return env.GetFuncError(env.UnSupportedDataType, nil, ctx.Path)
	}

	return deepValidate(val, ctx)
}

func validateStatusInfo(val reflect.Value, ctx *ValidationContext) error {
	statusField := val.FieldByName("StatusInfo")
	if !statusField.IsValid() {
		return env.GetFuncError(env.MissingStatusInfo, nil, ctx.Path)
	}
	return deepValidate(statusField, ctx)
}

// Enhanced Deep Validation
func deepValidate(val reflect.Value, ctx *ValidationContext) error {
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return env.GetFuncError(env.NilValueNotAllowed, nil, ctx.Path)
		}
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Map:
		return validateGenericMap(val, ctx)
	case reflect.Slice, reflect.Array:
		return validateSlice(val, ctx)
	case reflect.Struct:
		if isEnvData(val) { // Yapısal kontrol kullanılıyor
			if validationResult := validateEnvDataStruct(val, ctx); len(validationResult.Errors) > 0 {
				return validationResult.Errors[0]
			}
			return nil
		}
		return validateStruct(val, ctx)
	default:
		if !allowedValueKinds[val.Kind()] {
			return env.GetFuncError(env.UnSupportedDataType, nil, ctx.Path)
		}
	}
	return nil
}

// Concurrent Map Validation
func validateGenericMap(val reflect.Value, ctx *ValidationContext) error {
	if val.Type().Key().Kind() == reflect.Uintptr {
		return env.GetFuncError(env.AllFieldsRequiredWithInvalidKey, nil, ctx.Path)
	}

	var wg sync.WaitGroup
	errChan := make(chan error, val.Len())

	for _, key := range val.MapKeys() {
		wg.Add(1)
		go func(k reflect.Value) {
			defer wg.Done()

			newCtx := withContext(ctx, fmt.Sprintf("[%v]", k.Interface()))
			mapVal := val.MapIndex(k)

			if err := deepValidate(mapVal, newCtx); err != nil {
				errChan <- env.GetFuncError(env.InvalidMapValue, err, newCtx.Path)
			}
		}(key)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	for err := range errChan {
		return err
	}
	return nil
}

// Helper Functions
func withContext(ctx *ValidationContext, fieldName string) *ValidationContext {
	newCtx := &ValidationContext{
		Path: make([]string, len(ctx.Path)),
	}
	copy(newCtx.Path, ctx.Path)
	newCtx.Path = append(newCtx.Path, fieldName)
	return newCtx
}

func (o *ValidateEnvOutput) appendError(err error, ctx *ValidationContext) {
	path := strings.Join(ctx.Path, ".")
	o.Errors = append(o.Errors, err)
	o.ErrorMap[path] = append(o.ErrorMap[path], err)
}

func validateSlice(val reflect.Value, ctx *ValidationContext) error {
	for i := 0; i < val.Len(); i++ {
		element := val.Index(i)
		indexCtx := withContext(ctx, fmt.Sprintf("[%d]", i))

		if err := deepValidate(element, indexCtx); err != nil {
			return env.GetFuncError(env.InvalidSliceElement, err, indexCtx.Path)
		}
	}
	return nil
}

func validateStruct(val reflect.Value, ctx *ValidationContext) error {
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := val.Type().Field(i)
		fieldCtx := withContext(ctx, fieldType.Name)

		if !field.CanInterface() || fieldType.Tag.Get("cbor") == "-" {
			continue
		}

		if err := deepValidate(field, fieldCtx); err != nil {
			return err
		}
	}
	return nil
}

// Bütün dilleri desteklemek adına byte odaklı çalışma
func ValidateExternalEnvData[V any](data e.EnvData[V]) error {
	switch value := any(data.Value).(type) {
	case []byte:
		if len(value) == 0 {
			return env.GetFuncError(env.InvalidValue, nil, value)
		}
	default:
		return env.GetFuncError(env.InvalidMapValueType, nil, value)
	}
	return nil
}

func ValidateExternalEnvMapData[K comparable, V any](data e.EnvMapData[K, e.EnvData[V]]) error {
	if data.EnvInfos == nil {
		return env.GetFuncError(env.InvalidMapValue, nil, nil)
	}

	//alınan new env data status bilgisi kontrol edilir.
	if err := u.CheckNewDataStatusInfos(&e.CheckDataStatusInfosInput{
		Status:      data.StatusInfos.Status,
		ActiveAt:    data.StatusInfos.ActiveAt,
		ExpiresAt:   data.StatusInfos.ExpiresAt,
		Description: data.StatusInfos.Description,
	}); err != nil {
		return err
	}

	for key, envData := range data.EnvInfos {
		// Eğer `K` string ise boş olamaz
		var zeroK K
		if key == zeroK {
			return env.GetFuncError(env.InvalidMapValueType, nil, key)
		}

		// EnvData doğrulaması
		if err := ValidateExternalEnvData(envData); err != nil {
			return env.GetFuncError(env.InvalidMapValue, err, key)
		}
	}

	return nil
}
