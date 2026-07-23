package protocol

import (
	"fmt"
)

// Type prefix constants for RESP protocol
const (
	TypeSimpleString = '+'
	TypeError        = '-'
	TypeInteger      = ':'
	TypeBulkString   = '$'
	TypeArray        = '*'
)

// Value represents a parsed or serializable RESP payload.
type Value struct {
	Type   byte    // Type marker (+, -, :, $, *)
	Str    string  // Simple Strings & Errors
	Num    int64   // Integers
	Bulk   []byte  // Bulk Strings (raw binary bytes)
	Array  []Value // Arrays of RESP values
	IsNull bool    // Set to true for Null Bulk Strings ($-1\r\n) or Null Arrays (*-1\r\n)
}


// Value receiver
func (v Value) String() string {
	switch v.Type {
	case TypeSimpleString, TypeError:
		return v.Str
	case TypeInteger:
		return fmt.Sprintf("%d", v.Num)
	case TypeBulkString:
		if v.IsNull {
			return "(nil)"
		}
		return string(v.Bulk)
	case TypeArray:
		if v.IsNull {
			return "(nil)"
		}
		return fmt.Sprintf("%v", v.Array)
	default:
		return ""
	}
}