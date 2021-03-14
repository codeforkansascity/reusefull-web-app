package main

import (
	"bytes"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ses"
)

const (
	Sender  = "Re.Use.Full <hello@reusefull.org>"
	CharSet = "UTF-8"
)

type EmailArgs struct {
	Link      string
	Button    string
	Message   string
	Preheader string
}

//
// func sendLoginEmail(email, link string) error {
// 	var b bytes.Buffer
// 	err := tmpl.Execute(&b, EmailArgs{
// 		Link:      link,
// 		Button:    "Login",
// 		Message:   "Please click the following button to login.",
// 		Preheader: "Your login request",
// 	})
// 	if err != nil {
// 		return err
// 	}
//
// 	return sendEmail(email, "Login to Hyprcubd", b.String())
// }

// func sendInviteEmail(email, org, link string) error {
// 	var b bytes.Buffer
// 	err := tmpl.Execute(&b, EmailArgs{
// 		Link:      link,
// 		Button:    "Create Account",
// 		Message:   fmt.Sprintf("You have been invited to join %s within Hyprcubd. Please click the following button to setup your account.", org),
// 		Preheader: "You have been invited to Hyprcubd",
// 	})
// 	if err != nil {
// 		return err
// 	}
//
// 	return sendEmail(email, "Welcome to Hyprcubd!", b.String())
// }

func sendNewAccountEmail(email, link string) error {
	var b bytes.Buffer
	err := t.ExecuteTemplate(&b, "email.tmpl", EmailArgs{
		Link:      link,
		Button:    "Complete Registration",
		Message:   "Your account has been created and we are in the process of reviewing it. Please click the following button to verify your email and set your password.",
		Preheader: "New account",
	})
	if err != nil {
		return err
	}

	return sendEmail(email, "Re.Use.Full Account Registration", b.String())
}

/* 
	Notifies admin of new organizations pending approval.
*/
func sendAdminNotificationEmail(orgName string) error {
	notificationMessage := fmt.Sprintf("%s has completed registration, and is ready for approval. Please log in and visit https://app.reusefull.org/admin to review their listing.", orgName)
	return sendEmail("leslie@reusefull.org", "New Organization Registered", notificationMessage)
}

func sendEmail(email, subject, body string) error {
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: []*string{
				aws.String(email),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(body),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(CharSet),
				Data:    aws.String(subject),
			},
		},
		Source: aws.String(Sender),
	}

	_, err := sesSvc.SendEmail(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				fmt.Println(ses.ErrCodeMessageRejected, aerr.Error())
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				fmt.Println(ses.ErrCodeMailFromDomainNotVerifiedException, aerr.Error())
			case ses.ErrCodeConfigurationSetDoesNotExistException:
				fmt.Println(ses.ErrCodeConfigurationSetDoesNotExistException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}

		return err
	}
	return nil
}
