package util

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

// LoadYAML Loads YAML into a given object
func LoadYAML(y []byte, p interface{}) {
	// Load the YAML into the object p, which is a pointer to something that
	// can accept YAML
	err := yaml.Unmarshal(y[:len(y)], p)
	if err != nil {
		fmt.Printf("%s", err)
		panic(err)
	}
}
