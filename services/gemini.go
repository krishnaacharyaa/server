package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	models "whatsapp-bot/model"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

type AIClient interface {
	GeneratePropertyResponse(ctx context.Context, prompt string, properties []models.Property) (string, []string, error)
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

func (g *GeminiAIClient) GeneratePropertyResponse(ctx context.Context, prompt string, properties []models.Property) (string, []string, error) {
	// Build image list
	var allImageURLs []string
	var propertiesWithImages []map[string]interface{}

	for _, prop := range properties {
		propData := map[string]interface{}{
			"id":          prop.ID,
			"locality":    prop.Location.Locality,
			"price":       prop.Price,
			"description": fmt.Sprintf("%s %s", prop.PropertyType, prop.Specs.Furnishing),
		}

		// Add images if available
		if len(prop.Images) > 0 {
			propData["images"] = prop.Images
			allImageURLs = append(allImageURLs, prop.Images[0].URL) // Use first image as primary
		}

		propertiesWithImages = append(propertiesWithImages, propData)
	}

	propertiesJSON, _ := json.Marshal(propertiesWithImages)

	fullPrompt := fmt.Sprintf(`You are a property advisor in India. Suggest properties from this JSON data.
Include image references like [Image 1] when mentioning properties. Be concise and friendly.

Properties: %s
Question: %s`, propertiesJSON, prompt)

	model := g.client.GenerativeModel("gemini-2.5-pro")
	resp, err := model.GenerateContent(ctx, genai.Text(fullPrompt))
	if err != nil {
		return "", nil, err
	}

	if len(resp.Candidates) == 0 {
		return "", nil, fmt.Errorf("no response from AI")
	}

	responseText := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
	return responseText, allImageURLs, nil
}

func (g *GeminiAIClient) Close() error {
	if g.client != nil {
		return g.client.Close()
	}
	return nil
}
