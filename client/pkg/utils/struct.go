package utils

import "github.com/mitchellh/mapstructure"

// StructToMap is used to transform struct to map type
func StructToMap(in interface{}) (map[string]interface{}, error) {
	toMap := make(map[string]interface{})
	msDecoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &toMap,
	})
	if err != nil {
		return nil, err
	}
	if err := msDecoder.Decode(in); err != nil {
		return nil, err
	}
	return toMap, nil
}
