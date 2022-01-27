package email

import (
	"github.com/cambefus/gcp_go_utils/secrets"
	"testing"
)

const sendGridAPI = `SENDGRID_API_KEY`

var em *EMailer
var sendTo string

func setup(t *testing.T) {
	if em == nil {
		s, e := secrets.InitializeFromEnvironment(`utilities_config`)
		if e != nil {
			t.Error(e)
		}
		em = NewEMailer(s.GetString(sendGridAPI), s.GetString(`FROM_EMAIL`), `Bilbo`)
		sendTo = s.GetString(`EMAIL_RECIPIENT`)
	}
}

func Test_sendTextMsg(t *testing.T) {
	setup(t)
	e := em.SendTextMsg(`test plain text email msg`, sendTo, "this is a test email with no html")
	if e != nil {
		t.Error(e)
	}
}

func Test_sendAttachment(t *testing.T) {
	setup(t)
	const adata = `the quick brown fox jumped over the lazy dog.`
	em.AddAttachment([]byte(adata), `afile.txt`, "text/plain")
	e := em.SendTextMsg(`test plain text email msg with attachment`, sendTo, "this is a test email with file attached")
	if e != nil {
		t.Error(e)
	}
}

func Test_sendHTMLMsg(t *testing.T) {
	setup(t)
	msg := `
	<!DOCTYPE html>
	<html>
	<body>
	<h1>Heading 1</h1>
	<h2>Heading 2</h2>
	<h3>Heading 3</h3>
	<p>Here is a paragraph</p>
	</body>
	</html>
	`
	e := em.SendHTMLMsg(`test html email msg`, sendTo, msg)
	if e != nil {
		t.Error(e)
	}
}
