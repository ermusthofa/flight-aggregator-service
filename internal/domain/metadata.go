package domain

type Metadata struct {
	TotalResults       int
	ProvidersQueried   int
	ProvidersSucceeded int
	ProvidersFailed    int
	SearchTimeMs       int64
	CacheHit           bool
}
