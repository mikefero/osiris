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
	"os"

	"github.com/spf13/cobra"
)

var license string

// Options contains the options for the root command.
type Options struct {
	// License is the license of the application.
	License string
}

var rootCmd = &cobra.Command{
	Use:   "app-name",
	Short: "Application Name",
	Long:  `The app-name description.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute(opts Options) {
	license = opts.License
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
