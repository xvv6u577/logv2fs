package main

import (
	"fmt"

	yamlTools "github.com/caster8013/logv2rayfullstack/yaml"
)

func main() {
	err := yamlTools.GenerateAllClashxConfig()
	if err != nil {
		fmt.Printf("err: %v", err)
	}
}
