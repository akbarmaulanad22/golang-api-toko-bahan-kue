package model

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type UnixDate int64

func (ud *UnixDate) UnmarshalJSON(data []byte) error {
	// Buat variabel sementara
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	switch val := v.(type) {
	case string:
		t, err := time.Parse("2006-01-02", val)
		if err != nil {
			return fmt.Errorf("invalid date format: %w", err)
		}
		*ud = UnixDate(t.UnixMilli())
	case float64: // timestamp
		*ud = UnixDate(int64(val))
	default:
		return fmt.Errorf("unsupported date type: %T", v)
	}
	return nil
}

func (ud *UnixDate) ParseFromString(s string) error {
	if s == "" {
		return nil
	}

	// coba parse timestamp (int64)
	if ts, err := strconv.ParseInt(s, 10, 64); err == nil {
		*ud = UnixDate(ts)
		return nil
	}

	// coba parse format tanggal
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return fmt.Errorf("invalid date format: %w", err)
	}
	*ud = UnixDate(t.UnixMilli())
	return nil
}
