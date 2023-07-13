package scanner

type EnvironmentScanner interface {
	RunScan() (AssetList, error)
}
