package positions

import (
	"strings"
	"testing"
	"time"
)

func TestUpdated_MarshalJSON(t *testing.T) {
	u := &Updated{}
	expected := strings.Join([]string{"\"", time.Time(*u).Format("2006-01-02"), "\""}, "")
	actual, err := u.MarshalJSON()
	if err != nil {
		t.Error(err)
	}
	if expected != string(actual) {
		t.Errorf("Marshal failed: returned %s when expected %s", actual, expected)
	}
}

func TestUpdated_UnmarshalJSON(t *testing.T) {
	u := &Updated{}
	r := []byte("\"2020-02-02\"")
	expected := time.Date(2020, time.February, 2, 0, 0, 0, 0, time.UTC).String()
	if err := u.UnmarshalJSON(r); err != nil {
		t.Fatal(err)
	}
	actual := time.Time(*u).String()
	if actual != expected {
		t.Errorf("Unmarshal failed: returned %s when expected %s", actual, expected)
	}
}
