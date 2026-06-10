package bundle

import (
	"github.com/noobaa/noobaa-operator/v5/deploy"
)

// MustRead reads an embedded deploy file by path and returns its content as a string.
// It panics if the file is not found, matching the previous behavior where all
// bundled content was guaranteed to exist as compiled-in constants.
func MustRead(path string) string {
	data, err := deploy.FS.ReadFile(path)
	if err != nil {
		panic("bundle: missing embedded file: " + path)
	}
	return string(data)
}
