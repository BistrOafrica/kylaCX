package templates

import (
	"bytes"
	"html/template"
)

type ChangePasswordData struct {
	Name         string
	Email        string
	ClientEmail  string
	SupportEmail string
	Year         string
}

func CHANGE_PASSWORD_TEMP(data ChangePasswordData) (string, error) {
	const htmlTemplate = `
    <!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <style>
            html{
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
            p {
                color: #5c5c5c;
            }
            .button {
                background-color: #4CAF50;
                color: white;
                padding: 10px 20px;
                text-align: center;
                text-decoration: none;
                display: inline-block;
                border-radius: 4px;
            }
        </style>
    </head>
    <body>
        <div class="container">
            <div style="text-align: left; width: min-content;">
                <img src="https://kyla.com/_next/image?url=%2Flogo-color.png&w=384&q=75" alt="Company Logo" style="display: block; margin: 0 auto; width: 150px;">
            </div>
            <!-- caution emoji -->
            <div style="text-align: center; font-size: 30px; color: #ff8c00;"></div>
            <h1>Successful!: Change Account Password</h1>
            <br/>
            <p>Hi {{.Name}},</p>
            <p>This is to confirm that the password associated with your <br/>Bonga CX (kyla) account has been changed successfully.</p>
            <br/>
            <p>If you did not make this request, please contact us immediately.</p>
            <br/>
            <p>Thank you!</p>
            <p>Team kyla</p>

            <hr style="margin: 20px 0; border: 1px solid #e1e1e1;">
			<p style="text-align: center; color: #b2b2b2;">This message was sent to {{.ClientEmail}} by kyla Accounts Service for account security update</p>
            <p style="text-align: center; color: #b2b2b2;">If you have any questions, please contact us at <a href="mailto:{{.SupportEmail}}" style="color: #be7012; text-decoration: none;">{{.SupportEmail}}</a></p>
            <p style="text-align: center; color: #b2b2b2;">&copy; {{.Year}} kyla.Inc All rights reserved.</p>
        </div>       
    </body>
    </html>
    `

	// Parse the HTML template
	tmpl, err := template.New("email").Parse(htmlTemplate)
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
