package main

import (
	"log"

	"github.com/godbus/dbus"
	"go.uber.org/zap"

	"github.com/parrotmac/go-modemmanager/pkg/modem"
)

// Enumerates modems and dumps information about them
// May require elevated privileges
func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer logger.Sync()

	conn, err := dbus.SystemBus()
	if err != nil {
		logger.Fatal("system_bus.connection_failure", zap.Error(err))
	}

	mgr := modem.Manager{
		Logger:    logger,
		SystemBus: conn,
	}

	err = mgr.Scan()
	if err != nil {
		logger.Fatal("system_bus.scan_failure", zap.Error(err))
	}

	/*
		Get all modems
	*/
	modems, err := mgr.GetModemList()
	if err != nil {
		logger.Fatal("modem_manager.list_failure", zap.Error(err))
	}
	logger.Debug("modem_listing", zap.Any("data", modems))

	for _, m := range modems {
		/*
			Print info about the first Bearer, if any
		*/
		if len(m.Bearers) != 0 {
			bearer, err := mgr.GetBearer(m.Bearers[0])
			if err != nil {
				logger.Fatal("modem_manager.get_bearer_failure", zap.Error(err))
			}
			logger.Debug("modem_bearer_info", zap.Any("data", bearer))
		}

		/*
			Print info about the Sim, if available
		*/
		if m.Sim != "" {
			sim, err := mgr.GetSim(m.Sim)
			if err != nil {
				logger.Fatal("modem_manager.get_sim_failure", zap.Error(err))
			}
			logger.Debug("modem_sim_info", zap.Any("data", sim))
		}
	}
}
