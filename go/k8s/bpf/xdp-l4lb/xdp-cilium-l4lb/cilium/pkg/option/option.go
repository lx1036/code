package option

import (
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/lock"
)

// OptionSetting specifies the different choices each Option has.
type OptionSetting int

const (
    OptionDisabled OptionSetting = iota
    OptionEnabled
)

// VerifyFunc validates option key with value and may return an error if the
// option should not be applied
type VerifyFunc func(key string, value string) error

// ParseFunc parses the option value and may return an error if the option
// cannot be parsed or applied.
type ParseFunc func(value string) (OptionSetting, error)

// FormatFunc formats the specified value as textual representation option.
type FormatFunc func(value OptionSetting) string

// Option is the structure used to specify the semantics of a configurable
// boolean option
type Option struct {
    // Define is the name of the #define used for BPF programs
    Define string
    // Description is a short human readable description
    Description string
    // Immutable marks an option which is read-only
    Immutable bool
    // Requires is a list of required options, such options will be
    // automatically enabled as required.
    Requires []string
    // Parse is called to parse the option. If not specified, defaults to
    // NormalizeBool().
    Parse ParseFunc
    // FormatFunc is called to format the value for an option. If not
    // specified, defaults to formatting 0 as "Disabled" and other values
    // as "Enabled".
    Format FormatFunc
    // Verify is called prior to applying the option
    Verify VerifyFunc
}

// IntOptions member functions with external access do not require
// locking by the caller, while functions with internal access presume
// the caller to have taken care of any locking needed.
type IntOptions struct {
    optsMU  lock.RWMutex   // Protects all variables from this structure below this line
    Opts    OptionMap      `json:"map"`
    Library *OptionLibrary `json:"-"`
}

type OptionMap map[string]OptionSetting

type OptionLibrary map[string]*Option

func (o *IntOptions) getValue(key string) OptionSetting {
    value, exists := o.Opts[key]
    if !exists {
        return OptionDisabled
    }
    return value
}

func (o *IntOptions) GetValue(key string) OptionSetting {
    o.optsMU.RLock()
    v := o.getValue(key)
    o.optsMU.RUnlock()
    return v
}

// SetValidated sets the option `key` to the specified value. The caller is
// expected to have validated the input to this function.
func (o *IntOptions) SetValidated(key string, value OptionSetting) {
    o.optsMU.Lock()
    o.Opts[key] = value
    o.optsMU.Unlock()
}

func (o *IntOptions) IsEnabled(key string) bool {
    return o.GetValue(key) != OptionDisabled
}

func NewIntOptions(lib *OptionLibrary) *IntOptions {
    return &IntOptions{
        Opts:    OptionMap{},
        Library: lib,
    }
}
