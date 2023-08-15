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
	"errors"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/domodwyer/mailyak/v3"
)

var _ Mailer = (*SmtpClient)(nil)

type SmtpAuth string

const (
	SmtpAuthPlain SmtpAuth = "PLAIN"
	SmtpAuthLogin SmtpAuth = "LOGIN"
)

type AddressConfig struct {
	Name    string `mapstructure:"name" json:"name,omitempty" bson:"name,omitempty"`          // Proper name; may be empty.
	Address string `mapstructure:"address" json:"address,omitempty" bson:"address,omitempty"` // user@domain
}

// SmtpClient defines a SMTP mail client structure that implements
// `mailer.Mailer` interface.
type SmtpClient struct {
	Host       string        `mapstructure:"host" json:"host,omitempty" bson:"host,omitempty"`
	Port       int           `mapstructure:"port" json:"port,omitempty" bson:"port,omitempty"`
	Username   string        `mapstructure:"username" json:"username,omitempty" bson:"username,omitempty"`
	Password   string        `mapstructure:"password" json:"password,omitempty" bson:"password,omitempty"`
	Tls        bool          `mapstructure:"tls" json:"tls,omitempty" bson:"tls,omitempty"`
	AuthMethod SmtpAuth      `mapstructure:"auth" json:"auth_method,omitempty" bson:"auth_method,omitempty"` // default to "PLAIN"
	From       AddressConfig `mapstructure:"from" json:"from,omitempty" bson:"from,omitempty"`
}

// Send implements `mailer.Mailer` interface.
func (c SmtpClient) Send(m *Message) error {
	if m.From.Name == "" {
		m.From.Name = c.From.Name
	}
	if m.From.Address == "" {
		m.From.Address = c.From.Address
	}

	var smtpAuth smtp.Auth
	if c.Username != "" || c.Password != "" {
		switch c.AuthMethod {
		case SmtpAuthLogin:
			smtpAuth = &smtpLoginAuth{c.Username, c.Password}
		default:
			smtpAuth = smtp.PlainAuth("", c.Username, c.Password, c.Host)
		}
	}

	// create mail instance
	var yak *mailyak.MailYak
	if c.Tls {
		var tlsErr error
		yak, tlsErr = mailyak.NewWithTLS(fmt.Sprintf("%s:%d", c.Host, c.Port), smtpAuth, nil)
		if tlsErr != nil {
			return tlsErr
		}
	} else {
		yak = mailyak.New(fmt.Sprintf("%s:%d", c.Host, c.Port), smtpAuth)
	}

	if m.From.Name != "" {
		yak.FromName(m.From.Name)
	}
	yak.From(m.From.Address)
	yak.Subject(m.Subject)
	yak.HTML().Set(m.HTML)

	if m.Text != "" {
		yak.Plain().Set(m.Text)
	}

	if len(m.To) > 0 {
		yak.To(addressesToStrings(m.To, true)...)
	}

	if len(m.Bcc) > 0 {
		yak.Bcc(addressesToStrings(m.Bcc, true)...)
	}

	if len(m.Cc) > 0 {
		yak.Cc(addressesToStrings(m.Cc, true)...)
	}

	// add attachements (if any)
	for name, data := range m.Attachments {
		yak.Attach(name, data)
	}

	// add custom headers (if any)
	var hasMessageId bool
	for k, v := range m.Headers {
		if strings.EqualFold(k, "Message-ID") {
			hasMessageId = true
		}
		yak.AddHeader(k, v)
	}
	if !hasMessageId {
		// add a default message id if missing
		fromParts := strings.Split(m.From.Address, "@")
		if len(fromParts) == 2 {
			yak.AddHeader("Message-ID", fmt.Sprintf("<%s@%s>",
				PseudorandomString(15),
				fromParts[1],
			))
		}
	}

	return yak.Send()
}

// -------------------------------------------------------------------
// AUTH LOGIN
// -------------------------------------------------------------------

var _ smtp.Auth = (*smtpLoginAuth)(nil)

// smtpLoginAuth defines an AUTH that implements the LOGIN authentication mechanism.
//
// AUTH LOGIN is obsolete[1] but some mail services like outlook requires it [2].
//
// NB!
// It will only send the credentials if the connection is using TLS or is connected to localhost.
// Otherwise authentication will fail with an error, without sending the credentials.
//
// [1]: https://github.com/golang/go/issues/40817
// [2]: https://support.microsoft.com/en-us/office/outlook-com-no-longer-supports-auth-plain-authentication-07f7d5e9-1697-465f-84d2-4513d4ff0145?ui=en-us&rs=en-us&ad=us
type smtpLoginAuth struct {
	username, password string
}

// Start initializes an authentication with the server.
//
// It is part of the [smtp.Auth] interface.
func (a *smtpLoginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	// Must have TLS, or else localhost server.
	// Note: If TLS is not true, then we can't trust ANYTHING in ServerInfo.
	// In particular, it doesn't matter if the server advertises LOGIN auth.
	// That might just be the attacker saying
	// "it's ok, you can trust me with your password."
	if !server.TLS && !isLocalhost(server.Name) {
		return "", nil, errors.New("unencrypted connection")
	}

	return "LOGIN", nil, nil
}

// Next "continues" the auth process by feeding the server with the requested data.
//
// It is part of the [smtp.Auth] interface.
func (a *smtpLoginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch strings.ToLower(string(fromServer)) {
		case "username:":
			return []byte(a.username), nil
		case "password:":
			return []byte(a.password), nil
		}
	}

	return nil, nil
}

func isLocalhost(name string) bool {
	return name == "localhost" || name == "127.0.0.1" || name == "::1"
}
