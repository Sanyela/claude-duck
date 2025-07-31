package utils

import (
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/smtp"
	"strings"

	"claude/config"
)

// GenerateVerificationCode 生成6位数字验证码
func GenerateVerificationCode() string {
	code := ""
	for i := 0; i < 6; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		code += n.String()
	}
	return code
}

// IsAllowedEmailDomain 检查邮箱域名是否在允许列表中
func IsAllowedEmailDomain(email string) bool {
	domain := strings.Split(email, "@")
	if len(domain) != 2 {
		return false
	}

	emailDomain := domain[1]
	for _, allowedDomain := range config.AppConfig.AllowedEmailDomains {
		if emailDomain == allowedDomain {
			return true
		}
	}
	return false
}

// SendVerificationEmail 发送验证码邮件
func SendVerificationEmail(to, code, emailType string) error {
	from := config.AppConfig.SMTPFrom
	password := config.AppConfig.SMTPPassword
	host := config.AppConfig.SMTPHost
	port := config.AppConfig.SMTPPort

	log.Printf("SMTP配置 - Host: %s, Port: %s, User: %s, From: %s", host, port, config.AppConfig.SMTPUser, from)

	// 根据类型设置邮件标题和内容
	var subject, body string
	appName := config.AppConfig.AppName
	switch emailType {
	case "register":
		subject = appName + " 注册验证码"
		body = fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>账户验证</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px; border: 1px solid #ddd;">
        <div style="text-align: center; margin-bottom: 30px;">
            <h1 style="color: #007bff;">%s</h1>
            <h2 style="color: #666;">账户注册验证</h2>
        </div>
        
        <div style="background-color: #f8f9fa; padding: 20px; border-radius: 5px; margin: 20px 0;">
            <p>尊敬的用户，您好！</p>
            <p>感谢您注册%s服务。为确保账户安全，请使用以下验证码完成注册：</p>
            
            <div style="text-align: center; margin: 30px 0;">
                <span style="font-size: 24px; font-weight: bold; color: #007bff; background-color: #e7f3ff; padding: 10px 20px; border-radius: 5px; letter-spacing: 2px;">%s</span>
            </div>
            
            <p style="color: #666; font-size: 14px;">
                • 验证码有效期：%d分钟<br>
                • 请勿将验证码告诉他人<br>
                • 如非本人操作，请忽略此邮件
            </p>
        </div>
        
        <div style="margin-top: 30px; padding-top: 20px; border-top: 1px solid #eee; text-align: center; color: #999; font-size: 12px;">
            <p>此邮件由系统自动发送，请勿回复。</p>
            <p>%s团队</p>
        </div>
    </div>
</body>
</html>`, appName, appName, code, config.AppConfig.VerificationCodeExpireMinutes, appName)
	case "login":
		subject = appName + " 登录验证码"
		body = fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>登录验证</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px; border: 1px solid #ddd;">
        <div style="text-align: center; margin-bottom: 30px;">
            <h1 style="color: #007bff;">%s</h1>
            <h2 style="color: #666;">账户登录验证</h2>
        </div>
        
        <div style="background-color: #f8f9fa; padding: 20px; border-radius: 5px; margin: 20px 0;">
            <p>尊敬的用户，您好！</p>
            <p>您正在尝试登录%s服务。请使用以下验证码完成登录验证：</p>
            
            <div style="text-align: center; margin: 30px 0;">
                <span style="font-size: 24px; font-weight: bold; color: #28a745; background-color: #e8f5e8; padding: 10px 20px; border-radius: 5px; letter-spacing: 2px;">%s</span>
            </div>
            
            <p style="color: #666; font-size: 14px;">
                • 验证码有效期：%d分钟<br>
                • 请勿将验证码告诉他人<br>
                • 如非本人操作，请立即检查账户安全
            </p>
        </div>
        
        <div style="margin-top: 30px; padding-top: 20px; border-top: 1px solid #eee; text-align: center; color: #999; font-size: 12px;">
            <p>此邮件由系统自动发送，请勿回复。</p>
            <p>%s团队</p>
        </div>
    </div>
</body>
</html>`, appName, appName, code, config.AppConfig.VerificationCodeExpireMinutes, appName)
	default:
		return fmt.Errorf("unsupported email type: %s", emailType)
	}

	// 构建邮件消息，使用更简洁的头部
	headers := fmt.Sprintf(`From: %s
To: %s
Subject: %s
MIME-Version: 1.0
Content-Type: text/html; charset=UTF-8
Content-Transfer-Encoding: 8bit

%s`, from, to, subject, body)

	msg := []byte(headers)

	// 根据端口选择连接方式
	switch port {
	case "465":
		// 使用SSL连接
		return sendMailWithTLS(host, port, config.AppConfig.SMTPUser, password, from, []string{to}, msg)
	case "587":
		// 使用STARTTLS连接
		return sendMailWithSTARTTLS(host, port, config.AppConfig.SMTPUser, password, from, []string{to}, msg)
	default:
		// 使用标准SMTP连接（25端口）
		return sendMailStandard(host, port, config.AppConfig.SMTPUser, password, from, []string{to}, msg)
	}
}

// SendSettingsVerificationEmail 发送设置相关验证码邮件
func SendSettingsVerificationEmail(to, code, emailType string) error {
	from := config.AppConfig.SMTPFrom
	password := config.AppConfig.SMTPPassword
	host := config.AppConfig.SMTPHost
	port := config.AppConfig.SMTPPort

	log.Printf("SMTP配置 - Host: %s, Port: %s, User: %s, From: %s", host, port, config.AppConfig.SMTPUser, from)

	// 根据类型设置邮件标题和内容
	var subject, body string
	appName := config.AppConfig.AppName
	switch emailType {
	case "change_username":
		subject = appName + " 修改用户名验证码"
		body = fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>修改用户名验证</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px; border: 1px solid #ddd;">
        <div style="text-align: center; margin-bottom: 30px;">
            <h1 style="color: #007bff;">%s</h1>
            <h2 style="color: #666;">修改用户名验证</h2>
        </div>
        
        <div style="background-color: #f8f9fa; padding: 20px; border-radius: 5px; margin: 20px 0;">
            <p>尊敬的用户，您好！</p>
            <p>您正在尝试修改%s的用户名。为了保护您的账户安全，请使用以下验证码完成操作：</p>
            
            <div style="text-align: center; margin: 30px 0;">
                <span style="font-size: 24px; font-weight: bold; color: #17a2b8; background-color: #e1f7fa; padding: 10px 20px; border-radius: 5px; letter-spacing: 2px;">%s</span>
            </div>
            
            <p style="color: #666; font-size: 14px;">
                • 验证码有效期：%d分钟<br>
                • 请勿将验证码告诉他人<br>
                • 如非本人操作，请立即检查账户安全
            </p>
        </div>
        
        <div style="margin-top: 30px; padding-top: 20px; border-top: 1px solid #eee; text-align: center; color: #999; font-size: 12px;">
            <p>此邮件由系统自动发送，请勿回复。</p>
            <p>%s团队</p>
        </div>
    </div>
</body>
</html>`, appName, appName, code, config.AppConfig.VerificationCodeExpireMinutes, appName)
	case "change_password":
		subject = appName + " 修改密码验证码"
		body = fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>修改密码验证</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px; border: 1px solid #ddd;">
        <div style="text-align: center; margin-bottom: 30px;">
            <h1 style="color: #007bff;">%s</h1>
            <h2 style="color: #666;">修改密码验证</h2>
        </div>
        
        <div style="background-color: #f8f9fa; padding: 20px; border-radius: 5px; margin: 20px 0;">
            <p>尊敬的用户，您好！</p>
            <p>您正在尝试修改%s的登录密码。为了保护您的账户安全，请使用以下验证码完成操作：</p>
            
            <div style="text-align: center; margin: 30px 0;">
                <span style="font-size: 24px; font-weight: bold; color: #dc3545; background-color: #f8d7da; padding: 10px 20px; border-radius: 5px; letter-spacing: 2px;">%s</span>
            </div>
            
            <p style="color: #666; font-size: 14px;">
                • 验证码有效期：%d分钟<br>
                • 请勿将验证码告诉他人<br>
                • 如非本人操作，请立即检查账户安全并联系客服
            </p>
        </div>
        
        <div style="margin-top: 30px; padding-top: 20px; border-top: 1px solid #eee; text-align: center; color: #999; font-size: 12px;">
            <p>此邮件由系统自动发送，请勿回复。</p>
            <p>%s团队</p>
        </div>
    </div>
</body>
</html>`, appName, appName, code, config.AppConfig.VerificationCodeExpireMinutes, appName)
	default:
		return fmt.Errorf("unsupported email type: %s", emailType)
	}

	// 构建邮件消息，使用更简洁的头部
	headers := fmt.Sprintf(`From: %s
To: %s
Subject: %s
MIME-Version: 1.0
Content-Type: text/html; charset=UTF-8
Content-Transfer-Encoding: 8bit

%s`, from, to, subject, body)

	msg := []byte(headers)

	// 根据端口选择连接方式
	switch port {
	case "465":
		// 使用SSL连接
		return sendMailWithTLS(host, port, config.AppConfig.SMTPUser, password, from, []string{to}, msg)
	case "587":
		// 使用STARTTLS连接
		return sendMailWithSTARTTLS(host, port, config.AppConfig.SMTPUser, password, from, []string{to}, msg)
	default:
		// 使用标准SMTP连接（25端口）
		return sendMailStandard(host, port, config.AppConfig.SMTPUser, password, from, []string{to}, msg)
	}
}

// sendMailStandard 使用标准SMTP发送邮件（适用于25端口）
func sendMailStandard(host, port, username, password, from string, to []string, msg []byte) error {
	auth := smtp.PlainAuth("", username, password, host)
	addr := fmt.Sprintf("%s:%s", host, port)

	log.Printf("使用标准SMTP发送邮件到: %s:%s", host, port)

	err := smtp.SendMail(addr, auth, from, to, msg)
	if err != nil {
		log.Printf("标准SMTP发送失败: %v", err)
		return fmt.Errorf("failed to send email via standard SMTP: %v", err)
	}

	log.Printf("邮件发送成功（标准SMTP）: %v", to)
	return nil
}

// sendMailWithSTARTTLS 使用STARTTLS发送邮件（适用于587端口）
func sendMailWithSTARTTLS(host, port, username, password, from string, to []string, msg []byte) error {
	// 连接到服务器
	conn, err := net.Dial("tcp", host+":"+port)
	if err != nil {
		log.Printf("TCP连接失败: %v", err)
		return fmt.Errorf("failed to connect: %v", err)
	}
	defer conn.Close()

	// 创建SMTP客户端
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		log.Printf("SMTP客户端创建失败: %v", err)
		return fmt.Errorf("failed to create SMTP client: %v", err)
	}
	defer client.Quit()

	// 启动TLS
	tlsConfig := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: false,
	}

	if err = client.StartTLS(tlsConfig); err != nil {
		log.Printf("STARTTLS失败: %v", err)
		return fmt.Errorf("failed to start TLS: %v", err)
	}

	// 认证
	auth := smtp.PlainAuth("", username, password, host)
	if err = client.Auth(auth); err != nil {
		log.Printf("SMTP认证失败: %v", err)
		return fmt.Errorf("SMTP authentication failed: %v", err)
	}

	// 发送邮件
	if err = client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %v", err)
	}

	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return fmt.Errorf("failed to set recipient %s: %v", addr, err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to send email data: %v", err)
	}

	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("failed to write email content: %v", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close email writer: %v", err)
	}

	log.Printf("邮件发送成功（STARTTLS）: %v", to)
	return nil
}

// sendMailWithTLS 使用TLS发送邮件（适用于465端口）
func sendMailWithTLS(host, port, username, password, from string, to []string, msg []byte) error {
	// 建立TLS连接
	tlsConfig := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: false, // 在生产环境中应该设为false
	}

	conn, err := tls.Dial("tcp", host+":"+port, tlsConfig)
	if err != nil {
		log.Printf("TLS dial error: %v", err)
		return fmt.Errorf("failed to connect via TLS: %v", err)
	}
	defer conn.Close()

	// 创建SMTP客户端
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		log.Printf("SMTP client creation error: %v", err)
		return fmt.Errorf("failed to create SMTP client: %v", err)
	}
	defer client.Quit()

	// 认证
	auth := smtp.PlainAuth("", username, password, host)
	if err = client.Auth(auth); err != nil {
		log.Printf("SMTP auth error: %v", err)
		return fmt.Errorf("SMTP authentication failed: %v", err)
	}

	// 设置发件人
	if err = client.Mail(from); err != nil {
		log.Printf("SMTP Mail error: %v", err)
		return fmt.Errorf("failed to set sender: %v", err)
	}

	// 设置收件人
	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			log.Printf("SMTP Rcpt error for %s: %v", addr, err)
			return fmt.Errorf("failed to set recipient %s: %v", addr, err)
		}
	}

	// 发送邮件内容
	w, err := client.Data()
	if err != nil {
		log.Printf("SMTP Data error: %v", err)
		return fmt.Errorf("failed to send email data: %v", err)
	}

	_, err = w.Write(msg)
	if err != nil {
		log.Printf("SMTP Write error: %v", err)
		return fmt.Errorf("failed to write email content: %v", err)
	}

	err = w.Close()
	if err != nil {
		log.Printf("SMTP Close error: %v", err)
		return fmt.Errorf("failed to close email writer: %v", err)
	}

	log.Printf("Email sent successfully via TLS to: %v", to)
	return nil
}