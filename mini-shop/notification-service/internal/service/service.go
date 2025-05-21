package service

import (
	"encoding/json"
	"fmt"
	"log"
	"net/smtp"

	"github.com/Viltsev/notification-service/internal/messaging"
)

type NotificationService struct {
	rabbitMQ *messaging.RabbitMQ
	smtpHost string
	smtpPort string
	auth     smtp.Auth
	from     string
}

func NewNotificationService(rabbitMQ *messaging.RabbitMQ) *NotificationService {
	smtpHost := "smtp.mail.ru"
	smtpPort := "587"
	from := "pm_csit2025@mail.ru"
	auth := smtp.PlainAuth("", "pm_csit2025@mail.ru", "T7bHKeHHP8Afg2cCQwXS", smtpHost)

	return &NotificationService{
		rabbitMQ: rabbitMQ,
		smtpHost: smtpHost,
		smtpPort: smtpPort,
		auth:     auth,
		from:     from,
	}
}

func (s *NotificationService) SendEmail(to, subject, body string) error {
	msg := "From: " + s.from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n" +
		body

	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)
	log.Println("About to call smtp.SendMail")
	err := smtp.SendMail(addr, s.auth, s.from, []string{to}, []byte(msg))
	log.Printf("smtp.SendMail returned: %v", err)

	if err != nil {
		log.Printf("Failed to send email to %s: %v", to, err)
		return err
	}
	log.Printf("Email sent to %s", to)
	return nil
}

func (s *NotificationService) HandlePaymentEvent(body []byte) {
	var event map[string]interface{}
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("Failed to unmarshal payment event: %v", err)
		return
	}

	eventType, ok := event["type"].(string)
	if !ok {
		log.Println("Missing or invalid 'type' in event")
		return
	}

	orderIDFloat, ok := event["orderID"].(float64)
	if !ok {
		log.Println("Missing or invalid 'orderID' in event")
		return
	}
	orderID := int(orderIDFloat)

	amountFloat, ok := event["amount"].(float64)
	if !ok {
		log.Println("Missing or invalid 'amount' in event")
		return
	}
	amount := amountFloat

	email, ok := event["email"].(string)
	if !ok {
		log.Println("Missing or invalid 'email' in event")
		return
	}

	var subject, bodyMessage string
	switch eventType {
	case "PaymentCompleted":
		subject = fmt.Sprintf("Оплата заказа %d успешна", orderID)
		bodyMessage = fmt.Sprintf("Заказ %d успешно оплачен! C Вашего счета списано %.2f рублей", orderID, amount)
	case "PaymentFailed":
		subject = fmt.Sprintf("Оплата заказа %d не удалась", orderID)
		bodyMessage = fmt.Sprintf("Не удалось оплатить заказ %d! Недостаточно средств", orderID)
	default:
		log.Printf("Unknown event type: %s", eventType)
		return
	}

	log.Printf("Ready to send message!")
	if err := s.SendEmail(email, subject, bodyMessage); err != nil {
		log.Printf("Failed to send email to %s: %v", email, err)
	} else {
		log.Printf("Notification sent to %s", email)
	}
}
