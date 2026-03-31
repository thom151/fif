package openai

import (
	"context"
	"log"
	"github.com/sashabaranov/go-openai"
)

func GenerateFifScript(ctx context.Context, c *openai.Client, fifDetails, assistantID string) (resp openaiSmartResponse, err error) {
	thread, err := GenThread(ctx, c)
	if err != nil {
		log.Printf("generating thread failed: %v\n", err)
		return openaiSmartResponse{}, err
	}

	err = SendMessage(ctx, c, thread.ID, fifDetails)
	if err != nil {
		log.Printf("sending message failed: %v\n", err)
		return openaiSmartResponse{}, err
	}

	runID, err := GetRunID(ctx, c, thread.ID, assistantID)
	if err != nil {
		log.Printf("getting run id failed: %v\n", err)
		return openaiSmartResponse{}, err
	}

	aiResponse, err := GetResponse(ctx, c, thread.ID, runID)
	if err != nil {
		log.Printf("getting response failed: %v\n", err)
		return openaiSmartResponse{}, err
	}

	return aiResponse, nil
}
