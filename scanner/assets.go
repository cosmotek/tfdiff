package scanner

import (
	"github.com/gocarina/gocsv"
	"github.com/samber/lo"
)

type Asset struct {
	Identifier   string         `csv:"id"`
	AccountID    string         `csv:"account_id"`
	Region       string         `csv:"region"`
	Service      string         `csv:"service"`
	ResourceType string         `csv:"resource_type"`
	Metadata     map[string]any `csv:"metadata"`
}

type AssetList []Asset

func (r AssetList) Len() int {
	return len(r)
}
func (r AssetList) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}
func (r AssetList) Less(i, j int) bool {
	return len(r[i].Identifier) < len(r[j].Identifier)
}

func (r AssetList) ToCSV() (string, error) {
	return gocsv.MarshalString(r)
}

type AssetCounterEntries []lo.Entry[string, []Asset]

func (r AssetCounterEntries) Len() int {
	return len(r)
}
func (r AssetCounterEntries) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}
func (r AssetCounterEntries) Less(i, j int) bool {
	return len(r[i].Value) > len(r[j].Value)
}
