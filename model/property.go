package models

type Property struct {
	ID           string `json:"id"`
	ListingType  string `json:"listing_type"`  // rent/buy
	PropertyType string `json:"property_type"` // flat/house/pg/villa/plot
	Subtype      string `json:"subtype"`       // residential/commercial
	Status       string `json:"status"`        // available/rented/sold

	Bedrooms    int  `json:"bedrooms"` // Per person
	Price       int  `json:"price"`    // Per bed
	Deposit     int  `json:"deposit"`
	Maintenance int  `json:"maintenance"` // Usually included
	Negotiable  bool `json:"negotiable"`
	FinalPrice  int  `json:"final_price"`

	Location struct {
		City          string   `json:"city"`
		Locality      string   `json:"locality"`
		MicroLocality string   `json:"micro_locality"`
		Landmarks     []string `json:"landmarks"`
		Transport     struct {
			AutoRickshaw bool   `json:"auto_rickshaw"`
			NearestMetro string `json:"nearest_metro"`
		} `json:"transport"`
		Coordinates struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"coordinates"`
	} `json:"location"`

	Specs struct {
		Bedrooms         int    `json:"bedrooms"`    // Single/double sharing
		Bathrooms        int    `json:"bathrooms"`   // Shared/attached
		CarpetArea       int    `json:"carpet_area"` // Per person
		Furnishing       string `json:"furnishing"`  // furnished/semi-furnished/unfurnished
		Floor            int    `json:"floor"`
		TotalFloors      int    `json:"total_floors"`
		Age              int    `json:"age"`               // Years
		AvailableFrom    string `json:"available_from"`    // Date
		TenantPreference string `json:"tenant_preference"` // working_professional/student
		GenderPreference string `json:"gender_preference"` // male/female/any
	} `json:"specs"`

	// PG-specific fields
	PGSpecific struct {
		MealsIncluded  string `json:"meals_included"` // none/breakfast/lunch/dinner/all
		RoomSharing    string `json:"room_sharing"`   // single/double/triple/dormitory
		LaundryService bool   `json:"laundry_service"`
		Housekeeping   string `json:"housekeeping"` // daily/weekly
		CurfewTime     string `json:"curfew_time"`
		VisitorPolicy  string `json:"visitor_policy"` // day_only/restricted
		FoodType       string `json:"food_type"`      // veg/nonveg/both
	} `json:"pg_specific"`

	Amenities struct {
		Wifi           bool `json:"wifi"`
		AC             bool `json:"ac"`
		TV             bool `json:"tv"`
		Refrigerator   bool `json:"refrigerator"`
		WashingMachine bool `json:"washing_machine"`
		WaterPurifier  bool `json:"water_purifier"`
		PowerBackup    bool `json:"power_backup"`
		Security       bool `json:"security"`
		Gym            bool `json:"gym,omitempty"`
		SwimmingPool   bool `json:"swimming_pool,omitempty"`
	} `json:"amenities"`

	Images []struct {
		URL     string `json:"url"`
		Caption string `json:"caption,omitempty"`
	} `json:"images,omitempty"`
}

type PropertyResponse struct {
	Properties []Property `json:"properties"`
	Count      int        `json:"count"`
	Page       int        `json:"page,omitempty"`
	PerPage    int        `json:"per_page,omitempty"`
	TotalPages int        `json:"total_pages,omitempty"`
}

// FilterCriteria represents filters for property search
type FilterCriteria struct {
	MinPrice     int      `json:"min_price,omitempty"`
	MaxPrice     int      `json:"max_price,omitempty"`
	Locality     []string `json:"locality,omitempty"`
	PropertyType []string `json:"property_type,omitempty"`
	Furnishing   string   `json:"furnishing,omitempty"`
	Bedrooms     int      `json:"bedrooms,omitempty"`
	GenderPref   string   `json:"gender_pref,omitempty"`
	WithPhotos   bool     `json:"with_photos,omitempty"`
	Amenities    []string `json:"amenities,omitempty"`
}
