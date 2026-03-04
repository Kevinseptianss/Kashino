package mail

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
)

type MailService struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func NewMailService() *MailService {
	return &MailService{
		host:     os.Getenv("SMTP_HOST"),
		port:     os.Getenv("SMTP_PORT"),
		username: os.Getenv("SMTP_USER"),
		password: os.Getenv("SMTP_PASS"),
		from:     os.Getenv("SMTP_FROM"),
	}
}

func (s *MailService) getDefaults() (string, string, string, string, string) {
	host := s.host
	if host == "" {
		host = "smtp-relay.brevo.com"
	}
	port := s.port
	if port == "" {
		port = "587"
	}
	username := s.username
	password := s.password
	from := s.from
	if from == "" {
		from = username
	}
	return host, port, username, password, from
}

func (s *MailService) SendEmail(to string, subject, body string, isHTML bool) error {
	host, port, username, password, from := s.getDefaults()
	if username == "" || password == "" {
		return fmt.Errorf("SMTP credentials not configured")
	}

	contentType := "text/plain"
	if isHTML {
		contentType = "text/html"
	}

	header := make(map[string]string)
	header["From"] = from
	header["To"] = to
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = contentType + "; charset=\"utf-8\""

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	auth := smtp.PlainAuth("", username, password, host)

	// Port 465 is for direct SSL/TLS
	if port == "465" {
		tlsconfig := &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         host,
		}

		conn, err := tls.Dial("tcp", host+":"+port, tlsconfig)
		if err != nil {
			return fmt.Errorf("failed to dial tls: %v", err)
		}
		defer conn.Close()

		c, err := smtp.NewClient(conn, host)
		if err != nil {
			return fmt.Errorf("failed to create smtp client: %v", err)
		}
		defer c.Quit()

		if err = c.Auth(auth); err != nil {
			return fmt.Errorf("failed to authenticate: %v", err)
		}

		return s.send(c, from, to, message)
	}

	// Port 587 or 25 use STARTTLS
	c, err := smtp.Dial(host + ":" + port)
	if err != nil {
		return fmt.Errorf("failed to dial smtp: %v", err)
	}
	defer c.Quit()

	tlsconfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         host,
	}

	if err = c.StartTLS(tlsconfig); err != nil {
		return fmt.Errorf("failed to start tls: %v", err)
	}

	if err = c.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate: %v", err)
	}

	return s.send(c, from, to, message)
}

func (s *MailService) send(c *smtp.Client, from, to, msg string) error {
	if err := c.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %v", err)
	}
	if err := c.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set recipient: %v", err)
	}

	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %v", err)
	}

	_, err = w.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("failed to write message: %v", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %v", err)
	}

	return nil
}

func (s *MailService) SendOTP(to string, otp string) error {
	subject := "Kashino - Verification Code"
	htmlTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: 'Inter', -apple-system, sans-serif; background-color: #0a0a0f; margin: 0; padding: 0; color: #ffffff; }
        .container { max-width: 600px; margin: 40px auto; padding: 40px; background: #14141d; border-radius: 24px; border: 1px solid #2a2a35; box-shadow: 0 20px 40px rgba(0,0,0,0.4); }
        .logo { font-size: 28px; font-weight: 800; color: #9670dd; text-align: center; margin-bottom: 30px; letter-spacing: -1px; }
        .title { font-size: 24px; font-weight: 700; text-align: center; margin-bottom: 10px; color: #ffffff; }
        .subtitle { font-size: 16px; color: #8e8ea0; text-align: center; margin-bottom: 40px; }
        .otp-card { background: #1c1c27; border-radius: 16px; padding: 30px; text-align: center; border: 1px solid #333344; margin-bottom: 40px; }
        .otp-code { font-size: 42px; font-weight: 800; color: #9670dd; letter-spacing: 8px; margin: 0; }
        .expiry { font-size: 14px; color: #666677; margin-top: 15px; }
        .footer { text-align: center; font-size: 12px; color: #444455; line-height: 1.6; }
        .footer a { color: #9670dd; text-decoration: none; }
    </style>
</head>
<body>
    <div class="container">
        <div class="logo">KASHINO</div>
        <div class="title">Verify Your Email</div>
        <div class="subtitle">Use the verification code below to complete your signup.</div>
        
        <div class="otp-card">
            <h1 class="otp-code">%s</h1>
            <div class="expiry">Valid for 10 minutes</div>
        </div>
        
        <div class="footer">
            If you didn't request this email, you can safely ignore it.<br>
            &copy; 2026 Kashino. All rights reserved.<br>
        </div>
    </div>
</body>
</html>
`
	body := fmt.Sprintf(htmlTemplate, otp)
	return s.SendEmail(to, subject, body, true)
}

func (s *MailService) SendResetPassword(to string, token string) error {
	subject := "Kashino - Password Reset"
	htmlTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: 'Inter', -apple-system, sans-serif; background-color: #0a0a0f; margin: 0; padding: 0; color: #ffffff; }
        .container { max-width: 600px; margin: 40px auto; padding: 40px; background: #14141d; border-radius: 24px; border: 1px solid #2a2a35; }
        .logo { font-size: 28px; font-weight: 800; color: #9670dd; text-align: center; margin-bottom: 30px; }
        .otp-card { background: #1c1c27; border-radius: 16px; padding: 30px; text-align: center; border: 1px solid #333344; margin-bottom: 40px; }
        .btn { display: inline-block; padding: 14px 28px; background: #9670dd; color: #ffffff; text-decoration: none; border-radius: 12px; font-weight: 600; margin-top: 20px; }
        .footer { text-align: center; font-size: 12px; color: #444455; }
    </style>
</head>
<body>
    <div class="container">
        <div class="logo">KASHINO</div>
        <div style="text-align: center;">
            <h2>Reset Your Password</h2>
            <p style="color: #8e8ea0;">We received a request to reset your password. Use the token below to proceed:</p>
            <div class="otp-card">
                <code style="font-size: 24px; color: #9670dd;">%s</code>
            </div>
            <p style="font-size: 14px; color: #666677;">This token expires in 1 hour.</p>
        </div>
        <div class="footer">
            &copy; 2026 Kashino.
        </div>
    </div>
</body>
</html>
`
	body := fmt.Sprintf(htmlTemplate, token)
	return s.SendEmail(to, subject, body, true)
}
