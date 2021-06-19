package storage

import (
	"context"
	"log"
	"time"

	"github.com/TwinProduction/gatus/storage/store"
	"github.com/TwinProduction/gatus/storage/store/memory"
)

var (
	provider store.Store

	// initialized keeps track of whether the storage provider was initialized
	// Because store.Store is an interface, a nil check wouldn't be sufficient, so instead of doing reflection
	// every single time Get is called, we'll just lazily keep track of its existence through this variable
	initialized bool

	ctx        context.Context
	cancelFunc context.CancelFunc
)

// Get retrieves the storage provider
func Get() store.Store {
	if !initialized {
		log.Println("[storage][Get] Provider requested before it was initialized, automatically initializing")
		// TODO should this be disallowed? It's using an empty config.
		err := Initialize(nil)
		if err != nil {
			panic("failed to automatically initialize store: " + err.Error())
		}
	}
	return provider
}

// Initialize instantiates the storage provider based on the Config provider
func Initialize(cfg *Config) error {
	// initialize can happen multiple times, e.g. if the config file is updated,
	// but we only need to track if it happened at least once
	initialized = true

	var err error
	if cancelFunc != nil {
		// Stop the active autoSave task
		cancelFunc()
	}
	// TODO select file or database store here
	if cfg == nil || len(cfg.File) == 0 {
		log.Println("[storage][Initialize] Creating storage provider")
		provider, _ = memory.NewStore("")
	} else {
		ctx, cancelFunc = context.WithCancel(context.Background())
		log.Printf("[storage][Initialize] Creating storage provider with file=%s", cfg.File)
		provider, err = memory.NewStore(cfg.File)
		if err != nil {
			return err
		}
		go autoSave(cfg.AutoSaveInterval, ctx)
	}
	return nil
}

// autoSave automatically calls the SaveFunc function of the provider at every interval
func autoSave(interval time.Duration, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Printf("[storage][autoSave] Stopping active job")
			return
		case <-time.After(interval):
			log.Printf("[storage][autoSave] Saving")
			err := provider.Save()
			if err != nil {
				log.Println("[storage][autoSave] Save failed:", err.Error())
			}
		}
	}
}
