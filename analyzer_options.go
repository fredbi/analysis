package analyzer

type analyzeOptions struct {
	WithExtensions bool
}

type analyzeOption func(*analyzeOptions)

func analyzeOptionsApplied(options []analyzeOptions) *analyzeOptions {
	opts := &analyzeOptions{}
	for _, apply := range options {
		apply(opts)
	}

	return opts
}
