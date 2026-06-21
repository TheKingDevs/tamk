package tools

import (
	_ "embed"
)

//go:embed binaries/sdk/android.jar
var androidJar []byte
