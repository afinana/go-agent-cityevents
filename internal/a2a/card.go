package a2a

import "encoding/json"

// AgentCard represents the Google A2A Protocol Agent Card definition
type AgentCard struct {
	Name         string      `json:"name"`
	Description  string      `json:"description"`
	Role         string      `json:"role"`
	Capabilities []string    `json:"capabilities"`
	Interfaces   []Interface `json:"interfaces"`
}

type Interface struct {
	Type        string `json:"type"`
	Endpoint    string `json:"endpoint"`
	Method      string `json:"method"`
	Description string `json:"description"`
	Payload     any    `json:"payload,omitempty"`
}

// GetAgentCard returns the A2A definition for this agent
func GetAgentCard() AgentCard {
	return AgentCard{
		Name:        "MadridCityEventsAgent",
		Description: "An AI agent that provides semantic search capabilities over Madrid cultural events using a vector database.",
		Role:        "Data Retrieval",
		Capabilities: []string{
			"semantic_search",
			"event_discovery",
			"madrid_culture",
		},
		Interfaces: []Interface{
			{
				Type:        "REST",
				Endpoint:    "/api/search",
				Method:      "POST",
				Description: "Accepts a natural language query and returns relevant Madrid cultural events.",
				Payload: map[string]string{
					"query": "string (The natural language query)",
				},
			},
		},
	}
}

// ToJSON converts the agent card into a pretty-printed JSON byte slice
func (c AgentCard) ToJSON() ([]byte, error) {
	return json.MarshalIndent(c, "", "  ")
}
