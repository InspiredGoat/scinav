package main

import (
	"context"
	"fmt"
	openai "github.com/sashabaranov/go-openai"
)

func AIPrompt(prompt string) string {
	client := openai.NewClient("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")
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return ""
	}

	return resp.Choices[0].Message.Content
}

func AiTest() {
	client := openai.NewClient("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")
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Hello",
				},
			},
		},
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return
	}

	fmt.Println(resp.Choices[0].Message.Content)
}

// launch exe process and
