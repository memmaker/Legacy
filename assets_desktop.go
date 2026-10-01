//go:build !js

package main

import "os"

var assetFS = os.DirFS(".")
