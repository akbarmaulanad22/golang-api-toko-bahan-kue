package helper

import "time"

func ParseDateToMilli(dateStr string, isEnd bool) (int64, error) {
	if dateStr == "" {
		return 0, nil // kosong → ga dipakai
	}

	// format sesuai input user, misal "2006-01-02"
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return 0, err
	}

	if isEnd {
		// supaya end_date jam 23:59:59
		t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	}

	return t.UnixMilli(), nil
}
