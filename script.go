package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/smtp"
	"net/url"
	"os"
	"os/exec"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/atotto/clipboard"
)

type SMTPConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Mail     string `json:"mail"`
	Password string `json:"password"`
}

func getCurrentTime() string {
	t := time.Now()
	return t.Format("15:04:05")
}

func main() {
	a := app.New()
	w2 := a.NewWindow("SMTP Sender GUI 2023 | " + getCurrentTime())
	w2.Resize(fyne.NewSize(550, 350))
	refreshIcon := theme.ViewRefreshIcon()

	go func() {
		for {
			<-time.After(1 * time.Second)
			w2.SetTitle("SMTP Sender GUI 2023 | " + getCurrentTime())
		}
	}()

	sender := widget.NewEntry()
	sender.SetPlaceHolder("Sender mail ...")
	receiver := widget.NewEntry()
	receiver.SetPlaceHolder("Receiver mail..")
	subject := widget.NewEntry()
	subject.SetPlaceHolder("Subject mail..")
	message := widget.NewEntry()
	message.SetPlaceHolder("Message...")
	emailIcon := theme.MailSendIcon()

	host := widget.NewEntry()
	host.SetPlaceHolder("SMTP Host ...")
	port := widget.NewEntry()
	port.SetPlaceHolder("SMTP Port ...")
	mail := widget.NewEntry()
	mail.SetPlaceHolder("SMTP Mails ...")
	password := widget.NewPasswordEntry()
	password.SetPlaceHolder("SMTP Password ...")
	ConfirmIcon := theme.ConfirmIcon()

	content, err := ioutil.ReadFile("./config.json")
	var payload map[string]interface{}
	err = json.Unmarshal(content, &payload)
	if err != nil {
		log.Fatal(err)
	}

	hostLabel := container.NewHBox(
		widget.NewIcon(theme.HomeIcon()),
		widget.NewLabelWithStyle("HOST |", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(fmt.Sprintf("%s", payload["host"])),
	)

	ports := int(payload["port"].(float64))
	portLabel := container.NewHBox(
		widget.NewIcon(theme.StorageIcon()),
		widget.NewLabelWithStyle("PORT |", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(fmt.Sprintf("%d", ports)),
	)

	emailEntry := widget.NewEntry()
	emailEntry.SetText(fmt.Sprintf("%s", payload["mail"]))
	emailEntry.Disable()

	mailLabel := container.NewHBox(
		widget.NewIcon(theme.MailComposeIcon()),
		widget.NewLabelWithStyle(fmt.Sprintf("MAIL | "), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewButton(fmt.Sprintf("%s", payload["mail"]), func() {
			err := clipboard.WriteAll(emailEntry.Text)
			if err != nil {
				return
			}
		}),
	)

	JSONTab := container.NewVBox(
		hostLabel,
		portLabel,
		mailLabel,
	)
	SMTPsender := container.NewVBox(
		JSONTab,

		sender,
		receiver,
		subject,
		message,

		container.NewHBox(
			layout.NewSpacer(),
			container.NewCenter(
				widget.NewButtonWithIcon("Send Mail", emailIcon, func() {
					SMTPSenders(sender, receiver, subject, message)

				}),
			),
			layout.NewSpacer(),
		),
	)

	SMTPsettings := container.NewVBox(
		widget.NewLabel(" "),
		host,
		port,
		mail,
		password,
		container.NewHBox(
			layout.NewSpacer(),
			container.NewCenter(
				widget.NewButtonWithIcon("Saves", ConfirmIcon, func() {
					configData := SMTPConfig{
						Host:     host.Text,
						Port:     ports,
						Mail:     mail.Text,
						Password: password.Text,
					}

					jsonData, err := json.MarshalIndent(configData, "", "  ")
					if err != nil {
						return
					}

					err = ioutil.WriteFile("config.json", jsonData, 0644)
					if err != nil {
						return
					}
				}),
			),
			layout.NewSpacer(),
		),
	)

	HelpIcon := theme.HelpIcon()
	icon := widget.NewIcon(HelpIcon)
	text := container.NewVBox(
		container.NewHBox(
			icon,
			widget.NewLabelWithStyle("Informations ", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		),
		widget.NewLabel("If you have made a modification, \nplease interact with the button below to update the SMTP configurations."),
	)

	refreshButton := container.NewVBox(
		text,
		widget.NewButtonWithIcon("Reboot", refreshIcon, func() {
			restartApplication()
		}),
	)

	text1 := container.NewVBox(
		container.NewHBox(
			icon,
			widget.NewLabelWithStyle("About ", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		),
		widget.NewLabel("Young French developer, if you liked using this script it would be nice to give this project a star."),
	)

	About := container.NewVBox(
		text1,
		widget.NewButton("GitHub", func() {
			openURLInBrowser("https://github.com/GabrielCtz")
		}),
	)

	tabs := container.NewAppTabs(
		container.NewTabItem("   SMTP Sender", SMTPsender),
		container.NewTabItem("  Settings ", SMTPsettings),
		container.NewTabItem(" Informations", refreshButton),
		container.NewTabItem("About", About))

	w2.SetContent(tabs)
	w2.ShowAndRun()
}

func restartApplication() {
	cmd := exec.Command("cmd.exe", "/K", "restart.bat")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		os.Exit(1)
	}

	err = cmd.Wait()
	if err != nil {
		return
	}
}

var payload struct {
	Host     string `json:"host"`
	Mail     string `json:"mail"`
	Password string `json:"password"`
	Port     int    `json:"port"`
}

func SMTPSenders(sender, receiver, subject, msg *widget.Entry) {
	to := receiver.Text

	content, err := ioutil.ReadFile("./config.json")
	if err != nil {
		return
	}

	err = json.Unmarshal(content, &payload)
	if err != nil {
		return
	}

	auth := smtp.PlainAuth("", payload.Mail, payload.Password, payload.Host)
	message := "To: " + to + "\r\n" +
		"Subject: " + subject.Text + "\r\n\r\n" +
		msg.Text

	err = smtp.SendMail(payload.Host+":"+fmt.Sprint(payload.Port), auth, payload.Mail, []string{to}, []byte(message))
	if err != nil {
		return
	} else {
		return
	}
}

func openURLInBrowser(rawURL string) {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}

	a := app.New()
	a.OpenURL(u)
}
