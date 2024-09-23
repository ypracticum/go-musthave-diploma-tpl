package utils

import (
	"encoding/json"
	"time"
)

type RFC3339Date struct {
	time.Time
}

func (d *RFC3339Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Time.Format(time.RFC3339))
}

func (d *RFC3339Date) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	t, err := time.Parse(time.RFC3339, str)
	if err != nil {
		return err
	}
	d.Time = t
	return nil
}
