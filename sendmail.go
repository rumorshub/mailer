// The MIT License (MIT) Copyright (c) 2022 - present, Gani Georgiev
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software
// and associated documentation files (the "Software"), to deal in the Software without restriction,
// including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all copies or substantial
// portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
// LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
// WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
// SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package mailer

import (
	"bytes"
	"errors"
	"mime"
	"net/http"
	"os/exec"
	"strings"
)

var _ Mailer = (*SendMail)(nil)

// SendMail implements [mailer.Mailer] interface and defines a mail
// client that sends emails via the "sendmail" *nix command.
//
// This client is usually recommended only for development and testing.
type SendMail struct {
	CmdPath string `mapstructure:"cmd_path" json:"cmd_path,omitempty" bson:"cmd_path,omitempty"` // sendmail cmd path
}

// Send implements `mailer.Mailer` interface.
func (c SendMail) Send(m *Message) error {
	toAddresses := addressesToStrings(m.To, false)

	headers := make(http.Header)
	headers.Set("Subject", mime.QEncoding.Encode("utf-8", m.Subject))
	headers.Set("From", m.From.String())
	headers.Set("Content-Type", "text/html; charset=UTF-8")
	headers.Set("To", strings.Join(toAddresses, ","))

	var buffer bytes.Buffer

	if err := headers.Write(&buffer); err != nil {
		return err
	}
	if _, err := buffer.Write([]byte("\r\n")); err != nil {
		return err
	}
	if m.HTML != "" {
		if _, err := buffer.Write([]byte(m.HTML)); err != nil {
			return err
		}
	} else {
		if _, err := buffer.Write([]byte(m.Text)); err != nil {
			return err
		}
	}

	sendmail := exec.Command(c.CmdPath, strings.Join(toAddresses, ","))
	sendmail.Stdin = &buffer

	return sendmail.Run()
}

func findSendmailPath() (string, error) {
	options := []string{
		"/usr/sbin/sendmail",
		"/usr/bin/sendmail",
		"sendmail",
	}

	for _, option := range options {
		path, err := exec.LookPath(option)
		if err == nil {
			return path, err
		}
	}

	return "", errors.New("failed to locate a sendmail executable path")
}
