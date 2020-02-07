package positions

import (
	_ "github.com/mailru/easyjson/gen"
	"time"
)

type Position struct {
	Keyword  string    `json:"keyword"`
	Position uint64    `json:"position"`
	Url      string    `json:"url"`
	Volume   uint64    `json:"volume"`
	Results  uint64    `json:"results"`
	Cpc      float64   `json:"cpc"`
	Updated  time.Time `json:"updated"`
}
