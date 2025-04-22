package worker

import (
	"context"
	"log"
	"time"

	"payment_service/internal/service"
)

type ReminderWorker struct {
	svc           *service.PaymentService
	checkInterval time.Duration
}

func NewReminderWorker(svc *service.PaymentService, interval time.Duration) *ReminderWorker {
	return &ReminderWorker{
		svc:           svc,
		checkInterval: interval,
	}
}

func (w *ReminderWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			lessons, err := w.svc.GetUnpaidLessons(ctx)
			if err != nil {
				log.Printf("failed to get unpaid lessons: %v", err)
				continue
			}

			for _, lesson := range lessons {
				if err := w.svc.SendReminder(ctx, lesson); err != nil {
					log.Printf("failed to send reminder: %v", err)
				}
			}
			log.Println("Reminder worker tick")
		}
	}
}
