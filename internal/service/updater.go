package service

import (
	"context"
	"log"
	"time"
)

type Updater struct {
	service  *Service
	interval time.Duration
	stopCh   chan struct{}
}

func NewUpdater(service *Service, interval time.Duration) *Updater {
	return &Updater{
		service:  service,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

func (u *Updater) Start() {
	ticker := time.NewTicker(u.interval)
	go func() {
		u.refresh()

		for {
			select {
			case <-ticker.C:
				u.refresh()
			case <-u.stopCh:
				ticker.Stop()
				return
			}
		}
	}()
}

func (u *Updater) Stop() {
	u.stopCh <- struct{}{}
	log.Println("Updater stopped")
}

func (u *Updater) refresh() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := u.service.RefreshRates(ctx); err != nil {
		log.Printf("failed to refresh rates: %s", err.Error())
	}
}
