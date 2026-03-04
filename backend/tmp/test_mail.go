package main_test

import (
	"fmt"
	"kashino-backend/internal/mail"
	"log"
)

func main() {
	svc := mail.NewMailService()
	testEmail := "kevinseptiansaputra@gmail.com" // You can change this to your email for testing

	fmt.Printf("Sending test OTP to %s...\n", testEmail)
	err := svc.SendOTP(testEmail, "123456")
	if err != nil {
		log.Fatalf("Failed to send OTP: %v", err)
	}
	fmt.Println("OTP sent successfully!")

	fmt.Printf("Sending test reset token to %s...\n", testEmail)
	err = svc.SendResetPassword(testEmail, "test-reset-token")
	if err != nil {
		log.Fatalf("Failed to send reset email: %v", err)
	}
	fmt.Println("Reset email sent successfully!")
}
