package mailer

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"
	"strings"
	"time"

	"devclub.com/identity/internal/api/config"
	"github.com/wneessen/go-mail"
)

type Mailer struct {
	client    *mail.Client
	from      string
	templates map[string]*template.Template
	logger    *slog.Logger
}

func NewMailer(cfg *config.Config, logger *slog.Logger) (*Mailer, error) {

	client, err := mail.NewClient(
		cfg.MailHost,
		mail.WithPort(cfg.MailPort),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(cfg.MailUsername),
		mail.WithPassword(cfg.MailPassword),

		mail.WithTimeout(10*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create mail client: %w", err)
	}

	inviteTpl, err := template.New("invite").Parse(InviteEmailHTML)
	if err != nil {
		return nil, fmt.Errorf("failed to parse invite template: %w", err)
	}

	resetTpl, err := template.New("reset").Parse(ResetPasswordHTML)
	if err != nil {
		return nil, fmt.Errorf("failed to parse reset template: %w", err)
	}

	fromAddr := strings.TrimSpace(cfg.MailFrom)
	if fromAddr == "" {
		fromAddr = cfg.MailUsername
	} else if !strings.Contains(fromAddr, "@") && cfg.MailUsername != "" {
		fromAddr = fmt.Sprintf("%s <%s>", fromAddr, cfg.MailUsername)
	}

	return &Mailer{
		client: client,
		from:   fromAddr,
		templates: map[string]*template.Template{
			"invite": inviteTpl,
			"reset":  resetTpl,
		},
		logger: logger,
	}, nil
}

func (m *Mailer) sendAsync(to, subject, tplName string, data any) {
	go func() {
		tpl, ok := m.templates[tplName]
		if !ok {
			m.logger.Error("email template not found", "template", tplName)
			return
		}

		var body bytes.Buffer
		if err := tpl.Execute(&body, data); err != nil {
			m.logger.Error("failed to execute email template", "error", err, "template", tplName)
			return
		}

		msg := mail.NewMsg()
		if err := msg.From(m.from); err != nil {
			m.logger.Error("failed to set mail from address", "error", err)
			return
		}
		if err := msg.To(to); err != nil {
			m.logger.Error("failed to set mail to address", "error", err)
			return
		}
		msg.Subject(subject)
		msg.SetBodyString(mail.TypeTextHTML, body.String())

		if err := m.client.DialAndSend(msg); err != nil {
			m.logger.Error("failed to deliver email",
				"to", to,
				"template", tplName,
				"error", err,
			)
			return
		}

		m.logger.Info("email sent successfully", "to", to, "template", tplName)
	}()
}

func (m *Mailer) SendUserInvite(to, role, token, frontendUrl string) {
	data := map[string]any{
		"Role":      role,
		"ActionURL": fmt.Sprintf("%s/accept-invite?token=%s", frontendUrl, token),
		"ExpiresIn": "24 hours",
	}
	m.sendAsync(to, "You've been invited to join", "invite", data)
}

// SendPasswordReset sends a password reset email asynchronously
func (m *Mailer) SendPasswordReset(to, token string, frontendUrl string) {
	data := map[string]any{
		"ActionURL": fmt.Sprintf("%s?token=%s", frontendUrl, token),
		"ExpiresIn": "15 minutes",
	}
	m.sendAsync(to, "Password Reset Request", "reset", data)
}
