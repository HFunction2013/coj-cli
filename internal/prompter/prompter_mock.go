package prompter

// Mock is a hand-written stand-in for gh's moq-generated prompter mock.
// It records calls and returns whatever the test installed, so commands can
// be exercised without a terminal.
type Mock struct {
	SelectFunc                 func(string, string, []string) (int, error)
	SelectCalls                []SelectCall
	MultiSelectFunc            func(string, []string, []string) ([]int, error)
	MultiSelectCalls           []MultiSelectCall
	MultiSelectWithSearchFunc  func(string, string, []string, []string, func(string) MultiSelectSearchResult) ([]string, error)
	MultiSelectWithSearchCalls []MultiSelectWithSearchCall
	InputFunc                  func(string, string) (string, error)
	InputCalls                 []InputCall
	PasswordFunc               func(string) (string, error)
	PasswordCalls              []PasswordCall
	ConfirmFunc                func(string, bool) (bool, error)
	ConfirmCalls               []ConfirmCall
	ConfirmDeletionFunc        func(string) error
	ConfirmDeletionCalls       []string
}

type SelectCall struct {
	Prompt  string
	Default string
	Options []string
}

type MultiSelectCall struct {
	Prompt   string
	Defaults []string
	Options  []string
}

type MultiSelectWithSearchCall struct {
	Prompt            string
	SearchPrompt      string
	Defaults          []string
	PersistentOptions []string
}

type InputCall struct {
	Prompt  string
	Default string
}

type PasswordCall struct {
	Prompt string
}

type ConfirmCall struct {
	Prompt  string
	Default bool
}

func (m *Mock) Select(prompt, defaultValue string, options []string) (int, error) {
	m.SelectCalls = append(m.SelectCalls, SelectCall{prompt, defaultValue, options})
	if m.SelectFunc != nil {
		return m.SelectFunc(prompt, defaultValue, options)
	}
	return 0, nil
}

func (m *Mock) MultiSelect(prompt string, defaults, options []string) ([]int, error) {
	m.MultiSelectCalls = append(m.MultiSelectCalls, MultiSelectCall{prompt, defaults, options})
	if m.MultiSelectFunc != nil {
		return m.MultiSelectFunc(prompt, defaults, options)
	}
	return nil, nil
}

func (m *Mock) MultiSelectWithSearch(prompt, searchPrompt string, defaults, persistentOptions []string, searchFunc func(string) MultiSelectSearchResult) ([]string, error) {
	m.MultiSelectWithSearchCalls = append(m.MultiSelectWithSearchCalls,
		MultiSelectWithSearchCall{prompt, searchPrompt, defaults, persistentOptions})
	if m.MultiSelectWithSearchFunc != nil {
		return m.MultiSelectWithSearchFunc(prompt, searchPrompt, defaults, persistentOptions, searchFunc)
	}
	return defaults, nil
}

func (m *Mock) Input(prompt, defaultValue string) (string, error) {
	m.InputCalls = append(m.InputCalls, InputCall{prompt, defaultValue})
	if m.InputFunc != nil {
		return m.InputFunc(prompt, defaultValue)
	}
	return defaultValue, nil
}

func (m *Mock) Password(prompt string) (string, error) {
	m.PasswordCalls = append(m.PasswordCalls, PasswordCall{prompt})
	if m.PasswordFunc != nil {
		return m.PasswordFunc(prompt)
	}
	return "", nil
}

func (m *Mock) Confirm(prompt string, defaultValue bool) (bool, error) {
	m.ConfirmCalls = append(m.ConfirmCalls, ConfirmCall{prompt, defaultValue})
	if m.ConfirmFunc != nil {
		return m.ConfirmFunc(prompt, defaultValue)
	}
	return defaultValue, nil
}

func (m *Mock) ConfirmDeletion(requiredValue string) error {
	m.ConfirmDeletionCalls = append(m.ConfirmDeletionCalls, requiredValue)
	if m.ConfirmDeletionFunc != nil {
		return m.ConfirmDeletionFunc(requiredValue)
	}
	return nil
}

var _ Prompter = (*Mock)(nil)
