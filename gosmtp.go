package smtp

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"mime/multipart"
	"net/smtp"
	"net/textproto"
	"strings"
)

type Auth struct {
	Host string
	Port string
	Pass string
}

func Client(host string, port string, password string) *Auth {
	return &Auth{
		Host: host,
		Port: port,
		Pass: password,
	}
}

func (s *Auth) Send(
	from string,
	to []string,
	subject string,
	body string,
	attachments ...map[string]*bytes.Buffer,
) error {
	attachmentMap := make(map[string]*bytes.Buffer)
	if len(attachments) > 0 {
		attachmentMap = attachments[0]
	}

	message, err := s.message(from, to, subject, body, attachmentMap)
	if err != nil {
		return fmt.Errorf("failed to create email message: %w", err)
	}

	auth := smtp.PlainAuth("", from, s.Pass, s.Host)

	err = smtp.SendMail(s.Host+":"+s.Port, auth, from, to, message)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (s *Auth) Validate(user string) error {
	address := s.Host + ":" + s.Port

	if s.Port == "465" {
		tlsConfig := &tls.Config{
			ServerName: s.Host,
		}

		conn, err := tls.Dial("tcp", address, tlsConfig)
		if err != nil {
			return fmt.Errorf("failed to connect to SMTPS server: %w", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, s.Host)
		if err != nil {
			return fmt.Errorf("failed to create SMTPS client: %w", err)
		}
		defer client.Close()

		auth := smtp.PlainAuth("", user, s.Pass, s.Host)
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}

		return nil
	}

	client, err := smtp.Dial(address)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsConfig := &tls.Config{
			ServerName: s.Host,
		}

		if err = client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("failed to start TLS: %w", err)
		}
	} else {
		return fmt.Errorf("TLS not supported by the server, aborted")
	}

	auth := smtp.PlainAuth("", user, s.Pass, s.Host)
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	return nil
}

func (s *Auth) message(
	from string,
	to []string,
	subject string,
	body string,
	attachments map[string]*bytes.Buffer,
) ([]byte, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	buf.WriteString(fmt.Sprintf("From: %s\n", from))
	buf.WriteString(fmt.Sprintf("To: %s\n", strings.Join(to, ",")))
	buf.WriteString(fmt.Sprintf("Subject: %s\n", subject))
	buf.WriteString("MIME-Version: 1.0\n")
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\n\n", writer.Boundary()))

	if err := s.write(writer, "text/plain", body); err != nil {
		return nil, fmt.Errorf("failed to write email body: %w", err)
	}

	for filename, fileBuffer := range attachments {
		if err := s.attach(writer, filename, fileBuffer); err != nil {
			return nil, fmt.Errorf("failed to attach file %s: %w", filename, err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	return buf.Bytes(), nil
}

func (s *Auth) write(writer *multipart.Writer, contentType, content string) error {
	partHeader := make(textproto.MIMEHeader)
	partHeader.Set("Content-Type", contentType)
	part, err := writer.CreatePart(partHeader)
	if err != nil {
		return fmt.Errorf("failed to create part: %w", err)
	}
	_, err = part.Write([]byte(content))
	return err
}

func (s *Auth) attach(writer *multipart.Writer, filename string, fileBuffer *bytes.Buffer) error {
	attachmentHeader := make(textproto.MIMEHeader)
	attachmentHeader.Set("Content-Type", "application/octet-stream")
	attachmentHeader.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	attachment, err := writer.CreatePart(attachmentHeader)
	if err != nil {
		return fmt.Errorf("failed to create attachment part: %w", err)
	}
	_, err = attachment.Write(fileBuffer.Bytes())
	return err
}
