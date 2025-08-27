package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sashabaranov/go-openai"
)

type CutWord struct {
	Word  string `json:"word"`
	Index int    `json:"index"`
}

type openaiSmartResponse struct {
	FullScript string `json:"intro"`
}

func GenThread(ctx context.Context, c *openai.Client) (openai.Thread, error) {
	return c.CreateThread(ctx, openai.ThreadRequest{})
}

func SendMessage(ctx context.Context, c *openai.Client, thread_id, fifDetails string) error {
	fmt.Println("sending message using: ", thread_id)
	request := openai.MessageRequest{
		Role:    "user",
		Content: fifDetails,
	}

	_, err := c.CreateMessage(context.Background(), thread_id, request)
	return err
}

func GetRunID(ctx context.Context, c *openai.Client, thread_id, assistant_id string) (string, error) {
	run, err := c.CreateRun(context.Background(), thread_id, openai.RunRequest{
		AssistantID: assistant_id,
	})

	if err != nil {
		return "", err
	}

	return run.ID, nil
}

func GetResponse(ctx context.Context, c *openai.Client, thread_id, run_id string) (openaiSmartResponse, error) {

	const pollEvery = 750 * time.Millisecond
	ticker := time.NewTicker(pollEvery)
	defer ticker.Stop()

	var smartResp openaiSmartResponse

	for {

		select {
		case <-ctx.Done():
			return openaiSmartResponse{}, ctx.Err()
		case <-ticker.C:
		}

		run, err := c.RetrieveRun(context.Background(), thread_id, run_id)
		if err != nil {
			return openaiSmartResponse{}, err
		}

		switch run.Status {
		case "queued", "in_progress", "requres_action":
			continue

		case "completed":
			messagesList, err := c.ListMessage(context.Background(), thread_id, nil, nil, nil, nil, nil)
			if err != nil {
				return openaiSmartResponse{}, err
			}

			var assistantMsg openai.Message
			for _, msg := range messagesList.Messages {
				if msg.Role == "assistant" {
					assistantMsg = msg
					break
				}
			}

			if len(assistantMsg.Content) == 0 {
				return openaiSmartResponse{}, fmt.Errorf("no assistant content")
			}

			if assistantMsg.Content[0].Type != "text" {
				return openaiSmartResponse{}, fmt.Errorf("assistant message content is not text")
			}

			message := assistantMsg.Content[0].Text.Value

			err = json.Unmarshal([]byte(message), &smartResp)
			if err != nil {
				return openaiSmartResponse{}, err
			}
			return smartResp, nil

		case "failed", "candelled", "expired":
			return openaiSmartResponse{}, fmt.Errorf("run ended with status=%s", run.Status)
		default:
			return openaiSmartResponse{}, fmt.Errorf("unknown run status: %s", run.Status)
		}

	}
}
