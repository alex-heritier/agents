package main

import "sort"

func getProviderNames(cfg ProvidersConfig, specSelector func(ProviderConfig) *FileSpec) []string {
	names := make([]string, 0, len(cfg.Providers))
	for name, provider := range cfg.Providers {
		if specSelector != nil && specSelector(provider) == nil {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func getProviderFlagName(cfg ProvidersConfig, providerName string) string {
	provider, ok := cfg.Providers[providerName]
	if !ok {
		return providerName
	}
	if provider.Name != "" {
		return provider.Name
	}
	return providerName
}
