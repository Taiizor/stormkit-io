package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/stormkit-io/stormkit-io/src/ce/api/admin"
	"github.com/stormkit-io/stormkit-io/src/ce/api/router"
	"github.com/stormkit-io/stormkit-io/src/lib/config"
	"github.com/stormkit-io/stormkit-io/src/lib/database"
	"github.com/stormkit-io/stormkit-io/src/lib/rediscache"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp/limiter"
	"github.com/stormkit-io/stormkit-io/src/lib/slog"
	"github.com/stormkit-io/stormkit-io/src/lib/tracking"
	"github.com/stormkit-io/stormkit-io/src/lib/utils"
)

var httpPort = utils.StringToInt(utils.GetString(os.Getenv("STORMKIT_HTTP_POST"), "8080"))

// registerListeners registers listeners for the hosting machine.
func registerListeners() {
	service := rediscache.Service()

	handlers := map[string]rediscache.Handler{
		rediscache.EventInvalidateAdminCache: admin.ResetCache,
	}

	for event, handler := range handlers {
		if err := service.SubscribeAsync(event, handler); err != nil {
			slog.Errorf("failed to register event %s: %v", event, err)
		}
	}
}

func registerServices() {
	c := config.Get()
	r := router.Get()

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", httpPort),
		ReadTimeout:  c.HTTPTimeouts.ReadTimeout,
		WriteTimeout: c.HTTPTimeouts.WriteTimeout,
		IdleTimeout:  c.HTTPTimeouts.IdleTimeout,
		Handler:      r.WithContext().Handler(),
	}

	slog.Infof("api server listening on :%d", httpPort)
	log.Fatal(srv.ListenAndServe())
}

func main() {
	os.Setenv("STORMKIT_SERVICE_NAME", rediscache.ServiceApi)

	c := config.Get()

	if c.Tracking.Prometheus {
		tracking.Prometheus(tracking.PrometheusOpts{})
	}

	go limiter.Cleanup()

	conn := database.Connection()

	if config.IsStormkitCloud() && conn == nil {
		panic("database connection is nil -- check slog.errors")
	}

	registerListeners()
	registerServices()
}
