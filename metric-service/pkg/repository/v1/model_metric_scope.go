package v1

// Metadata for injectors
type ScopeMetric struct {
	ID         string
	MetricType string
	MetricName string
	Reference  string
	Core       string
	CPU        string
	Scope      string
	Default    bool
}

func GetScopeMetric(scope string) []ScopeMetric {
	resp := []ScopeMetric{
		{
			MetricType: "microsoft.sql.enterprise",
			MetricName: "microsoft.sql.enterprise.2019",
			Reference:  "server",
			Core:       "cores_per_processor",
			CPU:        "server_processors_numbers",
			Scope:      scope,
			Default:    true,
		},
		{
			MetricType: "windows.server.datacenter",
			MetricName: "windows.server.datacenter.2016",
			Reference:  "server",
			Core:       "cores_per_processor",
			CPU:        "server_processors_numbers",
			Scope:      scope,
			Default:    true,
		},
	}

	return resp
}
