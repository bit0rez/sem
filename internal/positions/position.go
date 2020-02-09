package positions

import (
	"time"

	_ "github.com/mailru/easyjson/gen"
)

// Position field type
type Updated time.Time

func (u *Updated) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, 12)
	b = append(b, '"')
	b = append(b, []byte(time.Time(*u).Format("2006-01-02"))...)
	b = append(b, '"')

	return b, nil
}

func (u *Updated) UnmarshalJSON(data []byte) error {
	var (
		t   time.Time
		err error
	)

	if len(data) == 12 {
		t, err = time.Parse("2006-01-02", string(data[1:11]))
		if err != nil {
			return err
		}
	} else {
		t := time.Time{}
		if err = t.UnmarshalJSON(data); err != nil {
			return err
		}
	}

	*u = Updated(t)

	return nil
}

type Position struct {
	Keyword  string   `json:"keyword"`
	Position uint64   `json:"position"`
	Url      string   `json:"url"`
	Volume   uint64   `json:"volume"`
	Results  uint64   `json:"results"`
	Cpc      float64  `json:"cpc"`
	Updated  *Updated `json:"updated"`
}
