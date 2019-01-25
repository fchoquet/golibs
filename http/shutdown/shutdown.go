package shutdown

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

// GracefulShutdown manages graceful shutdown of the passed server
func GracefulShutdown(server *http.Server, timeout time.Duration, logger logrus.FieldLogger) {
	// handles graceful shutdown
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			logger.Error(err.Error())
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	logger.Info("server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	server.Shutdown(ctx)
	logger.Info("server has shut down")
}
