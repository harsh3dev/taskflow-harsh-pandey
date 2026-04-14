package email

import (
	"bytes"
	"fmt"
	"net/smtp"
	"text/template"

	"github.com/harshpn/taskflow/internal/events"
)

// SMTPSender sends transactional emails via a plain SMTP server using
// stdlib net/smtp. No external dependency required.
type SMTPSender struct {
	host     string
	port     int
	username string
	password string
	from     string
}

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

func NewSMTPSender(cfg Config) *SMTPSender {
	return &SMTPSender{
		host:     cfg.Host,
		port:     cfg.Port,
		username: cfg.Username,
		password: cfg.Password,
		from:     cfg.From,
	}
}

// Send composes and delivers a notification email for the given event to toAddr.
func (s *SMTPSender) Send(toName, toAddr string, event events.TaskChangedEvent) error {
	subject, body, err := renderEmail(toName, event)
	if err != nil {
		return fmt.Errorf("render email: %w", err)
	}

	msg := buildMessage(s.from, toAddr, subject, body)
	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	var auth smtp.Auth
	if s.username != "" {
		auth = smtp.PlainAuth("", s.username, s.password, s.host)
	}

	if err := smtp.SendMail(addr, auth, s.from, []string{toAddr}, msg); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	return nil
}

func buildMessage(from, to, subject, body string) []byte {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "From: %s\r\n", from)
	fmt.Fprintf(&buf, "To: %s\r\n", to)
	fmt.Fprintf(&buf, "Subject: %s\r\n", subject)
	fmt.Fprintf(&buf, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&buf, "Content-Type: text/plain; charset=UTF-8\r\n")
	fmt.Fprintf(&buf, "\r\n")
	fmt.Fprintf(&buf, "%s\r\n", body)
	return buf.Bytes()
}

var subjectTemplates = map[events.ChangeKind]string{
	events.ChangeKindStatus:   `[TaskFlow] Task "{{.TaskTitle}}" status changed to {{.NewValue}}`,
	events.ChangeKindPriority: `[TaskFlow] Task "{{.TaskTitle}}" priority changed to {{.NewValue}}`,
	events.ChangeKindAssignee: `[TaskFlow] You have been assigned to "{{.TaskTitle}}"`,
	events.ChangeKindDueDate:  `[TaskFlow] Task "{{.TaskTitle}}" due date updated`,
}

var bodyTemplate = template.Must(template.New("body").Parse(`Hi {{.RecipientName}},

{{.ChangeDescription}}

Task: {{.TaskTitle}}
{{- if .OldValue}}
Previous value: {{.OldValue}}{{end}}
New value:      {{.NewValue}}

This is an automated notification from TaskFlow.
`))

type bodyData struct {
	RecipientName     string
	TaskTitle         string
	ChangeDescription string
	OldValue          string
	NewValue          string
}

func changeDescription(kind events.ChangeKind, taskTitle string) string {
	switch kind {
	case events.ChangeKindStatus:
		return fmt.Sprintf(`The status of task "%s" has been updated.`, taskTitle)
	case events.ChangeKindPriority:
		return fmt.Sprintf(`The priority of task "%s" has been changed.`, taskTitle)
	case events.ChangeKindAssignee:
		return fmt.Sprintf(`You have been assigned to task "%s".`, taskTitle)
	case events.ChangeKindDueDate:
		return fmt.Sprintf(`The due date for task "%s" has been updated.`, taskTitle)
	default:
		return fmt.Sprintf(`Task "%s" has been updated.`, taskTitle)
	}
}

func renderEmail(recipientName string, event events.TaskChangedEvent) (subject, body string, err error) {
	subjectTmplStr, ok := subjectTemplates[event.ChangeKind]
	if !ok {
		subjectTmplStr = `[TaskFlow] Task "{{.TaskTitle}}" updated`
	}
	subjectTmpl, err := template.New("subject").Parse(subjectTmplStr)
	if err != nil {
		return "", "", err
	}
	var subjectBuf bytes.Buffer
	if err = subjectTmpl.Execute(&subjectBuf, event); err != nil {
		return "", "", err
	}

	var bodyBuf bytes.Buffer
	if err = bodyTemplate.Execute(&bodyBuf, bodyData{
		RecipientName:     recipientName,
		TaskTitle:         event.TaskTitle,
		ChangeDescription: changeDescription(event.ChangeKind, event.TaskTitle),
		OldValue:          event.OldValue,
		NewValue:          event.NewValue,
	}); err != nil {
		return "", "", err
	}

	return subjectBuf.String(), bodyBuf.String(), nil
}
