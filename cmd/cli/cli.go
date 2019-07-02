package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ASCIILogo is noobaa's logo ascii art
const ASCIILogo = `
 /~~\\__~__//~~\
|               |
 \~\\_     _//~/
     \\   //
      |   |
      \~~~/
`

// CLI is the top command for noobaa CLI
var CLI = &cobra.Command{
	Use:  "noobaa",
	Long: "\n   NooBaa CLI \n" + ASCIILogo,
}

// InstallCommand installs to kubernetes
var InstallCommand = &cobra.Command{
	Use:   "install",
	Short: "Install to kubernetes",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Installing ...")
	},
}

func main() {
	CLI.AddCommand(InstallCommand)
	err := CLI.Execute()
	if err != nil {
		fmt.Printf("%+v\n", err)
	}
}
