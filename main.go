package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

// WebhookPayload represents the WhatsApp webhook payload structure
type WebhookPayload struct {
	Object string `json:"object"`
	Entry  []struct {
		ID      string `json:"id"`
		Changes []struct {
			Value struct {
				MessagingProduct string `json:"messaging_product"`
				Metadata         struct {
					DisplayPhoneNumber string `json:"display_phone_number"`
					PhoneNumberID      string `json:"phone_number_id"`
				} `json:"metadata"`
				Contacts []struct {
					Profile struct {
						Name string `json:"name"`
					} `json:"profile"`
					WaID string `json:"wa_id"`
				} `json:"contacts"`
				Messages []struct {
					From      string `json:"from"`
					ID        string `json:"id"`
					Timestamp string `json:"timestamp"`
					Type      string `json:"type"`
					Text      struct {
						Body string `json:"body"`
					} `json:"text"`
				} `json:"messages"`
			} `json:"value"`
			Field string `json:"field"`
		} `json:"changes"`
	} `json:"entry"`
}

// MessageResponse represents the structure for sending a message
type MessageResponse struct {
	MessagingProduct string `json:"messaging_product"`
	RecipientType    string `json:"recipient_type"`
	To               string `json:"to"`
	Type             string `json:"type"`
	Text             struct {
		Body string `json:"body"`
	} `json:"text"`
}

func init() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file (this is fine if using system env vars): %v", err)
	}
}

func main() {
	r := mux.NewRouter()

	// Webhook routes
	r.HandleFunc("/webhook", verifyWebhook).Methods("GET")
	r.HandleFunc("/webhook", handleWebhook).Methods("POST")

	// Health check endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Server is healthy"))
	}).Methods("GET")

	// Get port from environment or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	// Start server with enhanced logging
	log.Printf("🚀 Server starting on port %s", port)
	log.Printf("🔗 Webhook URL: https://[your-ngrok-url]/webhook")
	log.Printf("🔑 Verify Token: %s", os.Getenv("WEBHOOK_VERIFY_TOKEN"))
	log.Printf("📱 Phone Number ID: %s", os.Getenv("PHONE_NUMBER_ID"))

	server := &http.Server{
		Addr:    ":" + port,
		Handler: loggingMiddleware(r),
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("❌ Server failed to start: %v", err)
	}
}

// Middleware for logging all requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("🌐 %s %s %s", r.RemoteAddr, r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}

// Webhook verification handler
func verifyWebhook(w http.ResponseWriter, r *http.Request) {
	log.Println("🔍 Verification request received")

	verifyToken := os.Getenv("WEBHOOK_VERIFY_TOKEN")
	mode := r.URL.Query().Get("hub.mode")
	token := r.URL.Query().Get("hub.verify_token")
	challenge := r.URL.Query().Get("hub.challenge")

	log.Printf("🔎 Verification params - Mode: %s, Token: %s, Challenge: %s", mode, token, challenge)

	if mode == "subscribe" && token == verifyToken {
		log.Println("✅ Webhook verified successfully")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(challenge))
	} else {
		log.Println("❌ Webhook verification failed")
		w.WriteHeader(http.StatusForbidden)
	}
}

// Webhook handler for incoming messages
func handleWebhook(w http.ResponseWriter, r *http.Request) {
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
	var payload WebhookPayload
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
			for k, message := range change.Value.Messages {
				log.Printf("💬 Message %d from %s: %s", k+1, message.From, message.Text.Body)

				// Example: Echo back messages
				if message.Text.Body != "" {
					responseText := fmt.Sprintf("You said: %s", message.Text.Body)
					log.Printf("✉️ Preparing to send response: %s", responseText)
					sendMessage(message.From, responseText)
				}
			}
		}
	}

	w.WriteHeader(http.StatusOK)
}

// Send a message via WhatsApp API
func sendMessage(to, text string) {
	accessToken := os.Getenv("ACCESS_TOKEN")
	phoneNumberID := os.Getenv("PHONE_NUMBER_ID")

	if accessToken == "" || phoneNumberID == "" {
		log.Println("❌ Missing required environment variables (ACCESS_TOKEN or PHONE_NUMBER_ID)")
		return
	}

	url := fmt.Sprintf("https://graph.facebook.com/v18.0/%s/messages", phoneNumberID)
	log.Printf("📤 Sending message to URL: %s", url)

	payload := MessageResponse{
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

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
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
