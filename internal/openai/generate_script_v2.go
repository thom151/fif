package openai

import (
	"context"
	"encoding/json"
	"fmt"

	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

func GenerateFifScriptV2(
	ctx context.Context,
	c openai.Client,
	fifDetails string,
) (openaiSmartResponse, error) {

	instructions := `
Generate a personalised real estate video introduction using EXACTLY this template:

Hi {client_name}! I'm {agent_name}. Thank you for the opportunity to present this proposal for your property at {client_address}. With ...

STRICT RULES:
- Replace {client_name} with the provided Client Name.
- Replace {agent_name} with the provided Agent Name.
- Replace {client_address} with the provided Client Address.
- Do NOT change, rewrite, add, or remove any other words.
- Do NOT add any text before the template.
- Do NOT add any text after the template.
- The word "With" must remain exactly as written.
- Return only valid JSON.

Return exactly:

{
	"intro": "the completed template",
	"cut_index": 0
}

cut_index must be the zero-based word index of the word "With" in the completed intro.


So if:

Agent Name: John Smith
Client Name: Thomas
Client Address: 123 Smith Street, Melbourne

it must produce:

{
  "intro": "Hi Thomas! I'm John Smith. Thank you for the opportunity to present this proposal for your property at 123 Smith Street, Melbourne. With ...",
  "cut_index": 21
}

`
	resp, err := c.Responses.New(ctx, responses.ResponseNewParams{
		Model: "gpt-5.6-terra",

		Instructions: openai.String(instructions),

		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(fifDetails),
		},
	})

	if err != nil {
		return openaiSmartResponse{},
			fmt.Errorf("generating FIF script: %w", err)
	}

	responseText := resp.OutputText()

	if responseText == "" {
		return openaiSmartResponse{},
			fmt.Errorf("OpenAI returned an empty response")
	}

	var smartResp openaiSmartResponse

	if err := json.Unmarshal([]byte(responseText), &smartResp); err != nil {
		return openaiSmartResponse{},
			fmt.Errorf(
				"couldn't parse OpenAI response: %w; response=%q",
				err,
				responseText,
			)
	}

	return smartResp, nil
}
