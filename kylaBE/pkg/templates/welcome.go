package templates

import (
	"bytes"
	"html/template"
)

type WelcomeEmailData struct {
	Name              string
	Username          string
	Password          string
	LoginLinkRedirect string
	ClientEmail       string
	SupportEmail      string
	Year              string
}

func WELCOME_EMAIL(data WelcomeEmailData) (string, error) {
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
        <!-- caution emoji -->
        <div style="text-align: center; font-size: 30px; color: #ff8c00"></div>
        <h1>Woohoo! Welcome to Dialafrika! 🌟</h1>
        <br />
        <p>Hi {{ .Name}},</p>
        <br />
        <p>
            You're receiving this email because you've just created a new account,
            or someone has invited you to their organisation account, on Bonga by Dialafrika.
        </p>
        <p>
            We're thrilled to have you on board, and we can't wait to embark on this
            exciting journey with you. <br />
            Get ready for some amazing experience, great features, and lots of fun
            engaging with your customers and colleagues!
        </p>

        <p>Here are a few things you can do to get started:🚀</p>
        <ul>
            <li>Update your profile and set your password</li>
            <li>Invite your team members to join you</li>
            <li>Start creating your first project</li>
            <li>Explore the platform and let us know if you have any questions</li>
        </ul>

        <p>
            Find below credentials to your new Bonga account or click the button
            after that to get started
        </p>
        <br />
        <div
            style="
            background-color: #e3e3e3;
            padding: 10px;
            border-radius: 4px;
            margin: 10px 0;
            "
        >
            <p><strong>Username:</strong> {{ .Username}}</p>
            <p><strong>Password:</strong> {{ .Password}}</p>
        </div>

        <a href="{{ .LoginLinkRedirect}}" class="button"
            >Login to Get Started 🚀</a
        >
        <br />

        <p>Have the best Day,</p>
        <p>The Bonga Team at Dialafrika.Inc</p>

        <hr style="margin: 20px 0; border: 1px solid #e1e1e1" />
        <p style="text-align: left; color: #b2b2b2">
            This message was sent to {{.ClientEmail}} by Dialafrika Accounts Service
            for your new account creation<br />If you have any questions, please
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
