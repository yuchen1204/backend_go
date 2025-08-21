package service

import (
	"backend/internal/config"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"html"
	"log"
	"time"

	"gopkg.in/gomail.v2"
)

// MailService 邮件服务接口
type MailService interface {
	SendVerificationCode(to, code string) error
	SendResetPasswordCode(to, code string) error
	SendDeviceVerificationCode(to, code, deviceName, ip, ua string) error
	// 收到好友请求通知（发送给接收方）
	SendFriendRequestNotification(to, requesterName, requesterUsername, receiverName string, note string, requestCreatedAt time.Time) error
	// 好友请求结果通知（发送给另一方）result: accepted/rejected/cancelled
	SendFriendRequestResultNotification(to, otherPartyName, otherPartyUsername, result string, requestCreatedAt, handledAt time.Time) error
}

// SendFriendRequestNotification 收到好友请求通知
func (s *smtpMailService) SendFriendRequestNotification(to, requesterName, requesterUsername, receiverName string, note string, requestCreatedAt time.Time) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "您收到一条新的好友请求")

	body := fmt.Sprintf(`
        <p>您好 %s,</p>
        <p>您收到了来自 <b>%s</b> (@%s) 的好友请求。</p>
        <p>对方的附言：%s</p>
        <p>请求时间：%s</p>
        <p>请前往应用内的“好友-入站请求”页面进行处理。</p>
    `, html.EscapeString(receiverName), html.EscapeString(requesterName), html.EscapeString(requesterUsername), html.EscapeString(note), requestCreatedAt.Local().Format("2006-01-02 15:04:05"))
	m.SetBody("text/html", body)

	log.Printf("准备发送好友请求通知: to=%s from=%s requester=%s(@%s)", to, s.from, requesterName, requesterUsername)
	if err := s.dialer.DialAndSend(m); err != nil {
		log.Printf("发送好友请求通知失败: host=%s port=%d username=%s to=%s err=%v", s.dialer.Host, s.dialer.Port, s.dialer.Username, to, err)
		return fmt.Errorf("发送好友请求通知失败(host=%s port=%d user=%s to=%s): %w", s.dialer.Host, s.dialer.Port, s.dialer.Username, to, err)
	}
	log.Printf("发送好友请求通知成功: to=%s", to)
	return nil
}

// SendFriendRequestResultNotification 好友请求结果通知
func (s *smtpMailService) SendFriendRequestResultNotification(to, otherPartyName, otherPartyUsername, result string, requestCreatedAt, handledAt time.Time) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", to)
	var subject string
	switch result {
	case "accepted":
		subject = "您的好友请求已被接受"
	case "rejected":
		subject = "您的好友请求已被拒绝"
	case "cancelled":
		subject = "好友请求已撤回"
	default:
		subject = "好友请求状态更新"
	}
	m.SetHeader("Subject", subject)

	body := fmt.Sprintf(`
        <p>您好,</p>
        <p>与 <b>%s</b> (@%s) 的好友请求状态：<b>%s</b></p>
        <p>请求时间：%s</p>
        <p>处理时间：%s</p>
    `,
		html.EscapeString(otherPartyName),
		html.EscapeString(otherPartyUsername),
		html.EscapeString(result),
		requestCreatedAt.Local().Format("2006-01-02 15:04:05"),
		handledAt.Local().Format("2006-01-02 15:04:05"),
	)
	m.SetBody("text/html", body)

	log.Printf("准备发送好友请求结果通知: to=%s result=%s other=%s(@%s)", to, result, otherPartyName, otherPartyUsername)
	if err := s.dialer.DialAndSend(m); err != nil {
		log.Printf("发送好友请求结果通知失败: host=%s port=%d username=%s to=%s err=%v", s.dialer.Host, s.dialer.Port, s.dialer.Username, to, err)
		return fmt.Errorf("发送好友请求结果通知失败(host=%s port=%d user=%s to=%s): %w", s.dialer.Host, s.dialer.Port, s.dialer.Username, to, err)
	}
	log.Printf("发送好友请求结果通知成功: to=%s", to)
	return nil
}

// SendDeviceVerificationCode 发送设备验证验证码邮件
func (s *smtpMailService) SendDeviceVerificationCode(to, code, deviceName, ip, ua string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "设备登录验证验证码")

	body := fmt.Sprintf(`
        <p>您好,</p>
        <p>我们检测到一个新的设备正在尝试登录您的账号，需要进行二次验证。</p>
        <p>设备名称：<b>%s</b></p>
        <p>来源IP：<b>%s</b></p>
        <p>User-Agent：<b>%s</b></p>
        <p>您的设备验证码是：<b>%s</b></p>
        <p>此验证码将在5分钟后失效。</p>
        <p>如果非您本人操作，请尽快修改密码并检查账号安全。</p>
    `, deviceName, ip, ua, code)
    m.SetBody("text/html", body)

    log.Printf("准备发送设备验证验证码邮件: to=%s from=%s", to, s.from)
    if err := s.dialer.DialAndSend(m); err != nil {
        log.Printf("发送设备验证验证码邮件失败: host=%s port=%d username=%s to=%s err=%v", s.dialer.Host, s.dialer.Port, s.dialer.Username, to, err)
        return fmt.Errorf("发送设备验证邮件失败(host=%s port=%d user=%s to=%s): %w", s.dialer.Host, s.dialer.Port, s.dialer.Username, to, err)
    }
    log.Printf("发送设备验证验证码邮件成功: to=%s", to)

    return nil
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
	// 配置 TLS：显式设置 ServerName，并尽力加载系统根证书
	var rootCAs *x509.CertPool
	if pool, err := x509.SystemCertPool(); err != nil {
		log.Printf("加载系统根证书失败: %v", err)
	} else {
		rootCAs = pool
	}
	dialer.TLSConfig = &tls.Config{
		ServerName: cfg.Host,
		RootCAs:    rootCAs, // 若为nil，Go会回退到默认；安装了 ca-certificates 后应能正常读取
	}
	// 启动时打印SMTP关键信息（不打印密码）
	log.Printf("初始化SMTP拨号器: host=%s port=%d username=%s from=%s password_set=%t", cfg.Host, cfg.Port, cfg.Username, cfg.From, cfg.Password != "")
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
	log.Printf("准备发送注册验证码邮件: to=%s from=%s", to, s.from)
	if err := s.dialer.DialAndSend(m); err != nil {
		log.Printf("发送注册验证码邮件失败: host=%s port=%d username=%s to=%s err=%v", s.dialer.Host, s.dialer.Port, s.dialer.Username, to, err)
		return fmt.Errorf("发送邮件失败(host=%s port=%d user=%s to=%s): %w", s.dialer.Host, s.dialer.Port, s.dialer.Username, to, err)
	}
	log.Printf("发送注册验证码邮件成功: to=%s", to)

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
	log.Printf("准备发送重置密码验证码邮件: to=%s from=%s", to, s.from)
	if err := s.dialer.DialAndSend(m); err != nil {
		log.Printf("发送重置密码验证码邮件失败: host=%s port=%d username=%s to=%s err=%v", s.dialer.Host, s.dialer.Port, s.dialer.Username, to, err)
		return fmt.Errorf("发送重置密码邮件失败(host=%s port=%d user=%s to=%s): %w", s.dialer.Host, s.dialer.Port, s.dialer.Username, to, err)
	}
	log.Printf("发送重置密码验证码邮件成功: to=%s", to)

	return nil
}