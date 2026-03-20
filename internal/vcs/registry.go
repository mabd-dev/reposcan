package vcs

type Registry struct {
	providers map[Type]Provider
}

func NewRegistry(providers ...Provider) *Registry {
	registry := &Registry{
		providers: map[Type]Provider{},
	}

	for _, provider := range providers {
		if provider == nil {
			continue
		}
		registry.Register(provider)
	}

	return registry
}

func (r *Registry) Register(provider Provider) {
	if r.providers == nil {
		r.providers = map[Type]Provider{}
	}

	r.providers[provider.Type()] = provider
}

func (r *Registry) Get(repoType Type) (Provider, bool) {
	if r == nil {
		return nil, false
	}

	provider, ok := r.providers[repoType]
	return provider, ok
}
