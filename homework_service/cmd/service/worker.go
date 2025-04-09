package main

import (
	"context"
	"time"

	"homework_service/internal/repository"
	"homework_service/pkg/kafka"
	"homework_service/pkg/logger"
)

type ReminderWorker struct {
	assignmentRepo repository.AssignmentRepository
	kafkaProducer  *kafka.Producer
	logger         *logger.Logger
	interval       time.Duration
}

func NewReminderWorker(
	assignmentRepo repository.AssignmentRepository,
	kafkaProducer *kafka.Producer,
	logger *logger.Logger,
) *ReminderWorker {
	return &ReminderWorker{
		assignmentRepo: assignmentRepo,
		kafkaProducer:  kafkaProducer,
		logger:         logger,
		interval:       time.Minute,
	}
}

func (w *ReminderWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Reminder worker stopped")
			return
		case <-ticker.C:
			w.processReminders(ctx)
		}
	}
}

func (w *ReminderWorker) processReminders(ctx context.Context) {
	assignments, err := w.assignmentRepo.FindAssignmentsDueSoon(ctx, 24*time.Hour)
	if err != nil {
		w.logger.Errorf("Failed to get assignments due soon: %v", err)
		return
	}

	for _, assignment := range assignments {
		message := map[string]interface{}{
			"assignment_id": assignment.ID,
			"student_id":    assignment.StudentID,
			"tutor_id":      assignment.TutorID,
			"due_date":      assignment.DueDate,
			"title":         assignment.Title,
		}

		if err := w.kafkaProducer.Send(ctx, "assignment-reminders", message); err != nil {
			w.logger.Errorf("Failed to send reminder for assignment %s: %v", assignment.ID, err)
			continue
		}

		w.logger.Infof("Sent reminder for assignment %s", assignment.ID)
	}
}
