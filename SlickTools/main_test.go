package slicktools

import (
    "encoding/json"
    "errors"
    "io/ioutil"
)

type Config struct {
    Setting1 string `json:"setting1"`
    Setting2 int    `json:"setting2"`
}

func GenerateConfig(path string) (*Config, error) {
    data, err := ioutil.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var config Config
    err = json.Unmarshal(data, &config)
    if err != nil {
        return nil, err
    }

    return &config, nil
}

func ValidateJson(jsonData string) (bool, error) {
    var js map[string]interface{}
    if err := json.Unmarshal([]byte(jsonData), &js); err != nil {
        return false, errors.New("invalid JSON")
    }
    return true, nil
}
