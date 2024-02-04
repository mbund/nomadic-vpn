package provider

type Location struct {
	Lat, Long float32
	Code      string
	City      string
	Country   string
	Plans     map[string]Plan
}

type Plan struct {
	Price    float32
	Provider Provider
}

type Provider interface {
	// Create a new VPS instance
	//
	// Returns the IP address of the new instance
	CreateInstance(cloudConfig string) (string, error)

	// Destroy a VPS instance
	DestroyInstance(ip string) error
}

func InitializeProviders() {
	initializeVultr()
}

var Locations = map[string]Location{
	"ams": {
		City:    "Amsterdam",
		Country: "Netherlands",
		Lat:     0,
		Long:    0,
		Plans:   make(map[string]Plan),
	},
	"atl": {
		City:    "Atlanta",
		Country: "United States",
		Lat:     0,
		Long:    0,
		Plans:   make(map[string]Plan),
	},
	"blr": {
		City:    "Bangalore",
		Country: "India",
		Lat:     0,
		Long:    0,
		Plans:   make(map[string]Plan),
	},
	"ord": {
		City:    "Chicago",
		Country: "United States",
		Lat:     41.88425,
		Long:    -87.63245,
		Plans:   make(map[string]Plan),
	},
	"dfw": {
		City:    "Dallas",
		Country: "United States",
		Lat:     0,
		Long:    0,
		Plans:   make(map[string]Plan),
	},
	"del": {
		City:    "Delhi NCR",
		Country: "India",
		Lat:     0,
		Long:    0,
		Plans:   make(map[string]Plan),
	},
	"fra": {
		City:    "Frankfurt",
		Country: "Germany",
		Lat:     0,
		Long:    0,
		Plans:   make(map[string]Plan),
	},
	"hnl": {
		City:    "Honolulu",
		Country: "United States",
		Lat:     0,
		Long:    0,
		Plans:   make(map[string]Plan),
	},
	"jnb": {
		City:    "Johannesburg",
		Country: "South Africa",
		Lat:     0,
		Long:    0,
		Plans:   make(map[string]Plan),
	},
	"lhr": {
		City:    "London",
		Country: "United Kingdom",
		Lat:     0,
		Long:    0,
		Plans:   make(map[string]Plan),
	},
	"lax": {
		City:    "Los Angeles",
		Country: "United States",
		Lat:     0,
		Long:    0,
		Plans:   make(map[string]Plan),
	},
	"mad": {
		City:    "Madrid",
		Country: "Spain",
		Lat:     0,
		Long:    0,
		Plans:   make(map[string]Plan),
	},
	"man": {
		City:    "Manchester",
		Country: "United Kingdom",
		Lat:     0,
		Long:    0,
		Plans:   make(map[string]Plan),
	},
	"mel": {
		City:    "Melbourne",
		Country: "Australia",
		Lat:     0,
		Long:    0,
		Plans:   make(map[string]Plan),
	},
	"mex": {
		City:    "Mexico City",
		Country: "Mexico",
		Lat:     0,
		Long:    0,
		Plans:   make(map[string]Plan),
	},
	"mia": {
		City:    "Miami",
		Country: "United States",
		Lat:     0,
		Long:    0,
		Plans:   make(map[string]Plan),
	},
	"bom": {
		City:    "Mumbai",
		Country: "India",
		Lat:     0,
		Long:    0,
		Plans:   make(map[string]Plan),
	},
	"ewr": {
		City:    "New Jersey",
		Country: "United States",
		Lat:     0,
		Long:    0,
		Plans:   make(map[string]Plan),
	},
	"itm": {
		City:    "Osaka",
		Country: "Japan",
		Lat:     0,
		Long:    0,
		Plans:   make(map[string]Plan),
	},
	"cdg": {
		City:    "Paris",
		Country: "France",
		Lat:     0,
		Long:    0,
		Plans:   make(map[string]Plan),
	},
	"scl": {
		City:    "Santiago",
		Country: "Chile",
		Lat:     0,
		Long:    0,
		Plans:   make(map[string]Plan),
	},
	"sao": {
		City:    "S\u00e3o Paulo",
		Country: "Brazil",
		Lat:     0,
		Long:    0,
		Plans:   make(map[string]Plan),
	},
	"sea": {
		City:    "Seattle",
		Country: "United States",
		Lat:     0,
		Long:    0,
		Plans:   make(map[string]Plan),
	},
	"icn": {
		City:    "Seoul",
		Country: "Korea, Republic of",
		Lat:     0,
		Long:    0,
		Plans:   make(map[string]Plan),
	},
	"sjc": {
		City:    "Silicon Valley",
		Country: "United States",
		Lat:     0,
		Long:    0,
		Plans:   make(map[string]Plan),
	},
	"sgp": {
		City:    "Singapore",
		Country: "Singapore",
		Lat:     0,
		Long:    0,
		Plans:   make(map[string]Plan),
	},
	"sto": {
		City:    "Stockholm",
		Country: "Sweden",
		Lat:     0,
		Long:    0,
		Plans:   make(map[string]Plan),
	},
	"syd": {
		City:    "Sydney",
		Country: "Australia",
		Lat:     0,
		Long:    0,
		Plans:   make(map[string]Plan),
	},
	"tlv": {
		City:    "Tel Aviv",
		Country: "Israel",
		Lat:     0,
		Long:    0,
		Plans:   make(map[string]Plan),
	},
	"nrt": {
		City:    "Tokyo",
		Country: "Japan",
		Lat:     35.652832,
		Long:    139.839478,
		Plans:   make(map[string]Plan),
	},
	"yto": {
		City:    "Toronto",
		Country: "Canada",
		Lat:     0,
		Long:    0,
		Plans:   make(map[string]Plan),
	},
	"waw": {
		City:    "Warsaw",
		Country: "Poland",
		Lat:     0,
		Long:    0,
		Plans:   make(map[string]Plan),
	},
}
