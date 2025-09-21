package helper

import "time"

func ParseDateToMilli(dateStr string, isEnd bool) (int64, error) {
	if dateStr == "" {
		return 0, nil
	}

	loc, _ := time.LoadLocation("Asia/Jakarta") // pakai timezone lokal
	t, err := time.ParseInLocation("2006-01-02", dateStr, loc)
	if err != nil {
		return 0, err
	}

	if isEnd {
		// akhir hari lokal
		t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second + 999*time.Millisecond)
	}

	return t.UnixMilli(), nil
}
