package config

import (
	"reflect"
	"testing"
)

func TestLoadConfigNoFile(t *testing.T) {
	t.Parallel()

	testFile := ""
	got, err := LoadConfigFromYaml(testFile)

	if got != nil || err == nil {
		t.Errorf("got %q and error %q, wanted error and nil error", got, err)
	}
}

func TestLoadConfigCorrect(t *testing.T) {
	t.Parallel()

	testFile := "sunlight.yaml"
	got, err := LoadConfigFromYaml(testFile)
	want := map[string]string{
		"rome.ct.filippo.io/2025h1": "radiantlog-twig.ct.letsencrypt.org-2025h1b.key",
		"rome.ct.filippo.io/2025h2": "radiantlog-twig.ct.letsencrypt.org-2025h2b.key",
	}

	if !reflect.DeepEqual(got, want) || err != nil {
		t.Errorf("got %q and error %q, wanted nil and not nil error", got, err)
	}
}
