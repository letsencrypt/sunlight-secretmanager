package config

import (
	"reflect"
	"testing"
)

func TestLoadConfigNoFile(t *testing.T) {
	t.Parallel()

	testFile := ""
	gotSeeds, gotFiles, err := LoadConfigFromYaml(testFile)

	if gotSeeds != nil || gotFiles != nil || err == nil {
		t.Errorf("got %q and error %q, wanted error and nil error", gotSeeds, err)
	}
}

func TestLoadConfigCorrect(t *testing.T) {
	t.Parallel()

	testFile := "sunlight.yaml"
	gotSeeds, gotFiles, err := LoadConfigFromYaml(testFile)
	wantFiles := map[string]FileType{
		"rome.ct.filippo.io/2025h1": {"/etc/radiantlog-twig.ct.letsencrypt.org-2025h1b.key", "radiantlog-twig.ct.letsencrypt.org-2025h1b.key"},
		"rome.ct.filippo.io/2025h2": {"radiantlog-twig.ct.letsencrypt.org-2025h2b.key", "radiantlog-twig.ct.letsencrypt.org-2025h2b.key"},
	}
	wantSeeds := map[string]string{
		"rome.ct.filippo.io/2025h1": "radiantlog-twig.ct.letsencrypt.org-2025h1b.key",
		"rome.ct.filippo.io/2025h2": "radiantlog-twig.ct.letsencrypt.org-2025h2b.key",
	}

	if !reflect.DeepEqual(gotSeeds, wantSeeds) || !reflect.DeepEqual(gotFiles, wantFiles) || err != nil {
		t.Errorf("got %q and error %q, wanted nil and not nil error", gotSeeds, err)
	}
}
