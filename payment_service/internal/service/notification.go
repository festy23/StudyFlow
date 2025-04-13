package service

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
)

type kafkaNotifier struct {
	writer *kafka.Writer
}

func NewKafkaNotifier(brokers []string, topic string) *kafkaNotifier {
	w := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	return &kafkaNotifier{writer: w}
}

func (n *kafkaNotifier) SendPaymentNotification(ctx context.Context, lessonID, userID string) error {
	msg := struct {
		Type     string `json:"type"`
		LessonID string `json:"lesson_id"`
		UserID   string `json:"user_id"`
	}{
		Type:     "payment_submitted",
		LessonID: lessonID,
		UserID:   userID,
	}

	value, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	err = n.writer.WriteMessages(ctx, kafka.Message{
		Value: value,
	})
	if err != nil {
		return err
	}

	log.Printf("sent payment notification for lesson %s", lessonID)
	return nil
}

func (n *kafkaNotifier) SendReminderNotification(ctx context.Context, lessonID, userID string) error {
	msg := struct {
		Type     string `json:"type"`
		LessonID string `json:"lesson_id"`
		UserID   string `json:"user_id"`
	}{
		Type:     "payment_reminder",
		LessonID: lessonID,
		UserID:   userID,
	}

	value, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	err = n.writer.WriteMessages(ctx, kafka.Message{
		Value: value,
	})
	if err != nil {
		return err
	}

	log.Printf("sent reminder notification for lesson %s", lessonID)
	return nil
}

func (n *kafkaNotifier) Close() error {
	return n.writer.Close()
}
