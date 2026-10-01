//go:build js

package main

import "embed"

// Saves need no shim on wasm: os file calls fail with ENOSYS and saveload.go already logs and skips.
//
//go:embed assets/*.png assets/Legacy.ldtk assets/dialogues assets/npc assets/scrolls
var assetFS embed.FS
