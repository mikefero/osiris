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

	"github.com/mikefero/osiris/internal/client"
	"github.com/mikefero/osiris/internal/config"
	"github.com/mikefero/osiris/internal/logger"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

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
			if err := client.GatherData(ctx); err != nil {
				logger.Error("error executing dump", zap.Error(err))
				return err
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
