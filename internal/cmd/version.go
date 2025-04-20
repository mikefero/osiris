/*
Copyright © 2025 Michael Fero

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package cmd contains the command line package.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// AppName is the name of the application.
	AppName string
	// Version is the version of the application.
	Version string
	// Commit is the git commit hash of the source tree.
	Commit string
	// OsArch is the OS and architecture of the build.
	OsArch string
	// GoVersion is the version of go used to build the application.
	GoVersion string
	// BuildDate is the date the application was built.
	BuildDate string
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the app-name version",
	Long: `The version command prints the version of app-name along with a git
commit hash of the source tree, OS, architecture, go version, and build date.`,
	Run: func(_ *cobra.Command, _ []string) {
		formatVersion := Version
		if len(formatVersion) == 0 {
			formatVersion = "dev"
		}
		if len(Commit) > 0 {
			formatVersion = fmt.Sprintf("%s (%s)", formatVersion, Commit)
		}
		if len(OsArch) > 0 {
			formatVersion = fmt.Sprintf("%s %s", formatVersion, OsArch)
		}
		fmt.Printf("%s version %s\n", AppName, formatVersion) //nolint:forbidigo
		if len(GoVersion) > 0 {
			fmt.Printf("go version %s\n", GoVersion) //nolint:forbidigo
		}
		if len(BuildDate) > 0 {
			fmt.Printf("Built on %s\n", BuildDate) //nolint:forbidigo
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
