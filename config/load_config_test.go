package config_test

import (
	"reflect"
	"testing"

	"github.com/letsencrypt/sunlight-secretmanager/config"
)

func TestLoadConfigNoFile(t *testing.T) {
	t.Parallel()

	testFile := ""
	got, err := config.LoadConfigFromYaml(testFile)

	if got != nil || err == nil {
		t.Errorf("got %q and error %q, wanted error and nil error", got, err)
	}
}

func TestLoadConfigCorrect(t *testing.T) {
	t.Parallel()

	testFile := "../cmd/sunlight-secretmanager/sunlight.yaml"
	got, err := config.LoadConfigFromYaml(testFile)
	want := map[string]string{
		"rome.ct.filippo.io/2024h2": "/etc/sunlight/rome2024h2.key",
		"rome.ct.filippo.io/2025h1": "/etc/sunlight/rome2025h1.key",
		"rome.ct.filippo.io/2025h2": "/etc/sunlight/rome2025h2.key",
	}

	if !reflect.DeepEqual(got, want) || err != nil {
		t.Errorf("got %q and error %q, wanted nil and not nil error", got, err)
	}
}
