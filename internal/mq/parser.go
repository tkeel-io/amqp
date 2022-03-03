package mq

import "encoding/json"

func parser(data []byte) ([]byte, error) {
	info := make(map[string]interface{})
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}
	if r, err := json.Marshal(info["data"]); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}
