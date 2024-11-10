package mail

import (
	"testing"

	"bitbucket.org/jessyw/go_simplebank/util"
	"github.com/stretchr/testify/require"
)

func TestSendEmailWithGmail(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	config, err := util.LoadConfig("..")
	require.NoError(t, err)

	sender := NewGmailSender(
		config.EmailSenderName,
		config.EmailSenderAddress,
		config.EmailSenderPassword,
	)
	subject := "A test email"
	body := `
	<h1>Hello world</h1>
	<p>This is a test message from <a href="http://simplebank.com">Go simplebank</a></p>
	`
	to := []string{"jamz971@gmail.com"}
	attachFiles := []string{"../README.md"}

	err = sender.SendEmail(to, nil, nil, subject, body, attachFiles)
	require.NoError(t, err)
}
