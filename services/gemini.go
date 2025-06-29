package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

type AIClient interface {
	GeneratePropertyResponse(ctx context.Context, prompt string, propertiesJSON []byte) (string, error)
	Close() error
}

type GeminiAIClient struct {
	client *genai.Client
}

func NewAIClient() (AIClient, error) {
	if err := godotenv.Load(); err != nil {
		log.Printf("Note: No .env file found, using system environment variables")
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable is required")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create AI client: %w", err)
	}

	return &GeminiAIClient{client: client}, nil
}
func (g *GeminiAIClient) GeneratePropertyResponse(ctx context.Context, prompt string, propertiesJSON []byte) (string, error) {
	fullPrompt := fmt.Sprintf(`You are a helpful and friendly Indian property broker. Talk in a simple and casual way, just like local agents in India talk to customers on phone or WhatsApp.

Your tone should feel human, warm, and slightly casual – not robotic or overly polished. Talk like a local broker in Gurgaon or Bangalore.

🔹 Respond based on the user query – if they are not asking about PG, don’t mention PG. 
🔹 Always show a maximum of 3 properties, even if more are available.
🔹 Use phrases like:
  - "Yes, we have some good options"
  - "₹12k rent, semi-furnished, near metro also"
  - "Available now only, can visit anytime"
  - "2BHK, spacious and peaceful locality"
  - "Let me know, I will arrange everything"

📝 Use this structure:
- Acknowledge: “Yes, got some options for you”
- Mention key things: location, price, furnishing, special features
- Use bullets (•) for each property (max 3)
- End with friendly note like: "Let me know if like any of these options, I’ll arrange visit."

💡 Use Indian English, not formal. Keep it friendly and real. Avoid terms like "residence", "dormitory", "occupants".

💰 Format prices as ₹12k, ₹18.5k etc.
🏠 Use local real-estate terms: "PG", "2BHK", "semi-furnished", "sharing room", etc.

Here’s the property data in JSON:
%s

And here is the user’s question:
%s

Give the reply in this desi Indian agent tone.`, propertiesJSON, prompt)

	model := g.client.GenerativeModel("gemini-1.5-flash")
	var temperature float32 = 0.7
	var topP float32 = 0.9
	var candidateCount int32 = 1

	model.GenerationConfig = genai.GenerationConfig{
		Temperature:    &temperature,
		TopP:           &topP,
		CandidateCount: &candidateCount,
	}

	resp, err := model.GenerateContent(ctx, genai.Text(fullPrompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	response := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])

	// Post-processing to desify the output
	response = strings.ReplaceAll(response, "dormitory", "sharing room")
	response = strings.ReplaceAll(response, "$", "₹")
	response = strings.ReplaceAll(response, "per month", "/month")
	response = strings.ReplaceAll(response, "Rs.", "₹")
	response = strings.ReplaceAll(response, "semi furnished", "semi-furnished")
	response = strings.ReplaceAll(response, "furnished", "fully furnished")

	// Optional cleanup for extra whitespace, excessive bullets, etc., can be added

	return response, nil
}

func (g *GeminiAIClient) Close() error {
	if g.client != nil {
		return g.client.Close()
	}
	return nil
}
