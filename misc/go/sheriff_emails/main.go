// sheriff_emails is an application that emails the next sheriff every week.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"strings"

	"go.skia.org/infra/go/common"
	"go.skia.org/infra/go/email"
	"go.skia.org/infra/go/httputils"
	"go.skia.org/infra/go/metadata"
	"go.skia.org/infra/go/sklog"
	"go.skia.org/infra/go/util"
)

const (
	ROTATIONS_GMAIL_CACHED_TOKEN = "rotations_gmail_cached_token"

	NEXT_SHERIFF_JSON_URL = "http://skia-tree-status.appspot.com/next-sheriff"

	EXTRA_RECIPIENT = "rmistry@google.com"

	EMAIL_TEMPLATE = `
Hi {{.SheriffName}},
<br/><br/>

You will be the {{.SheriffType}} for the coming week ({{.ScheduleStart}} - {{.ScheduleEnd}}).
<br/><br/>

Documentation for {{.SheriffType}}s is in {{.SheriffDoc}}.
<br/><br/>

The schedule for {{.SheriffType}}s is in {{.SheriffSchedules}}.
<br/><br/>

If you need to swap shifts with someone (because you are out sick or on vacation), please get approval from the person you want to swap with. Then send an email to skiabot@google.com to have someone make the change in the database (or directly ping rmistry@).
<br/><br/>

Please let skiabot@google.com know if you have any other questions.
<br/><br/>

Thanks!
`
)

type ShiftType struct {
	shiftName           string
	schedulesLink       string
	documentationLink   string
	nextSheriffEndpoint string
}

const (
	// This placeholder is used to signify that nobody will be sheriffing that
	// week.
	NO_SHERIFF = "nobody"
)

var (
	local            = flag.Bool("local", false, "Running locally if true. As opposed to in production.")
	emailTokenPath   = flag.String("email_token_path", "", "The file where the email token can be found.")
	skiaSheriffShift = &ShiftType{shiftName: "Skia Sheriff", schedulesLink: "http://skia-tree-status.appspot.com/sheriff", documentationLink: "https://skia.org/dev/sheriffing", nextSheriffEndpoint: "http://skia-tree-status.appspot.com/next-sheriff"}
	gpuWranglerShift = &ShiftType{shiftName: "GPU Wrangler", schedulesLink: "http://skia-tree-status.appspot.com/gpu-sheriff", documentationLink: "https://skia.org/dev/sheriffing/gpu", nextSheriffEndpoint: "http://skia-tree-status.appspot.com/next-gpu-sheriff"}
	robocopShift     = &ShiftType{shiftName: "Android Robocop", schedulesLink: "http://skia-tree-status.appspot.com/robocop", documentationLink: "https://skia.org/dev/sheriffing/android", nextSheriffEndpoint: "http://skia-tree-status.appspot.com/next-robocop"}
	trooperShift     = &ShiftType{shiftName: "Infra Trooper", schedulesLink: "http://skia-tree-status.appspot.com/trooper", documentationLink: "https://skia.org/dev/sheriffing/trooper", nextSheriffEndpoint: "http://skia-tree-status.appspot.com/next-trooper"}
	allShiftTypes    = []*ShiftType{skiaSheriffShift, gpuWranglerShift, robocopShift, trooperShift}
)

// sendEmail sends an email with the specified header and body to the recipients.
func sendEmail(tokenPath string, recipients []string, senderDisplayName, subject, body, markup string) error {
	gmail, err := email.NewGMail(
		"292895568497-u2m421dk2htq171bfodi9qoqtb5smuea.apps.googleusercontent.com",
		"jv-g54CaPS783QV6H8SdagYn",
		tokenPath)
	if err != nil {
		return fmt.Errorf("Could not initialize gmail object: %s", err)
	}
	if err := gmail.SendWithMarkup(senderDisplayName, recipients, subject, body, markup); err != nil {
		return fmt.Errorf("Could not send email: %s", err)
	}
	return nil
}

func main() {
	defer common.LogPanic()
	common.Init()

	tokenPath := *emailTokenPath
	if tokenPath == "" {
		if *local {
			sklog.Error("Must specify --email_token_path")
			return
		} else {
			// Look in metadata for email token.
			cachedGMailToken := metadata.Must(metadata.ProjectGet(ROTATIONS_GMAIL_CACHED_TOKEN))
			tokenFile, err := ioutil.TempFile("", "sheriff-emails")
			if err != nil {
				sklog.Error("Could not create temp file")
				return
			}
			tokenPath = tokenFile.Name()
			defer util.Remove(tokenPath)
			if err := ioutil.WriteFile(tokenPath, []byte(cachedGMailToken), os.ModePerm); err != nil {
				sklog.Fatalf("Failed to cache token: %s", err)
			}
		}
	}

	defer sklog.Flush()

	client := httputils.NewTimeoutClient()
	for _, shiftType := range allShiftTypes {

		res, err := client.Get(shiftType.nextSheriffEndpoint)
		if err != nil {
			sklog.Fatalf("Could not HTTP Get: %s", err)
		}
		defer util.Close(res.Body)

		var jsonType map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&jsonType); err != nil {
			sklog.Fatalf("Could not unmarshal JSON: %s", err)
		}
		sheriffEmail, _ := jsonType["username"].(string)
		if sheriffEmail == NO_SHERIFF {
			sklog.Infof("Skipping emailing %s because %s was specified", shiftType.shiftName, NO_SHERIFF)
			continue
		}
		sheriffUsername := strings.Split(string(sheriffEmail), "@")[0]

		emailTemplateParsed := template.Must(template.New("sheriff_email").Parse(EMAIL_TEMPLATE))
		emailBytes := new(bytes.Buffer)
		if err := emailTemplateParsed.Execute(emailBytes, struct {
			SheriffName      string
			SheriffType      string
			SheriffSchedules string
			SheriffDoc       string
			ScheduleStart    string
			ScheduleEnd      string
		}{
			SheriffName:      sheriffUsername,
			SheriffType:      shiftType.shiftName,
			SheriffSchedules: shiftType.schedulesLink,
			SheriffDoc:       shiftType.documentationLink,
			ScheduleStart:    jsonType["schedule_start"].(string),
			ScheduleEnd:      jsonType["schedule_end"].(string),
		}); err != nil {
			sklog.Errorf("Failed to execute template: %s", err)
			return
		}

		emailSubject := fmt.Sprintf("%s is the next %s", sheriffUsername, shiftType.shiftName)
		viewActionMarkup, err := email.GetViewActionMarkup(shiftType.schedulesLink, "View Schedule", "Direct link to the schedule")
		if err != nil {
			sklog.Errorf("Failed to get view action markup: %s", err)
			return
		}
		senderDisplayName := fmt.Sprintf("%s Rotation", shiftType.shiftName)
		if err := sendEmail(tokenPath, []string{sheriffEmail, EXTRA_RECIPIENT}, senderDisplayName, emailSubject, emailBytes.String(), viewActionMarkup); err != nil {
			sklog.Fatalf("Error sending email to sheriff: %s", err)
		}
	}
}
