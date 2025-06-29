package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
	"whatsapp-bot/config"
	models "whatsapp-bot/model"
	"whatsapp-bot/services"
)

type WhatsAppHandler struct {
	cfg            *config.AppConfig
	propertySvc    *services.PropertyService
	aiClient       services.AIClient
	messageCounter int // For tracking message sequence
}

func NewWhatsAppHandler(cfg *config.AppConfig, propertySvc *services.PropertyService) (*WhatsAppHandler, error) {
	aiClient, err := services.NewAIClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create AI client: %w", err)
	}

	return &WhatsAppHandler{
		cfg:         cfg,
		propertySvc: propertySvc,
		aiClient:    aiClient,
	}, nil
}

func (wh *WhatsAppHandler) VerifyWebhook(w http.ResponseWriter, r *http.Request) {
	log.Println("🔍 Verification request received")

	query := r.URL.Query()
	mode := query.Get("hub.mode")
	token := query.Get("hub.verify_token")
	challenge := query.Get("hub.challenge")

	log.Printf("🔎 Verification params - Mode: %s, Token: %s", mode, token)

	if mode == "subscribe" && token == wh.cfg.VerifyToken {
		log.Println("✅ Webhook verified successfully")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(challenge))
		return
	}

	log.Println("❌ Webhook verification failed")
	w.WriteHeader(http.StatusForbidden)
}

func (wh *WhatsAppHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	log.Println("📩 Incoming webhook request")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("❌ Error reading request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	log.Printf("📦 Raw payload size: %d bytes", len(body))

	var payload models.WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("❌ Error parsing JSON: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for _, entry := range payload.Entry {
		log.Printf("📋 Processing entry ID: %s", entry.ID)

		for _, change := range entry.Changes {
			for _, message := range change.Value.Messages {
				wh.messageCounter++
				log.Printf("💬 Message #%d from %s: %s",
					wh.messageCounter,
					message.From,
					message.Text.Body)

				go wh.processMessage(message.From, message.Text.Body)
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	log.Printf("✅ Request processed in %v", time.Since(startTime))
}

func (wh *WhatsAppHandler) processMessage(sender, message string) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	properties := wh.propertySvc.GetProperties()

	// Get AI response with potential image references
	responseText, imageURLs, err := wh.aiClient.GeneratePropertyResponse(
		ctx,
		message,
		properties,
	)
	if err != nil {
		log.Printf("❌ AI error: %v", err)
		responseText = "Sorry, I couldn't process your request. Please try again later."
	}

	// Send text response first
	if err := wh.sendTextMessage(sender, responseText); err != nil {
		log.Printf("❌ Failed to send text message: %v", err)
		return
	}

	// Send images if available
	if len(imageURLs) > 0 {
		wh.sendImages(sender, imageURLs)
	}
}

func (wh *WhatsAppHandler) sendTextMessage(to, text string) error {
	payload := models.MessageResponse{
		MessagingProduct: "whatsapp",
		RecipientType:    "individual",
		To:               to,
		Type:             "text",
		Text: struct {
			Body string `json:"body"`
		}{Body: text},
	}

	return wh.sendAPIRequest(payload)
}

func (wh *WhatsAppHandler) sendImages(to string, imageURLs []string) {
	maxImages := 5 // WhatsApp limit
	if len(imageURLs) > maxImages {
		imageURLs = imageURLs[:maxImages]
	}

	for _, imgURL := range imageURLs {
		payload := map[string]interface{}{
			"messaging_product": "whatsapp",
			"recipient_type":    "individual",
			"to":                to,
			"type":              "image",
			"image": map[string]interface{}{
				"link": imgURL,
			},
		}

		if err := wh.sendAPIRequest(payload); err != nil {
			log.Printf("❌ Failed to send image: %v", err)
			continue
		}

		time.Sleep(500 * time.Millisecond) // Rate limiting
	}
}

func (wh *WhatsAppHandler) sendAPIRequest(payload interface{}) error {
	if wh.cfg.AccessToken == "" || wh.cfg.PhoneNumberID == "" {
		return fmt.Errorf("missing required configuration (ACCESS_TOKEN or PHONE_NUMBER_ID)")
	}

	url := fmt.Sprintf("https://graph.facebook.com/v18.0/%s/messages", wh.cfg.PhoneNumberID)

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+wh.cfg.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
