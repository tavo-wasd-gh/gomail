// Package gomail provides simple emailing abstraction.
package gomail

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/smtp"
	"net/textproto"
	"strings"
)

// Email sends an email with the specified host, port, from address, recipients,
// subject, body, and attachments. It returns an error if the sending fails.
func Email(
	host, port, password string,
	from string, to []string,
	subject, body string,
	attachments map[string]*bytes.Buffer,
) error {
	message := Message(from, to, subject, body, attachments)
	auth := smtp.PlainAuth("", from, password, host)
	err := smtp.SendMail(host+":"+port, auth, from, to, message)
	if err != nil {
		return err
	}

	return nil
}

// Message takes parts of a message and returns the crafted message as []byte.
func Message(
	from string,
	to []string,
	subject, body string,
	attachments map[string]*bytes.Buffer,
) []byte {
	var buf bytes.Buffer

	writer := multipart.NewWriter(&buf)

	buf.WriteString("From: " + from + "\n")
	buf.WriteString("To: " + strings.Join(to, ",") + "\n")
	buf.WriteString(subject)
	buf.WriteString("MIME-Version: 1.0\n")
	buf.WriteString("Content-Type: multipart/mixed; boundary=" + writer.Boundary() + "\n\n")

	write(writer, "text/plain", body)

	for filename, fileBuffer := range attachments {
		attach(writer, filename, fileBuffer)
	}

	writer.Close()
	return buf.Bytes()
}

// write writes the message body or part of an email.
func write(writer *multipart.Writer, contentType, content string) {
	partHeader := make(textproto.MIMEHeader)
	partHeader.Set("Content-Type", contentType)
	part, _ := writer.CreatePart(partHeader)
	part.Write([]byte(content))
}

// attach adds an attachment to the email.
func attach(writer *multipart.Writer, filename string, fileBuffer *bytes.Buffer) {
	attachmentHeader := make(textproto.MIMEHeader)
	attachmentHeader.Set("Content-Type", "application/octet-stream")
	attachmentHeader.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	attachment, _ := writer.CreatePart(attachmentHeader)
	attachment.Write(fileBuffer.Bytes())
}
