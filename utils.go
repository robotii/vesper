package vesper

import (
	"os"
)

func exit(code int) {
	Cleanup()
	os.Exit(code)
}
