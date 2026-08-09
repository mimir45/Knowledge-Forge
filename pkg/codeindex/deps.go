package codeindex

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
)

// Deps reads the declared dependency versions from whichever build manifests the repo
// has. Drift's fifth verdict — "declared dep version bumped, the note may describe old
// behaviour" — is a comparison between two of these maps, so the map is the unit, not
// the individual file.
func Deps(root string) map[string]string {
	out := map[string]string{}
	mergeMaven(out, filepath.Join(root, "pom.xml"))
	mergeGradle(out, filepath.Join(root, "build.gradle"))
	mergeGradle(out, filepath.Join(root, "build.gradle.kts"))
	mergeNPM(out, filepath.Join(root, "package.json"))
	return out
}

// Maven dependencies are read by regex, not by an XML model. The map only needs
// groupId:artifactId -> version; parsing the full POM (profiles, parents, property
// interpolation) is a resolver's job and this is a change detector.
var mavenDep = regexp.MustCompile(
	`(?s)<dependency>.*?<groupId>([^<]+)</groupId>.*?<artifactId>([^<]+)</artifactId>` +
		`(?:.*?<version>([^<]+)</version>)?.*?</dependency>`)

func mergeMaven(out map[string]string, path string) {
	src, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, m := range mavenDep.FindAllStringSubmatch(string(src), -1) {
		v := m[3]
		if v == "" {
			v = "managed" // inherited from a BOM; still worth recording as present
		}
		out[m[1]+":"+m[2]] = v
	}
}

var gradleDep = regexp.MustCompile(`['"]([\w.\-]+):([\w.\-]+):([\w.\-]+)['"]`)

func mergeGradle(out map[string]string, path string) {
	src, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, m := range gradleDep.FindAllStringSubmatch(string(src), -1) {
		out[m[1]+":"+m[2]] = m[3]
	}
}

func mergeNPM(out map[string]string, path string) {
	src, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var pkg struct {
		Deps    map[string]string `json:"dependencies"`
		DevDeps map[string]string `json:"devDependencies"`
	}
	if json.Unmarshal(src, &pkg) != nil {
		return
	}
	for _, m := range []map[string]string{pkg.Deps, pkg.DevDeps} {
		for k, v := range m {
			out[k] = v
		}
	}
}
