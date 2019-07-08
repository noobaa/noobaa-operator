package version

import "fmt"

var (
	// Version is the noobaa-operator version (semver)
	Version = "0.1.0"
)

func main() {
	fmt.Println(Version)
}
