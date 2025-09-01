package openai

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

func GenerateFifScript(ctx context.Context, c *openai.Client, fifDetails, assistantID string) (resp openaiSmartResponse, err error) {
	thread, err := GenThread(ctx, c)
	if err != nil {
		return openaiSmartResponse{}, err
	}

	err = SendMessage(ctx, c, thread.ID, fifDetails)
	if err != nil {
		return openaiSmartResponse{}, err
	}

	runID, err := GetRunID(ctx, c, thread.ID, assistantID)
	if err != nil {
		return openaiSmartResponse{}, err
	}

	aiResponse, err := GetResponse(ctx, c, thread.ID, runID)
	if err != nil {
		return openaiSmartResponse{}, err
	}

	return aiResponse, nil
}
