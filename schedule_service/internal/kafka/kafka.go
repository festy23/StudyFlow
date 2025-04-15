// Файл: internal/kafka/event_sender.go
package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type EventSender struct {
	writer        *kafka.Writer
	reminderTopic string
}

type ReminderEvent struct {
	LessonID       string    `json:"lesson_id"`
	SlotID         string    `json:"slot_id"`
	TutorID        string    `json:"tutor_id"`
	StudentID      string    `json:"student_id"`
	StartsAt       time.Time `json:"starts_at"`
	EndsAt         time.Time `json:"ends_at"`
	ReminderType   string    `json:"reminder_type"` // "24h" or "1h"
	ConnectionLink string    `json:"connection_link,omitempty"`
}

func NewEventSender(brokers []string, reminderTopic string) *EventSender {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        reminderTopic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Async:        false,
	}

	return &EventSender{
		writer:        writer,
		reminderTopic: reminderTopic,
	}
}

func (s *EventSender) Close() error {
	return s.writer.Close()
}

func (s *EventSender) SendReminderEvent(ctx context.Context, event ReminderEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal reminder event: %w", err)
	}

	message := kafka.Message{
		Key:   []byte(event.LessonID),
		Value: data,
		Time:  time.Now(),
	}

	if err := s.writer.WriteMessages(ctx, message); err != nil {
		return fmt.Errorf("failed to send reminder event: %w", err)
	}

	return nil
}
