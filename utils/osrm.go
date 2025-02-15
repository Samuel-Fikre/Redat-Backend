package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type OSRMResponse struct {
	Code   string `json:"code"`
	Routes []struct {
		Distance float64 `json:"distance"`
		Duration float64 `json:"duration"`
		Geometry string  `json:"geometry"`
	} `json:"routes"`
}

type MapService struct {
	BaseURL string
}

func NewMapService() *MapService {
	return &MapService{
		BaseURL: "http://router.project-osrm.org", // Using public OSRM instance
	}
}

func (m *MapService) GetRoute(fromLng, fromLat, toLng, toLat float64) (*OSRMResponse, error) {
	url := fmt.Sprintf("%s/route/v1/driving/%f,%f;%f,%f?overview=full&geometries=geojson",
		m.BaseURL, fromLng, fromLat, toLng, toLat)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result OSRMResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
} 