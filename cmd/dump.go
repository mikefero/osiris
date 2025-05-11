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
	"context"
	"fmt"

	"github.com/mikefero/osiris/internal/app"
	"github.com/spf13/cobra"
)

var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dump a control plane configuration",
	Long: `The dump command gathers a control plane configuration, sanitizes it
(if enabled), and saves it to a file.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		startCtx, startCancel := context.WithCancel(context.Background())
		defer startCancel()
		app := app.NewDump()
		if err := app.Start(startCtx); err != nil {
			return fmt.Errorf("unable to start dump operation: %w", err)
		}

		stopCtx, stopCancel := context.WithCancel(context.Background())
		defer stopCancel()
		if err := app.Stop(stopCtx); err != nil {
			return fmt.Errorf("unable to stop dump operation: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(dumpCmd)
}
