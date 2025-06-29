package main

import (
	"log"
	"net/http"
	"whatsapp-bot/config"
	"whatsapp-bot/handlers"
	"whatsapp-bot/services"

	"github.com/gorilla/mux"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize services
	propertySvc, err := services.NewPropertyService(cfg.PropertiesFile)
	if err != nil {
		log.Fatalf("❌ Failed to initialize property service: %v", err)
	}

	// Initialize handlers
	whatsappHandler, err := handlers.NewWhatsAppHandler(cfg, propertySvc)
	if err != nil {
		log.Fatalf("❌ Failed to initialize WhatsApp handler: %v", err)
	}

	healthHandler := handlers.NewHealthHandler()

	// Set up router
	r := mux.NewRouter()

	// Webhook routes
	r.HandleFunc("/webhook", whatsappHandler.VerifyWebhook).Methods("GET")
	r.HandleFunc("/webhook", whatsappHandler.HandleWebhook).Methods("POST")

	// Health check endpoint
	r.Handle("/health", healthHandler).Methods("GET")

	// Middleware
	r.Use(loggingMiddleware)

	// Start server
	log.Printf("🚀 Server starting on port %s", cfg.Port)
	log.Printf("🔗 Webhook URL: https://[your-ngrok-url]/webhook")
	log.Printf("🔑 Verify Token: %s", cfg.VerifyToken)
	log.Printf("📱 Phone Number ID: %s", cfg.PhoneNumberID)
	log.Printf("🤖 Gemini API Key: %s", cfg.GeminiAPIKey)
	log.Printf("🏠 Properties loaded: %d", len(propertySvc.GetProperties()))

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("❌ Server failed to start: %v", err)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("🌐 %s %s %s", r.RemoteAddr, r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}
