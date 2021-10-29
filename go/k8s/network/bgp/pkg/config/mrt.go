package config

import "fmt"

// typedef for identity gobgp:mrt-type.
type MrtType string

const (
	MRT_TYPE_UPDATES MrtType = "updates"
	MRT_TYPE_TABLE   MrtType = "table"
)

var MrtTypeToIntMap = map[MrtType]int{
	MRT_TYPE_UPDATES: 0,
	MRT_TYPE_TABLE:   1,
}

var IntToMrtTypeMap = map[int]MrtType{
	0: MRT_TYPE_UPDATES,
	1: MRT_TYPE_TABLE,
}

func (v MrtType) Validate() error {
	if _, ok := MrtTypeToIntMap[v]; !ok {
		return fmt.Errorf("invalid MrtType: %s", v)
	}
	return nil
}

func (v MrtType) ToInt() int {
	i, ok := MrtTypeToIntMap[v]
	if !ok {
		return -1
	}
	return i
}

// struct for container gobgp:config.
type MrtConfig struct {
	// original -> gobgp:dump-type
	DumpType MrtType `mapstructure:"dump-type" json:"dump-type,omitempty"`
	// original -> gobgp:file-name
	// Configures a file name to be written.
	FileName string `mapstructure:"file-name" json:"file-name,omitempty"`
	// original -> gobgp:table-name
	// specify the table name with route server setup.
	TableName string `mapstructure:"table-name" json:"table-name,omitempty"`
	// original -> gobgp:dump-interval
	DumpInterval uint64 `mapstructure:"dump-interval" json:"dump-interval,omitempty"`
	// original -> gobgp:rotation-interval
	RotationInterval uint64 `mapstructure:"rotation-interval" json:"rotation-interval,omitempty"`
}

func (lhs *MrtConfig) Equal(rhs *MrtConfig) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if lhs.DumpType != rhs.DumpType {
		return false
	}
	if lhs.FileName != rhs.FileName {
		return false
	}
	if lhs.TableName != rhs.TableName {
		return false
	}
	if lhs.DumpInterval != rhs.DumpInterval {
		return false
	}
	if lhs.RotationInterval != rhs.RotationInterval {
		return false
	}
	return true
}

// struct for container gobgp:mrt.
type Mrt struct {
	// original -> gobgp:file-name
	// original -> gobgp:mrt-config
	Config MrtConfig `mapstructure:"config" json:"config,omitempty"`
}

func (lhs *Mrt) Equal(rhs *Mrt) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if !lhs.Config.Equal(&(rhs.Config)) {
		return false
	}
	return true
}
