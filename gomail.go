package gomail

import (
	"bytes"
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

func Client(host, port, password string) *Auth {
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
	var attachmentMap map[string]*bytes.Buffer
	if len(attachments) > 0 {
		attachmentMap = attachments[0]
	} else {
		attachmentMap = make(map[string]*bytes.Buffer)
	}

	message := s.message(from, to, subject, body, attachmentMap)
	auth := smtp.PlainAuth("", from, s.Pass, s.Host)

	err := smtp.SendMail(s.Host+":"+s.Port, auth, from, to, message)
	if err != nil {
		return err
	}

	return nil
}

func (s *Auth) Validate(user string) error {
	auth := smtp.PlainAuth("", user, s.Pass, s.Host)

	client, err := smtp.Dial(s.Host + ":" + s.Port)
	if err != nil {
		return fmt.Errorf("Failed to connect to SMTP server: %w", err)
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err = client.StartTLS(nil); err != nil {
			return fmt.Errorf("Failed to start TLS: %w", err)
		}
	}

	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("Authentication failed: %w", err)
	}

	return nil
}

func (s *Auth) message(
	from string,
	to []string,
	subject string,
	body string,
	attachments map[string]*bytes.Buffer,
) []byte {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	buf.WriteString("From: " + from + "\n")
	buf.WriteString("To: " + strings.Join(to, ",") + "\n")
	buf.WriteString("Subject: " + subject + "\n")
	buf.WriteString("MIME-Version: 1.0\n")
	buf.WriteString("Content-Type: multipart/mixed; boundary=" + writer.Boundary() + "\n\n")

	s.writePart(writer, "text/plain", body)

	for filename, fileBuffer := range attachments {
		s.attachFile(writer, filename, fileBuffer)
	}

	writer.Close()
	return buf.Bytes()
}

func (s *Auth) writePart(writer *multipart.Writer, contentType, content string) {
	partHeader := make(textproto.MIMEHeader)
	partHeader.Set("Content-Type", contentType)
	part, _ := writer.CreatePart(partHeader)
	part.Write([]byte(content))
}

func (s *Auth) attachFile(writer *multipart.Writer, filename string, fileBuffer *bytes.Buffer) {
	attachmentHeader := make(textproto.MIMEHeader)
	attachmentHeader.Set("Content-Type", "application/octet-stream")
	attachmentHeader.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	attachment, _ := writer.CreatePart(attachmentHeader)
	attachment.Write(fileBuffer.Bytes())
}
