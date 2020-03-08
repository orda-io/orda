package model

import (
	"encoding/json"
	"strconv"
)

type OrtooType interface{}

func ConvertType(t interface{}) (OrtooType, error) {
	switch v := t.(type) {
	case int, int8, int16, int32, int64:
		var v64 int64
		switch vv := v.(type) {
		case int:
			v64 = int64(vv)
		case int8:
			v64 = int64(vv)
		case int16:
			v64 = int64(vv)
		case int32:
			v64 = int64(vv)
		case int64:
			v64 = vv
		}
		str := strconv.FormatInt(v64, 10)
		return str, nil
	case uint, uint8, uint16, uint32, uint64:
		var v64 uint64
		switch vv := v.(type) {
		case uint:
			v64 = uint64(vv)
		case uint8:
			v64 = uint64(vv)
		case uint16:
			v64 = uint64(vv)
		case uint32:
			v64 = uint64(vv)
		case uint64:
			v64 = vv
		}
		str := strconv.FormatUint(v64, 10)
		return str, nil
	case float32, float64:
		str := strconv.FormatFloat(v.(float64), 'e', -1, 64)
		return str, nil
	case string:
		return v, nil
	case bool:
		return v, nil
	default:
	}
	data, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	return string(data), nil
}
