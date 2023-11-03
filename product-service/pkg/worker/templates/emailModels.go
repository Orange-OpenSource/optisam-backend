package templates

type MaintennceEmailParams struct {
	ProductName      string
	SKU              string
	EndOfMaintenance string
	Scope            string
}

type Data struct {
	Type          string
	EmailTemplate string
	Items         []MaintennceEmailParams
}
