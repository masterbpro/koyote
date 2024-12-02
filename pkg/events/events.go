package events

import (
	"encoding/json"
	"fmt"
	log "github.com/gookit/slog"
	"github.com/koyote/pkg/config"
	"github.com/koyote/pkg/redis"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

type Event interface {
	TemplateMessage() (string, error)
}

func (e GitlabJobEvent) TemplateMessage() (string, error) {
	result, err := templateJobEventMessage(e, "job.tpl")
	if err != nil {
		return "", errors.Wrap(err, "Error while templating message!")
	}

	return result, nil
}

func (e GitlabMergeRequestEvent) TemplateMessage() (string, error) {
	result, err := templateMREventMessage(e, "merge_request.tpl")
	if err != nil {
		return "", errors.Wrap(err, "Error while templating message!")
	}

	return result, nil
}

func (e GitlabNoteEvent) TemplateMessage() (string, error) {
	result, err := templateNoteEventMessage(e, "note.tpl")
	if err != nil {
		return "", errors.Wrap(err, "Error while templating message!")
	}

	return result, nil
}

func (e GitlabPipelineEvent) TemplateMessage() (string, error) {
	result, err := templatePipelineEventMessage(e, "pipeline.tpl")
	if err != nil {
		return "", errors.Wrap(err, "Error while templating message!")
	}

	return result, nil
}

func (e GitlabPushEvent) TemplateMessage() (string, error) {
	result, err := templatePushEventMessage(e, "push.tpl")
	if err != nil {
		return "", errors.Wrap(err, "Error while templating message!")
	}

	return result, nil
}

func EventMatcher(eventJSON []byte, chatID, threadID string) error {
	var receivedEventType GitlabEventTypeDetector
	err := json.Unmarshal(eventJSON, &receivedEventType)
	if err != nil {
		return errors.Wrap(err, "Error while parsing event JSON!")
	}

	event, err := eventComparator(receivedEventType.ObjectKind, eventJSON)
	if err != nil {
		return errors.Wrap(err, "Error while templating message!")
	}

	eventMessage, err := event.TemplateMessage()
	if err != nil {
		return errors.Wrap(err, "Error while templating message!")
	}

	chatIDInt64, err := strconv.ParseInt(chatID, 10, 64)
	if err != nil {
		return errors.Wrap(err, "invalid chatID")
	}

	var threadIDInt int
	if threadID != "" {
		parsedThreadID, err := strconv.Atoi(threadID)
		if err != nil {
			return errors.Wrap(err, "invalid threadID")
		}
		threadIDInt = parsedThreadID
	}

	//err = telegram.SendEventMessage(chatIDInt64, threadIDInt, eventMessage)
	err = errors.New("d")
	if err != nil {
		if config.GlobalAppConfig.Redis.Enabled {
			log.Warn("Error while sending event to Telegram. Trying to save in Redis")
			redis.PublishEventToRedisChannel(
				redis.MessageType{
					ChatID:   chatIDInt64,
					ThreadID: threadIDInt,
					Message:  eventMessage,
					Time:     time.Now(),
				},
			)
		} else {
			return errors.Wrap(err, "Error while sending event to Telegram and Redis was disabled. Event may be lost :(")
		}
	}

	return nil
}

func eventComparator(eventType string, data []byte) (Event, error) {
	var event Event
	switch eventType {
	case "build":
		event = &GitlabJobEvent{}
	case "merge_request":
		event = &GitlabMergeRequestEvent{}
	case "note":
		event = &GitlabNoteEvent{}
	case "pipeline":
		event = &GitlabPipelineEvent{}
	case "push":
		event = &GitlabPushEvent{}
	default:
		return nil, fmt.Errorf("Unknown event type: %s", eventType)
	}

	return event, json.Unmarshal(data, &event)
}
