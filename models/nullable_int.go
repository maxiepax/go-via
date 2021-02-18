package models

import (
	"database/sql"
	"encoding/json"
	"strconv"
)

type NullInt32 struct {
	sql.NullInt32
}

// NullInt64 MarshalJSON interface redefinition
func (r NullInt32) MarshalJSON() ([]byte, error) {
	if r.Valid {
		return json.Marshal(r.Int32)
	} else {
		return json.Marshal(nil)
	}
}

func (ns *NullInt32) UnmarshalJSON(text []byte) error {
	txt := string(text)
	ns.Valid = true
	if txt == "null" {
		ns.Valid = false
		return nil
	}
	i, err := strconv.ParseInt(txt, 10, 32)
	if err != nil {
		ns.Valid = false
		return err
	}
	j := int32(i)
	ns.Int32 = j
	return nil
}
