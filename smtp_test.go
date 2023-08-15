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
	"net/smtp"
	"testing"
)

func TestLoginAuthStart(t *testing.T) {
	auth := smtpLoginAuth{username: "test", password: "123456"}

	scenarios := []struct {
		name        string
		serverInfo  *smtp.ServerInfo
		expectError bool
	}{
		{
			"localhost without tls",
			&smtp.ServerInfo{TLS: false, Name: "localhost"},
			false,
		},
		{
			"localhost with tls",
			&smtp.ServerInfo{TLS: true, Name: "localhost"},
			false,
		},
		{
			"127.0.0.1 without tls",
			&smtp.ServerInfo{TLS: false, Name: "127.0.0.1"},
			false,
		},
		{
			"127.0.0.1 with tls",
			&smtp.ServerInfo{TLS: false, Name: "127.0.0.1"},
			false,
		},
		{
			"::1 without tls",
			&smtp.ServerInfo{TLS: false, Name: "::1"},
			false,
		},
		{
			"::1 with tls",
			&smtp.ServerInfo{TLS: false, Name: "::1"},
			false,
		},
		{
			"non-localhost without tls",
			&smtp.ServerInfo{TLS: false, Name: "example.com"},
			true,
		},
		{
			"non-localhost with tls",
			&smtp.ServerInfo{TLS: true, Name: "example.com"},
			false,
		},
	}

	for _, s := range scenarios {
		method, resp, err := auth.Start(s.serverInfo)

		hasErr := err != nil
		if hasErr != s.expectError {
			t.Fatalf("[%s] Expected hasErr %v, got %v", s.name, s.expectError, hasErr)
		}

		if hasErr {
			continue
		}

		if len(resp) != 0 {
			t.Fatalf("[%s] Expected empty data response, got %v", s.name, resp)
		}

		if method != "LOGIN" {
			t.Fatalf("[%s] Expected LOGIN, got %v", s.name, method)
		}
	}
}

func TestLoginAuthNext(t *testing.T) {
	auth := smtpLoginAuth{username: "test", password: "123456"}

	{
		// example|false
		r1, err := auth.Next([]byte("example:"), false)
		if err != nil {
			t.Fatalf("[example|false] Unexpected error %v", err)
		}
		if len(r1) != 0 {
			t.Fatalf("[example|false] Expected empty part, got %v", r1)
		}

		// example|true
		r2, err := auth.Next([]byte("example:"), true)
		if err != nil {
			t.Fatalf("[example|true] Unexpected error %v", err)
		}
		if len(r2) != 0 {
			t.Fatalf("[example|true] Expected empty part, got %v", r2)
		}
	}

	// ---------------------------------------------------------------

	{
		// username:|false
		r1, err := auth.Next([]byte("username:"), false)
		if err != nil {
			t.Fatalf("[username|false] Unexpected error %v", err)
		}
		if len(r1) != 0 {
			t.Fatalf("[username|false] Expected empty part, got %v", r1)
		}

		// username:|true
		r2, err := auth.Next([]byte("username:"), true)
		if err != nil {
			t.Fatalf("[username|true] Unexpected error %v", err)
		}
		if str := string(r2); str != auth.username {
			t.Fatalf("[username|true] Expected %s, got %s", auth.username, str)
		}

		// uSeRnAmE:|true
		r3, err := auth.Next([]byte("uSeRnAmE:"), true)
		if err != nil {
			t.Fatalf("[uSeRnAmE|true] Unexpected error %v", err)
		}
		if str := string(r3); str != auth.username {
			t.Fatalf("[uSeRnAmE|true] Expected %s, got %s", auth.username, str)
		}
	}

	// ---------------------------------------------------------------

	{
		// password:|false
		r1, err := auth.Next([]byte("password:"), false)
		if err != nil {
			t.Fatalf("[password|false] Unexpected error %v", err)
		}
		if len(r1) != 0 {
			t.Fatalf("[password|false] Expected empty part, got %v", r1)
		}

		// password:|true
		r2, err := auth.Next([]byte("password:"), true)
		if err != nil {
			t.Fatalf("[password|true] Unexpected error %v", err)
		}
		if str := string(r2); str != auth.password {
			t.Fatalf("[password|true] Expected %s, got %s", auth.password, str)
		}

		// pAsSwOrD:|true
		r3, err := auth.Next([]byte("pAsSwOrD:"), true)
		if err != nil {
			t.Fatalf("[pAsSwOrD|true] Unexpected error %v", err)
		}
		if str := string(r3); str != auth.password {
			t.Fatalf("[pAsSwOrD|true] Expected %s, got %s", auth.password, str)
		}
	}
}
