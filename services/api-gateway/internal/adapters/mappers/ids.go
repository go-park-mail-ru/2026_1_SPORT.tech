package mappers

import (
	"fmt"
	"math"
)

func int32ToInt64(value int32) int64 {
	return int64(value)
}

func int32SliceToInt64Slice(values []int32) []int64 {
	result := make([]int64, 0, len(values))
	for _, value := range values {
		result = append(result, int64(value))
	}

	return result
}

func optionalInt32ToInt64(value *int32) *int64 {
	if value == nil {
		return nil
	}

	converted := int64(*value)
	return &converted
}

func optionalInt64ToInt32(field string, value *int64) (*int32, error) {
	if value == nil {
		return nil, nil
	}

	converted, err := int64ToInt32(field, *value)
	if err != nil {
		return nil, err
	}

	return &converted, nil
}

func int64ToInt32(field string, value int64) (int32, error) {
	if value < math.MinInt32 || value > math.MaxInt32 {
		return 0, fmt.Errorf("%s value %d is out of int32 range", field, value)
	}

	return int32(value), nil
}
