package openai

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

func GenerateFifScript(ctx context.Context, c *openai.Client, fifDetails, assistantID string) (script string, err error) {
	thread, err := GenThread(ctx, c)
	if err != nil {
		return "", err
	}

	err = SendMessage(ctx, c, thread.ID, fifDetails)
	if err != nil {
		return "", err
	}

	runID, err := GetRunID(ctx, c, thread.ID, assistantID)
	if err != nil {
		return "", err
	}

	aiResponse, err := GetResponse(ctx, c, thread.ID, runID)
	if err != nil {
		return "", err
	}

	return aiResponse.FullScript, nil
}
