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
	cfg         *config.AppConfig
	propertySvc *services.PropertyService
	aiClient    services.AIClient
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

	mode := r.URL.Query().Get("hub.mode")
	token := r.URL.Query().Get("hub.verify_token")
	challenge := r.URL.Query().Get("hub.challenge")

	log.Printf("🔎 Verification params - Mode: %s, Token: %s, Challenge: %s", mode, token, challenge)

	if mode == "subscribe" && token == wh.cfg.VerifyToken {
		log.Println("✅ Webhook verified successfully")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(challenge))
	} else {
		log.Println("❌ Webhook verification failed")
		w.WriteHeader(http.StatusForbidden)
	}
}

func (wh *WhatsAppHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	log.Println("📩 Incoming webhook request")

	// Read and log raw body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("❌ Error reading request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	log.Printf("📦 Raw payload: %s", string(body))

	// Parse JSON payload
	var payload models.WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("❌ Error parsing JSON: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("📝 Parsed payload: %+v", payload)

	// Process each entry
	for i, entry := range payload.Entry {
		log.Printf("📋 Entry %d: %s", i+1, entry.ID)

		for j, change := range entry.Changes {
			log.Printf("🔄 Change %d in field: %s", j+1, change.Field)

			// Process messages
			for _, entry := range payload.Entry {
				for _, change := range entry.Changes {
					for _, message := range change.Value.Messages {
						// Get all properties as JSON
						properties := wh.propertySvc.GetProperties()
						propertiesJSON, err := json.Marshal(properties)
						if err != nil {
							log.Printf("❌ Error marshaling properties: %v", err)
							continue
						}

						// Generate AI response with full property data
						ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
						defer cancel()

						responseText, err := wh.aiClient.GeneratePropertyResponse(
							ctx,
							message.Text.Body,
							propertiesJSON,
						)
						if err != nil {
							log.Printf("❌ AI error: %v", err)
							responseText = "Sorry, I couldn't process your request. Please try again later."
						}

						wh.sendMessage(message.From, responseText)
					}
				}
			}
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (wh *WhatsAppHandler) sendMessage(to, text string) {
	if wh.cfg.AccessToken == "" || wh.cfg.PhoneNumberID == "" {
		log.Println("❌ Missing required configuration (ACCESS_TOKEN or PHONE_NUMBER_ID)")
		return
	}

	url := fmt.Sprintf("https://graph.facebook.com/v18.0/%s/messages", wh.cfg.PhoneNumberID)
	log.Printf("📤 Sending message to URL: %s", url)

	payload := models.MessageResponse{
		MessagingProduct: "whatsapp",
		RecipientType:    "individual",
		To:               to,
		Type:             "text",
		Text: struct {
			Body string `json:"body"`
		}{Body: text},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("❌ Error marshaling payload: %v", err)
		return
	}

	log.Printf("📨 Outgoing payload: %s", string(payloadBytes))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("❌ Error creating request: %v", err)
		return
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", wh.cfg.AccessToken))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ Error sending request: %v", err)
		return
	}
	defer resp.Body.Close()

	responseBody, _ := io.ReadAll(resp.Body)
	log.Printf("📩 Response status: %d, body: %s", resp.StatusCode, string(responseBody))

	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ Message sending failed with status: %d", resp.StatusCode)
	}
}
