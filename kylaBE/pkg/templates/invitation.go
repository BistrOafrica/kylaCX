package templates

import (
	"bytes"
	"html/template"
)

type InvitationEmailData struct {
	OrganisationName string
	InvitationURL    string
	ExpirationHours  int
	ClientEmail      string
	SupportEmail     string
	Year             string
}

func INVITATION_EMAIL(data InvitationEmailData) (string, error) {
	// Parse the HTML template
	htmlMessage := `
    <!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <style>
        html {
            height: fit-content;
            min-height: max-content;
            width: 100%;
        }
        body {
            font-family: Arial, Helvetica, sans-serif;
            background-color: #f5f5f5;
            display: grid;
            width: 100%;
            height: fit-content;
            justify-content: center;
            align-items: stretch;
        }
        .container {
            width: fit-content;
            height: 100%;
            margin: auto auto;
            background-color: #ffffff;
            padding: 60px 30px;
            border-radius: 3px;
            border: 1px solid #e1e1e1;
        }
        h1 {
            color: #333333;
        }
        p,
        li {
            color: #5c5c5c;
        }
        .button {
            background-color: #e1881b;
            color: white;
            padding: 10px 20px;
            text-align: center;
            text-decoration: none;
            display: inline-block;
            border-radius: 4px;
            width: calc(100% - 40px);
        }
        a {
            color: #707070;
            text-decoration: none;
        }
        </style>
    </head>
    <body>
        <div class="container">
        <div style="text-align: left; width: min-content">
            <img
            src="https://dialafrika.com/_next/image?url=%2Flogo-color.png&w=384&q=75"
            alt="Company Logo"
            style="display: block; margin: 0 auto; width: 150px"
            />
        </div>
        <h1>You've Been Invited! 🎉</h1>
        <br />
        <p>Hello,</p>
        <br />
        <p>
            You've been invited to join {{ .OrganisationName }} on Bonga by Dialafrika.
            We're excited to have you on board!
        </p>
        <p>
            Click the button below to accept the invitation and get started with your journey.
            This invitation will expire in {{ .ExpirationHours }} hours.
        </p>
        <br />
        <a href="{{ .InvitationURL }}" class="button">Accept Invitation 🚀</a>
        <br />
        <p>Have the best Day,</p>
        <p>The Bonga Team at Dialafrika.Inc</p>

        <hr style="margin: 20px 0; border: 1px solid #e1e1e1" />
        <p style="text-align: left; color: #b2b2b2">
            This message was sent to {{.ClientEmail}} by Dialafrika Accounts Service
            for your invitation to join {{ .OrganisationName }}<br />If you have any questions, please
            contact us at <a href="mailto:{{ .SupportEmail}}">{{ .SupportEmail}}</a>
        </p>

        <p style="text-align: left; color: #b2b2b2">
            Bonga: A product of Dialafrika Inc. <br /> 16192 coastal Highway, Lewes,
            Delaware 19958, County of Sussex
        </p>
        <br />
        <div style="display: flex; width:100%; justify-content: space-evenly">
            <a href="#">Unsubscribe</a>|
            <a href="https://dialafrika.com/privacy">Privacy Policy</a>|
            <a href="https://dialafrika.com/terms">Terms of Service</a>|
            <a href="https://dialafrika.com/contact">Contact Us</a>|
            <a href="https://dialafrika.com/about">About Us</a>|
        </div>
        <p style="text-align: center; color: #b2b2b2">
            &copy; {{ .Year}} Dialafrika.Inc. All rights reserved.
        </p>
        </div>
    </body>
    </html>
    `
	tmpl, err := template.New("email").Parse(htmlMessage)
	if err != nil {
		return "", err
	}

	// Execute the template with the provided data
	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, data); err != nil {
		return "", err
	}

	return tpl.String(), nil
}
