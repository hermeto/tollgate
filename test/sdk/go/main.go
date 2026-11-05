// Command compat runs the official OpenAI and Anthropic Go SDKs against Tollgate, changing only the
// base URL (spec 002 FR-013, SC-001). Environment: OPENAI_BASE_URL (ending in /v1) and
// ANTHROPIC_BASE_URL (without /v1). The keys are fake; the simulators never call a real provider.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/anthropics/anthropic-sdk-go/v2"
	anthropicoption "github.com/anthropics/anthropic-sdk-go/v2/option"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

const fakeKey = "tg-example-sdk-compat"

func check(cond bool, msg string) error {
	if !cond {
		return errors.New(msg)
	}
	return nil
}

func openaiSuite(ctx context.Context, baseURL string) error {
	client := openai.NewClient(option.WithBaseURL(baseURL), option.WithAPIKey(fakeKey), option.WithMaxRetries(0))
	msgs := []openai.ChatCompletionMessageParamUnion{openai.UserMessage("Say hello.")}

	resp, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{Model: "gpt-4o-mini", Messages: msgs, MaxTokens: openai.Int(32)})
	if err != nil {
		return fmt.Errorf("chat: %w", err)
	}
	u := resp.Usage
	if err := check(resp.Choices[0].Message.Content != "" && u.TotalTokens == u.PromptTokens+u.CompletionTokens, "chat: content or usage"); err != nil {
		return err
	}

	stream := client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Model: "gpt-4o-mini", Messages: msgs,
		StreamOptions: openai.ChatCompletionStreamOptionsParam{IncludeUsage: openai.Bool(true)},
	})
	acc := openai.ChatCompletionAccumulator{}
	for stream.Next() {
		if !acc.AddChunk(stream.Current()) {
			return errors.New("chat stream: chunk could not be accumulated")
		}
	}
	if err := stream.Err(); err != nil {
		return fmt.Errorf("chat stream: %w", err)
	}
	_ = stream.Close()
	if err := check(len(acc.Choices) == 1 && acc.Choices[0].Message.Content != "" && acc.Usage.CompletionTokens > 0, "chat stream: text or usage"); err != nil {
		return err
	}

	emb, err := client.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Model: "text-embedding-3-small",
		Input: openai.EmbeddingNewParamsInputUnion{OfArrayOfStrings: []string{"first", "second"}},
	})
	if err != nil {
		return fmt.Errorf("embeddings: %w", err)
	}
	if err := check(len(emb.Data) == 2 && len(emb.Data[0].Embedding) == 1536, "embeddings: wrong shape"); err != nil {
		return err
	}

	page, err := client.Models.List(ctx)
	if err != nil {
		return fmt.Errorf("models: %w", err)
	}
	var ids []string
	for _, m := range page.Data {
		ids = append(ids, m.ID)
	}
	if err := check(slices.Contains(ids, "gpt-4o-mini"), "models: gpt-4o-mini not listed"); err != nil {
		return err
	}
	fmt.Println("PASS go openai")
	return nil
}

func anthropicSuite(ctx context.Context, baseURL string) error {
	client := anthropic.NewClient(anthropicoption.WithBaseURL(baseURL), anthropicoption.WithAPIKey(fakeKey), anthropicoption.WithMaxRetries(0))
	params := anthropic.MessageNewParams{
		Model:     "claude-sonnet-5-5",
		MaxTokens: 32,
		Messages:  []anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock("Say hello."))},
	}
	msg, err := client.Messages.New(ctx, params)
	if err != nil {
		return fmt.Errorf("messages: %w", err)
	}
	if err := check(len(msg.Content) == 1 && msg.Content[0].Text != "" && msg.Usage.OutputTokens > 0, "messages: text or usage"); err != nil {
		return err
	}

	stream := client.Messages.NewStreaming(ctx, params)
	text, stop := "", ""
	for stream.Next() {
		switch ev := stream.Current().AsAny().(type) {
		case anthropic.ContentBlockDeltaEvent:
			if d, ok := ev.Delta.AsAny().(anthropic.TextDelta); ok {
				text += d.Text
			}
		case anthropic.MessageDeltaEvent:
			stop = string(ev.Delta.StopReason)
		}
	}
	if err := stream.Err(); err != nil {
		return fmt.Errorf("messages stream: %w", err)
	}
	_ = stream.Close()
	if err := check(text != "" && (stop == "end_turn" || stop == "max_tokens"), "messages stream: incomplete"); err != nil {
		return err
	}
	fmt.Println("PASS go anthropic")
	return nil
}

func main() {
	ctx := context.Background()
	if err := openaiSuite(ctx, os.Getenv("OPENAI_BASE_URL")); err != nil {
		fmt.Println("FAIL go:", err)
		os.Exit(1)
	}
	if err := anthropicSuite(ctx, os.Getenv("ANTHROPIC_BASE_URL")); err != nil {
		fmt.Println("FAIL go:", err)
		os.Exit(1)
	}
}
