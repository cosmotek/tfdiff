package scanner

import "context"

type EnvironmentScanner interface {
	RunScan(context.Context) (AssetList, error)
}
