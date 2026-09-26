package prompter

import (
	"fmt"
	"slices"
)

// multiSelectWithSearch mirrors gh's implementation in
// internal/prompter/prompter.go: it keeps a "Search" sentinel at the top of
// the list, remembers what has already been selected, and re-runs searchFunc
// whenever the user asks to search again.
//
// Why this instead of loading everything once:
// the option list is dynamic, so callers get back keys (values), not indices.
// Selected items survive later searches, and persistent options are always
// shown so the user never loses the thing they just picked.
func multiSelectWithSearch(p Prompter, prompt, searchPrompt string, defaultValues, persistentValues []string, searchFunc func(string) MultiSelectSearchResult) ([]string, error) {
	selectedOptions := slices.Clone(defaultValues)

	// optionKeyLabels uniquely identifies options and provides optional
	// display labels.
	optionKeyLabels := make(map[string]string)
	for _, k := range selectedOptions {
		optionKeyLabels[k] = k
	}

	searchResult := searchFunc("")
	if searchResult.Err != nil {
		return nil, fmt.Errorf("failed to search: %w", searchResult.Err)
	}
	searchResultKeys := searchResult.Keys
	moreResults := searchResult.MoreResults

	for i, k := range searchResultKeys {
		if i < len(searchResult.Labels) {
			optionKeyLabels[k] = searchResult.Labels[i]
		}
	}

	for {
		// Build the dynamic option list:
		// search sentinel, selections, search results, persistent options.
		optionKeys := make([]string, 0, 1+len(selectedOptions)+len(searchResultKeys)+len(persistentValues))
		optionLabels := make([]string, 0, len(optionKeys))

		// 1. Search sentinel.
		optionKeys = append(optionKeys, "")
		if moreResults > 0 {
			optionLabels = append(optionLabels, fmt.Sprintf("Search (%d more)", moreResults))
		} else {
			optionLabels = append(optionLabels, "Search")
		}

		// 2. Selections.
		for _, k := range selectedOptions {
			optionKeys = append(optionKeys, k)
			optionLabels = append(optionLabels, labelFor(optionKeyLabels, k))
		}

		// 3. Search results.
		for _, k := range searchResultKeys {
			// Already selected or persistent — adding it again would duplicate.
			if slices.Contains(selectedOptions, k) || slices.Contains(persistentValues, k) {
				continue
			}
			optionKeys = append(optionKeys, k)
			optionLabels = append(optionLabels, labelFor(optionKeyLabels, k))
		}

		// 4. Persistent options.
		for _, k := range persistentValues {
			if slices.Contains(selectedOptions, k) {
				continue
			}
			optionKeys = append(optionKeys, k)
			optionLabels = append(optionLabels, labelFor(optionKeyLabels, k))
		}

		selectedOptionLabels := make([]string, len(selectedOptions))
		for i, k := range selectedOptions {
			selectedOptionLabels[i] = labelFor(optionKeyLabels, k)
		}

		selectedIdxs, err := p.MultiSelect(prompt, selectedOptionLabels, optionLabels)
		if err != nil {
			return nil, err
		}

		pickedSearch := false
		var newSelectedOptions []string
		for _, idx := range selectedIdxs {
			if idx == 0 { // Search sentinel selected
				pickedSearch = true
				continue
			}
			if idx < 0 || idx >= len(optionKeys) {
				continue
			}
			key := optionKeys[idx]
			if key == "" {
				continue
			}
			newSelectedOptions = append(newSelectedOptions, key)
		}

		selectedOptions = newSelectedOptions
		for _, k := range selectedOptions {
			if _, ok := optionKeyLabels[k]; !ok {
				optionKeyLabels[k] = k
			}
		}

		if pickedSearch {
			query, err := p.Input(searchPrompt, "")
			if err != nil {
				return nil, err
			}
			searchResult := searchFunc(query)
			if searchResult.Err != nil {
				return nil, searchResult.Err
			}
			searchResultKeys = searchResult.Keys
			moreResults = searchResult.MoreResults
			for i, k := range searchResultKeys {
				if i < len(searchResult.Labels) {
					optionKeyLabels[k] = searchResult.Labels[i]
				}
			}
			continue
		}

		return selectedOptions, nil
	}
}

func labelFor(m map[string]string, k string) string {
	if l, ok := m[k]; ok && l != "" {
		return l
	}
	return k
}
