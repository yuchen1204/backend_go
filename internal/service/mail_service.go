package service

import (
	"backend/internal/config"
	"fmt"

	"gopkg.in/gomail.v2"
)

// MailService 邮件服务接口
type MailService interface {
	SendVerificationCode(to, code string) error
	SendResetPasswordCode(to, code string) error
}

// smtpMailService SMTP邮件服务实现
type smtpMailService struct {
	dialer *gomail.Dialer
	from   string
}

// NewMailService 创建邮件服务实例
func NewMailService(cfg *config.SMTPConfig) MailService {
	// 创建一个拨号器，用于连接SMTP服务器
	dialer := gomail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)
	return &smtpMailService{
		dialer: dialer,
		from:   cfg.From,
	}
}

// SendVerificationCode 发送验证码邮件
func (s *smtpMailService) SendVerificationCode(to, code string) error {
	// 创建邮件消息
	m := gomail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "您的注册验证码")
	
	// 设置邮件正文
	body := fmt.Sprintf(`
		<p>您好,</p>
		<p>感谢您注册！您的验证码是：<b>%s</b></p>
		<p>此验证码将在5分钟后失效。</p>
		<p>如果您没有请求此验证码，请忽略此邮件。</p>
	`, code)
	m.SetBody("text/html", body)

	// 发送邮件
	if err := s.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("发送邮件失败: %w", err)
	}

	return nil
}

// SendResetPasswordCode 发送重置密码验证码邮件
func (s *smtpMailService) SendResetPasswordCode(to, code string) error {
	// 创建邮件消息
	m := gomail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "密码重置验证码")
	
	// 设置邮件正文
	body := fmt.Sprintf(`
		<p>您好,</p>
		<p>您申请重置密码。您的验证码是：<b>%s</b></p>
		<p>此验证码将在5分钟后失效。</p>
		<p>如果您没有申请重置密码，请忽略此邮件并确保您的账户安全。</p>
		<p>为了您的账户安全，请不要将此验证码泄露给他人。</p>
	`, code)
	m.SetBody("text/html", body)

	// 发送邮件
	if err := s.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("发送重置密码邮件失败: %w", err)
	}

	return nil
} 