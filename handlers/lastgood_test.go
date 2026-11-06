package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/rezuscloud/platform-website/obs"
)

type flakyClient struct {
	calls int
	data  obs.LiveData
	err   bool
}

func (f *flakyClient) Fetch(_ context.Context) (obs.LiveData, error) {
	f.calls++
	if f.err {
		return obs.LiveData{}, errors.New("boom")
	}
	return f.data, nil
}

func TestLastGoodClientServesStaleOnError(t *testing.T) {
	good := obs.LiveData{Hosts: []obs.Host{{Name: "n1"}}}
	fc := &flakyClient{data: good}
	lg := &lastGoodClient{inner: fc}

	if _, err := lg.Fetch(context.Background()); err != nil {
		t.Fatal(err)
	}
	fc.err = true
	got, err := lg.Fetch(context.Background())
	if err != nil {
		t.Fatalf("stale fetch must not error: %v", err)
	}
	if len(got.Hosts) != 1 || got.Hosts[0].Name != "n1" {
		t.Fatalf("stale data not served: %+v", got)
	}
}

func TestLastGoodClientServesStaleOnEmpty(t *testing.T) {
	good := obs.LiveData{Services: []obs.Service{{Name: "s1"}}}
	fc := &flakyClient{data: good}
	lg := &lastGoodClient{inner: fc}
	lg.Fetch(context.Background())
	fc.data = obs.LiveData{} // empty but no error
	got, err := lg.Fetch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Services) != 1 {
		t.Fatalf("empty fetch must fall back to last good, got %+v", got)
	}
}
