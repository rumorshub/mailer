package mailer

import (
	"os/exec"

	"github.com/roadrunner-server/endure/v2/dep"
	"github.com/roadrunner-server/errors"
)

const (
	PluginName = "mailer"

	smtpKey     = PluginName + ".smtp"
	sendmailKey = PluginName + ".sendmail"
)

type Plugin struct {
	mailer Mailer
}

func (p *Plugin) Init(cfg Configurer) error {
	const op = errors.Op("mailer_plugin_init")

	if !cfg.Has(smtpKey) && !cfg.Has(sendmailKey) {
		return errors.E(op, errors.Disabled)
	}

	if cfg.Has(smtpKey) {
		var client SmtpClient
		if err := cfg.UnmarshalKey(smtpKey, &client); err != nil {
			return errors.E(op, err)
		}

		p.mailer = client
	} else if cfg.Has(sendmailKey) {
		var sendMail SendMail
		if err := cfg.UnmarshalKey(sendmailKey, &sendMail); err != nil {
			return errors.E(op, err)
		}
		if sendMail.CmdPath == "" {
			cmdPath, err := findSendmailPath()
			if err != nil {
				return errors.E(op, err)
			}
			sendMail.CmdPath = cmdPath
		} else if path, err := exec.LookPath(sendMail.CmdPath); err != nil {
			return errors.E(op, err)
		} else {
			sendMail.CmdPath = path
		}

		p.mailer = sendMail
	} else {
		return errors.E(op, errors.Disabled)
	}

	return nil
}

func (p *Plugin) Provides() []*dep.Out {
	return []*dep.Out{
		dep.Bind((*Mailer)(nil), p.Mailer),
	}
}

func (p *Plugin) Mailer() Mailer {
	return p.mailer
}

func (p *Plugin) Name() string {
	return PluginName
}
