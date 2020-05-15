package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"text/template"
)

// generate looks for migration sql in a directory for the given dialect and
// applies the templates below to the contents of the files, building up a
// migrations map for the dialect
func generate(dialect string) {
	baseDir := os.Getenv("GEN_BASEPATH") + fmt.Sprint("/internal/db/migrations")
	dir, err := os.Open(fmt.Sprintf("%s/%s", baseDir, dialect))
	if err != nil {
		fmt.Printf("error opening dir with dialect %s: %v\n", dialect, err)
		os.Exit(1)
	}
	names, err := dir.Readdirnames(0)
	if err != nil {
		fmt.Printf("error reading dir names with dialect %s: %v\n", dialect, err)
		os.Exit(1)
	}
	outBuf := bytes.NewBuffer(nil)
	valuesBuf := bytes.NewBuffer(nil)

	sort.Strings(names)

	for _, name := range names {
		if !strings.HasSuffix(name, ".sql") {
			continue
		}
		contents, err := ioutil.ReadFile(fmt.Sprintf("%s/%s/%s", baseDir, dialect, name))
		if err != nil {
			fmt.Printf("error opening file %s with dialect %s: %v", name, dialect, err)
			os.Exit(1)
		}
		migrationsValueTemplate.Execute(valuesBuf, struct {
			Name     string
			Contents string
		}{
			Name:     name,
			Contents: string(contents),
		})
	}
	migrationsTemplate.Execute(outBuf, struct {
		Type   string
		Values string
	}{
		Type:   dialect,
		Values: valuesBuf.String(),
	})

	outFile := fmt.Sprintf("%s/%s.gen.go", baseDir, dialect)
	if err := ioutil.WriteFile(outFile, outBuf.Bytes(), 0644); err != nil {
		fmt.Printf("error writing file %q: %v\n", outFile, err)
		os.Exit(1)
	}
}

var migrationsTemplate = template.Must(template.New("").Parse(
	`// Code generated by "make migrations"; DO NOT EDIT.
package migrations
	
import (
	"bytes"
)

var {{ .Type }}Migrations = map[string]*fakeFile{
	"migrations": {
		name: "migrations",
	},
	{{ .Values }}
}
`))

var migrationsValueTemplate = template.Must(template.New("").Parse(
	`"migrations/{{ .Name }}": {
		name: "{{ .Name }}",
		bytes: []byte(` + "`\n{{ .Contents }}\n`" + `),
	},
`))