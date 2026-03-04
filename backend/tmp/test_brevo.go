package main

import (
	"fmt"
	"kashino-backend/internal/mail"
	"log"
	"os"
)

func main() {
	// For testing, we can override env vars directly
	os.Setenv("SMTP_HOST", "smtp-relay.brevo.com")
	os.Setenv("SMTP_PORT", "587")
	os.Setenv("SMTP_USER", "a3ec72001@smtp-brevo.com")
	os.Setenv("SMTP_PASS", "YOUR_SMTP_PASS")
	os.Setenv("SMTP_FROM", "admin@kashino.my.id")

	svc := mail.NewMailService()
	testRecipient := "kevinseptiansaputra@gmail.com"

	fmt.Printf("Testing Brevo SMTP with recipient: %s...\n", testRecipient)

	err := svc.SendOTP(testRecipient, "888999")
	if err != nil {
		log.Fatalf("FAILED: %v", err)
	}

	fmt.Println("SUCCESS: Test email sent via Brevo!")
}
