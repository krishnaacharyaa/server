package services

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"sync"
	models "whatsapp-bot/model"
)

type PropertyService struct {
	properties []models.Property
	mu         sync.RWMutex
}

func NewPropertyService(filename string) (*PropertyService, error) {
	ps := &PropertyService{}

	if err := ps.loadProperties(filename); err != nil {
		return nil, err
	}

	return ps, nil
}

func (ps *PropertyService) loadProperties(filename string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	log.Printf("📂 Loading properties from file: %s", filename)

	file, err := os.ReadFile(filename)
	if err != nil {
		log.Printf("❌ Error reading properties file: %v", err)
		return err
	}

	if err := json.Unmarshal(file, &ps.properties); err != nil {
		log.Printf("❌ Error parsing properties JSON: %v", err)
		return err
	}

	log.Printf("✅ Loaded %d properties", len(ps.properties))
	return nil
}

func (ps *PropertyService) GetProperties() []models.Property {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.properties
}

func (ps *PropertyService) SearchProperties(query string) []models.Property {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	// Simple search implementation - can be enhanced
	var results []models.Property
	for _, prop := range ps.properties {
		// Basic search logic - expand as needed
		if strings.Contains(strings.ToLower(prop.Location.Locality), strings.ToLower(query)) ||
			strings.Contains(strings.ToLower(prop.Specs.TenantPreference), strings.ToLower(query)) {
			results = append(results, prop)
		}
	}

	log.Printf("🔍 Found %d properties matching query: %s", len(results), query)
	return results
}
