package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type Voice struct {
	VoiceID   string       `json:"voice_id"`
	VoiceName string       `json:"voice_name"`
	Models    []VoiceModel `json:"models"`
	Gender    string       `json:"gender"`
	Age       string       `json:"age"`
	UseCases  []string     `json:"use_cases"`
}

// UnmarshalJSON accepts both the legacy string and V3's localized voice_name.
func (v *Voice) UnmarshalJSON(data []byte) error {
	type fields struct {
		VoiceID  string       `json:"voice_id"`
		Models   []VoiceModel `json:"models"`
		Gender   string       `json:"gender"`
		Age      string       `json:"age"`
		UseCases []string     `json:"use_cases"`
	}
	var raw struct {
		VoiceName json.RawMessage `json:"voice_name"`
		*fields
	}
	raw.fields = &fields{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	name := ""
	if err := json.Unmarshal(raw.VoiceName, &name); err != nil {
		var localized map[string]string
		if err := json.Unmarshal(raw.VoiceName, &localized); err != nil {
			return err
		}
		name = localized["eng"]
		if name == "" {
			name = localized["kor"]
		}
		for _, value := range localized {
			if name == "" {
				name = value
			}
		}
	}

	v.VoiceID, v.VoiceName, v.Models = raw.VoiceID, name, raw.Models
	v.Gender, v.Age, v.UseCases = raw.Gender, raw.Age, raw.UseCases
	return nil
}

type VoiceModel struct {
	Version  string   `json:"version"`
	Emotions []string `json:"emotions"`
}

type ListVoicesParams struct {
	Model   string
	Gender  string
	Age     string
	UseCase string
}

type RecommendedVoice struct {
	VoiceID   string  `json:"voice_id"`
	VoiceName string  `json:"voice_name"`
	Score     float64 `json:"score"`
}

type RecommendVoicesParams struct {
	Query string
	Count int
}

func (c *Client) ListVoices(p ListVoicesParams) ([]Voice, error) {
	q := url.Values{}
	if p.Model != "" {
		q.Set("model", p.Model)
	}
	if p.Gender != "" {
		q.Set("gender", p.Gender)
	}
	if p.Age != "" {
		q.Set("age", p.Age)
	}
	if p.UseCase != "" {
		q.Set("use_cases", p.UseCase)
	}

	path := "/v3/voices"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	data, err := c.get(path)
	if err != nil {
		return nil, err
	}

	var voices []Voice
	if err := json.Unmarshal(data, &voices); err != nil {
		return nil, err
	}

	return voices, nil
}

func (c *Client) GetVoice(voiceID string) (*Voice, error) {
	data, err := c.get("/v3/voices/" + url.PathEscape(voiceID))
	if err != nil {
		return nil, err
	}

	var voice Voice
	if err := json.Unmarshal(data, &voice); err != nil {
		return nil, err
	}

	return &voice, nil
}

func (c *Client) RecommendVoices(p RecommendVoicesParams) ([]RecommendedVoice, error) {
	query := strings.TrimSpace(p.Query)
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}
	if p.Count == 0 {
		p.Count = 5
	}
	if p.Count < 1 || p.Count > 10 {
		return nil, fmt.Errorf("count must be between 1 and 10")
	}

	q := url.Values{}
	q.Set("query", query)
	q.Set("count", strconv.Itoa(p.Count))

	data, err := c.get("/v1/voices/recommendations?" + q.Encode())
	if err != nil {
		return nil, err
	}

	var voices []RecommendedVoice
	if err := json.Unmarshal(data, &voices); err != nil {
		return nil, err
	}

	return voices, nil
}
