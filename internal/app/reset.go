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
	"fmt"
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

// NewReset creates a new fx application for the reset command.
// It provides the necessary dependencies and registers the reset functionality.
func NewReset() *fx.App {
	return fx.New(
		fx.Provide(
			config.NewConfig,
			func(config *config.Config) (*zap.Logger, error) {
				return logger.NewLogger(config.Logger, logger.LoggerCommandTypeReset)
			},
		),
		fx.WithLogger(func(logger *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: logger}
		}),
		fx.Invoke(registerReset),
	)
}

func registerReset(lc fx.Lifecycle, config *config.Config, logger *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("starting osiris",
				zap.String("version", Version),
				zap.String("commit", Commit),
				zap.String("os-arch", OsArch),
				zap.String("go-version", GoVersion),
				zap.String("build-date", BuildDate),
			)
			logger.Info("Starting reset operation")
			client := client.NewClient(config, logger)
			if err := deleteData(ctx, client, logger); err != nil {
				logger.Error("error executing reset", zap.Error(err))
				return fmt.Errorf("error deleting data: %w", err)
			}
			logger.Info("Reset completed successfully")
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

func deleteData(ctx context.Context, client *client.Client, logger *zap.Logger) error {
	// Get ordered resources for deletion - Leaf items need to be deleted first
	registry := resource.NewRegistry()
	logger.Debug("Generating resource dependency graph for deletion")
	levels, err := registry.GetResourcesForDeletion()
	if err != nil {
		return fmt.Errorf("error generating deletion order: %w", err)
	}

	logger.Info("Deleting data from resources",
		zap.Int("levels", len(levels)),
		zap.Int("resource-count", len(registry.GetResources())))

	// Process each level in sequence
	startTime := time.Now()
	for levelIdx, level := range levels {
		levelStartTime := time.Now()
		logger.Debug("Processing deletion level",
			zap.Int("level", levelIdx+1),
			zap.Int("levels", len(level)))

		var wg sync.WaitGroup
		errChan := make(chan error, len(level))
		levelCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		// Process all resources at this level in parallel
		for _, res := range level {
			wg.Add(1)
			go func(r resource.Resource) {
				defer wg.Done()
				resStartTime := time.Now()

				// Get all items for this resource
				logger.Debug("Listing resource items", zap.String("resource", r.Name()))
				resourceData, listErr := r.List(levelCtx, client, logger)
				if listErr != nil {
					logger.Error("error listing resource",
						zap.String("resource", r.Name()),
						zap.Error(listErr))
					errChan <- fmt.Errorf("error listing resource %s: %w", r.Name(), listErr)
					return
				}

				itemCount := len(resourceData.Data)
				if itemCount == 0 {
					logger.Debug("No items to delete",
						zap.String("resource", r.Name()),
						zap.Duration("duration", time.Since(resStartTime)))
					return
				}
				logger.Info("Deleting resource items",
					zap.String("resource", r.Name()),
					zap.Int("count", itemCount))

				// Delete each item for this resource - fail fast on first error
				for i, item := range resourceData.Data {
					// Check if the context is done before proceeding with deletion
					select {
					case <-levelCtx.Done():
						return // Context was canceled, stop processing
					default:
						// Continue with deletion
					}

					if deleteErr := r.Delete(levelCtx, client, item, logger); deleteErr != nil {
						logger.Error("error deleting item",
							zap.String("resource", r.Name()),
							zap.Int("item", i+1),
							zap.Int("total", itemCount),
							zap.Error(deleteErr))
						errChan <- fmt.Errorf("error deleting item %d/%d for %s: %w",
							i+1, itemCount, r.Name(), deleteErr)
						return
					}
				}

				logger.Info("Successfully deleted items from resource",
					zap.String("resource", r.Name()),
					zap.Int("count", itemCount),
					zap.Duration("duration", time.Since(resStartTime)))
			}(res)
		}

		// Set up a channel to signal completion
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		// Wait for either completion, error, or context cancellation
		select {
		case <-ctx.Done():
			logger.Warn("Context was canceled while deleting resources",
				zap.Error(ctx.Err()))
			return ctx.Err()
		case err := <-errChan:
			logger.Error("Error occurred during resource deletion",
				zap.Int("level", levelIdx+1),
				zap.Error(err))
			return err
		case <-done:
			// All goroutines completed successfully
		}

		levelDuration := time.Since(levelStartTime)
		logger.Info("Completed deletion level",
			zap.Int("level", levelIdx+1),
			zap.Duration("duration", levelDuration))
	}

	totalDuration := time.Since(startTime)
	logger.Info("Successfully deleted all resources",
		zap.Int("levels", len(levels)),
		zap.Int("resource-count", len(registry.GetResources())),
		zap.Duration("duration", totalDuration))

	return nil
}
