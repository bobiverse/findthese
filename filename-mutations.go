package main

import (
	"path/filepath"
	"strings"
)

// Filenames that points there could be more similar files
var fnamesSimilar = map[string][]string{
	"composer.json": []string{"composer.lock", "composer.phar"},
	".htaccess":     []string{".htpasswd"},
	"Dockerfile": []string{
		"Dockerfile.production",
		"Dockerfile.prod",
		"Dockerfile.dev",
		"Dockerfile.local",
		"Dockerfile.loc",
		"docker-compose.yml",
		".env",
	},
}

// Generate list of file mutations
// given argument can be single filename [file.txt]
// or path [path/to/file.txt]
func filePathMutations(fpath string, patterns []string) []string {
	fname := filepath.Base(fpath)
	basedir := filepath.Dir(fpath)

	var mutations []string
	mutations = append(mutations, fpath) // keep original

	// Append file names that are similar or related to this
	if similars, haveSimilar := fnamesSimilar[fname]; haveSimilar {
		for _, sim := range similars {
			sim = filepath.Join(basedir, sim)
			mutations = append(mutations, sim)
		}
	}

	// go and mutate!
	for _, pattern := range patterns {
		smut := fname + pattern // as suffix

		// replace asterisk with fname
		if strings.Contains(pattern, "*") {
			smut = strings.Replace(pattern, "*", fname, 1)
		}

		smut = filepath.Join(filepath.Dir(fpath), smut)
		mutations = append(mutations, smut)
	}

	// color.Red("MUT: %v", mutations)
	return mutations
}
