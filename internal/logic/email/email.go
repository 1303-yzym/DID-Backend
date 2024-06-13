package email

import (
	"context"
	"fmt"
	"github.com/emersion/go-imap"
	id "github.com/emersion/go-imap-id"
	"github.com/emersion/go-imap/client"
	"github.com/gogf/gf-demo-user/v2/internal/model"
	"github.com/gogf/gf-demo-user/v2/internal/service"
	"log"
	"net/smtp"
)

type (
	sEmail struct{}
)

func init() {
	service.RegisterEmail(New())
}

func New() service.IEmail {
	return &sEmail{}
}

func (s sEmail) SendEmail(ctx context.Context, in model.EmailSendInput) (err error) {
	// 设置认证信息。
	auth := smtp.PlainAuth("", "yzym_143307lingyu@163.com", "QWHZYJRZTCPXYXAC", "smtp.163.com")

	// 设置发送的邮件内容。
	message := []byte("To: " + in.To + "\r\n" +
		"Subject: " + in.Subject + "\r\n" +
		"\r\n" +
		in.Body + "\r\n")

	// 发送邮件。
	err = smtp.SendMail("smtp.163.com:25", auth, "yzym_143307lingyu@163.com", []string{in.To}, message)
	if err != nil {
		fmt.Println("Error sending email:", err)
		return
	}
	fmt.Println("Email sent successfully.")
	return nil

	/*
		//由于国内IP会被谷歌服务器屏蔽，因此使用网易的服务
			// SMTP服务器地址和端口号
			smtpHost := "smtp.gmail.com"
			smtpPort := "587"
			// 设置邮件内容
			message := []byte("To: " + in.To + "\r\n" +
				"Subject: " + in.Subject + "\r\n" +
				"\r\n" +
				in.Body + "\r\n")

			// 连接到SMTP服务器
			auth := smtp.PlainAuth("", in.From, in.Password, smtpHost)

			// 发送邮件
			err = smtp.SendMail(smtpHost+":"+smtpPort, auth, in.From, []string{in.To}, message)
			if err != nil {
				fmt.Println("Error sending email:", err)
				return err
			}

			fmt.Println("Email sent successfully.")
			return nil

	*/
}

func (s sEmail) GetEmail(ctx context.Context) []byte {
	log.Println("连接服务器中...")

	c, err := client.DialTLS("imap.163.com:993", nil)
	idClient := id.NewClient(c)
	idClient.ID(
		id.ID{
			id.FieldName:    "IMAPClient",
			id.FieldVersion: "3.1.0",
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("连接成功")
	defer c.Logout()

	// 登录
	if err := c.Login("yzym_143307lingyu@163.com", "QWHZYJRZTCPXYXAC"); err != nil {
		log.Fatal(err)
	}
	log.Println("登陆成功")

	// 邮箱文件夹列表
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", mailboxes)
	}()

	log.Println("邮箱文件夹:")
	for m := range mailboxes {
		log.Println("* " + m.Name)
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	// 选择收件箱
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}
	// 获取最新的邮件
	if mbox.Messages == 0 {
		log.Println("No message in mailbox")
		return nil
	}

	seqset := new(imap.SeqSet)
	seqset.AddRange(uint32(mbox.Messages), uint32(mbox.Messages))

	messages := make(chan *imap.Message, 1)
	done = make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, messages)
	}()

	log.Println("Last message:")
	if msg, ok := <-messages; ok {
		log.Println("* Envelope:", msg.Envelope)
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}
	return nil
}
