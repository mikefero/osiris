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
package cmd

import (
	"fmt"

	"github.com/mikefero/osiris/internal/app"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the app-name version",
	Long: `The version command prints the version of app-name along with a git
commit hash of the source tree, OS, architecture, go version, and build date.`,
	Run: func(_ *cobra.Command, _ []string) {
		formatVersion := app.Version
		if len(formatVersion) == 0 {
			formatVersion = "dev"
		}
		if len(app.Commit) > 0 {
			formatVersion = fmt.Sprintf("%s (%s)", formatVersion, app.Commit)
		}
		if len(app.OsArch) > 0 {
			formatVersion = fmt.Sprintf("%s %s", formatVersion, app.OsArch)
		}
		fmt.Printf("%s version %s\n", app.AppName, formatVersion) //nolint:forbidigo
		if len(app.GoVersion) > 0 {
			fmt.Printf("go version %s\n", app.GoVersion) //nolint:forbidigo
		}
		if len(app.BuildDate) > 0 {
			fmt.Printf("Built on %s\n", app.BuildDate) //nolint:forbidigo
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
