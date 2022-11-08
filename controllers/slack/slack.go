package slack

import (
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	slackclient "github.com/slack-go/slack"
	corev1 "k8s.io/api/core/v1"
	"strconv"
)

type Notification struct {
	Name string
}

// SendEvent SendMessage sends a message using all current senders
func SendEvent(event *corev1.Event, secret *corev1.Secret) {
	token := secret.Data["token"]
	channel := secret.Data["channel"]
	logrus.Infof("Sending event to slack channel %s", channel)
	logrus.Infof("Sending event to slack token %s", token)

	api := slackclient.New(string(token))
	attachment := slackclient.Attachment{
		Fields: []slackclient.AttachmentField{
			{
				Title: "Object Kind: " + event.InvolvedObject.Kind,
			},
			{
				Title: "Object Name: " + event.InvolvedObject.Name,
			},
			{
				Title: "Namespace: " + event.InvolvedObject.Namespace,
			},
			{
				Title: "Count: " + strconv.Itoa(int(event.Count)),
			},
			{
				Title: "Reason: " + event.Reason,
			},
			{
				Title: "First Timestamp: " + event.FirstTimestamp.String(),
			},
			{
				Title: "Last Timestamp: " + event.LastTimestamp.String(),
			},
		},
	}

	// Send message to Slack
	channelID, timestamp, err := api.PostMessage(
		string(channel),
		slack.MsgOptionAttachments(attachment),
		slackclient.MsgOptionAsUser(true),
	)
	if err != nil {
		logrus.Errorf("error sending message: %v", err)
	} else {
		logrus.Infof("Message successfully sent to channel %s at %s", channelID, timestamp)
	}
}
