package experiment

const VERSION = "0.4.3"

type User struct {
	UserId             string                 `json:"user_id,omitempty"`
	DeviceId           string                 `json:"device_id,omitempty"`
	Country            string                 `json:"country,omitempty"`
	Region             string                 `json:"region,omitempty"`
	Dma                string                 `json:"dma,omitempty"`
	City               string                 `json:"city,omitempty"`
	Language           string                 `json:"language,omitempty"`
	Platform           string                 `json:"platform,omitempty"`
	Version            string                 `json:"version,omitempty"`
	Os                 string                 `json:"os,omitempty"`
	DeviceManufacturer string                 `json:"device_manufacturer,omitempty"`
	DeviceBrand        string                 `json:"device_brand,omitempty"`
	DeviceModel        string                 `json:"device_model,omitempty"`
	Carrier            string                 `json:"carrier,omitempty"`
	Library            string                 `json:"library,omitempty"`
	UserProperties     map[string]interface{} `json:"user_properties,omitempty"`
}

type Variant struct {
	Value   string      `json:"value,omitempty"`
	Payload interface{} `json:"payload,omitempty"`
}
