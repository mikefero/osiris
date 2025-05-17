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
package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/mikefero/osiris/internal/client"
	"github.com/mikefero/osiris/internal/config"
	"github.com/mikefero/osiris/internal/logger"
	"github.com/mikefero/osiris/internal/resource"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

// NewDump creates a new fx application for the dump command.
// It provides the necessary dependencies and registers the dump functionality.
func NewDump() *fx.App {
	return fx.New(
		fx.Provide(
			config.NewConfig,
			func(config *config.Config) (*zap.Logger, error) {
				return logger.NewLogger(config.Logger, logger.LoggerCommandTypeDump)
			},
		),
		fx.WithLogger(func(logger *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: logger}
		}),
		fx.Invoke(registerDump),
	)
}

func registerDump(lc fx.Lifecycle, config *config.Config, logger *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("starting osiris",
				zap.String("version", Version),
				zap.String("commit", Commit),
				zap.String("os-arch", OsArch),
				zap.String("go-version", GoVersion),
				zap.String("build-date", BuildDate),
			)
			logger.Info("Starting dump")
			client := client.NewClient(config, logger)
			if results, err := listData(ctx, client, logger); err != nil {
				logger.Error("error executing dump", zap.Error(err))
				return fmt.Errorf("error listing data: %w", err)
			} else {
				if err := writeResults(results, logger, config.OutputFile); err != nil {
					logger.Error("error writing results",
						zap.String("output-filename", config.OutputFile),
						zap.Error(err))
					return fmt.Errorf("error writing results: %w", err)
				}
			}
			logger.Info("Dump completed successfully")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("stopping osiris")
			if err := logger.Sync(); err != nil {
				logger.Error("failed to sync logger", zap.Error(err))
			}
			return nil
		},
	})
}

// ListData lists data from all resources in the resource registry.
// It uses goroutines to GET data concurrently and collects the results.
func listData(ctx context.Context, client *client.Client, logger *zap.Logger) ([]resource.ResourceData, error) {
	resources := resource.ResourceRegistry
	errChan := make(chan error, len(resources))
	var mutex sync.Mutex
	var results []resource.ResourceData
	var wg sync.WaitGroup

	logger.Info("Listing data from resources",
		zap.Int("resource-count", len(resources)))

	// Iterate over the resources and start a goroutine for each one
	startTime := time.Now()
	for _, res := range resources {
		wg.Add(1)
		go func(res resource.Resource) {
			defer wg.Done()

			// List the resource items
			data, err := res.List(ctx, client, logger)
			if err != nil {
				logger.Error("error listing resource",
					zap.String("resource", res.Name()),
					zap.Error(err))
				errChan <- fmt.Errorf("error listing resource %s: %w", res.Name(), err)
				return
			}

			mutex.Lock()
			results = append(results, data)
			mutex.Unlock()
		}(res)
	}

	// Rest of the function remains the same...
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		logger.Warn("Context was canceled while listing data from resources",
			zap.Error(ctx.Err()))
		return nil, ctx.Err()
	case <-done:
		close(errChan)
		if len(errChan) > 0 {
			err := <-errChan
			logger.Error("Error occurred while listing data from resources",
				zap.Error(err))
			return nil, err
		}
	}

	logger.Info("Successfully listed data from resources",
		zap.Int("resource-count", len(resources)),
		zap.Duration("duration", time.Since(startTime)))

	return results, nil
}

func writeResults(results []resource.ResourceData, logger *zap.Logger, outputFilename string) error {
	// Create a map where the keys are the endpoint names
	resultMap := make(map[string][]map[string]interface{})

	// Convert the slice of Results to a map
	for _, result := range results {
		resultMap[result.Name] = result.Data
	}

	logger.Info("Marshaling results to JSON",
		zap.Int("endpointCount", len(resultMap)))

	// Marshal the map to JSON with pretty formatting
	startTime := time.Now()
	jsonData, err := json.MarshalIndent(resultMap, "", "  ")
	if err != nil {
		logger.Error("error marshaling results", zap.Error(err))
		return fmt.Errorf("error marshaling results: %w", err)
	}

	logger.Debug("Writing results to file",
		zap.String("output-filename", outputFilename),
		zap.Int("bytes", len(jsonData)),
		zap.Duration("duration", time.Since(startTime)))

	if err := os.WriteFile(outputFilename, jsonData, 0o600); err != nil {
		logger.Error("error writing file",
			zap.String("output-filename", outputFilename),
			zap.Error(err))
		return fmt.Errorf("error writing file: %w", err)
	}

	logger.Info("Successfully wrote results to JSON file",
		zap.String("output-filename", outputFilename),
		zap.Int("bytes", len(jsonData)),
		zap.Duration("duration", time.Since(startTime)))

	return nil
}
