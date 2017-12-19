package core

import (
	"os"
	"testing"
)

func TestGetObjectKey(t *testing.T) {
	host, _ := os.Hostname()
	cases := []struct {
		desc   string
		key    string
		format string
		seq    int
		out    string
	}{
		{"base", "20170224173000", "{{.Output}}/{{.Year}}/{{.Month}}/{{.Day}}/{{.Hostname}}-{{.Year}}{{.Month}}{{.Day}}{{.Hour}}{{.Minute}}_{{.Seq}}.log.gz", 1, "output/2017/02/24/" + host + "-201702241730_1.log.gz"},
		{"padding day", "20170204173000", "{{.Day}}", 1, "04"},
	}

	e, _ := NewEpoch("", "", "output")
	for _, c := range cases {
		e.epochKey = c.key
		e.keyFmt = c.format
		s, err := e.GetObjectKey(c.seq)
		if err != nil {
			t.Errorf("%q: GetObjectKey(%d) => error: %q", c.desc, c.seq, err)

		} else if s != c.out {
			t.Errorf("%q: GetObjectKey(%d) => %s, wants %q", c.desc, c.seq, s, c.out)
		}
	}
}

func TestParseEpoch(t *testing.T) {
	etests := []struct {
		desc string
		in   string
		out  string
	}{
		{"base", "time:24/Feb/2017:10:00:07 +0900\thost:127.0.0.1", "20170224100000"},
		{"just before begin carried", "time:24/Feb/2017:10:29:59 +0900\thost:127.0.0.1", "20170224100000"},
		{"just after begin carried", "time:24/Feb/2017:10:30:00 +0900\thost:127.0.0.1", "20170224103000"},
		{"time is after host", "host:127.0.0.1\ttime:24/Feb/2017:10:00:00 +0900", "20170224100000"},
		{"time is closed in the bracket", "host:127.0.0.1\ttime:[24/Feb/2017:10:00:00 +0900]", "20170224100000"},
	}
	for _, tt := range etests {
		s := parseEpoch(tt.in, "tsv", 30)
		if s != tt.out {
			t.Errorf("%q: parseEpoch(%q) => %s, wants %q", tt.desc, tt.in, s, tt.out)
		}
	}
}
