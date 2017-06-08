package main

// This program packs JavaScript files into a Go source file.

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
)

const tpl = `package lib

var Scripts = map[string]string {
	{{ range $name, $body := . }}"{{$name}}": {{ $body }},
	{{ end }}
}
`

func main() {
	// The goal is to produce a map like map[filename]body.
	outFile := os.Args[1]
	filePatterns := os.Args[2:]

	files := expandPaths(filePatterns)
	out, err := os.Create(outFile)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	scripts := map[string]string{}
	for _, f := range files {
		data, err := ioutil.ReadFile(f)
		if err != nil {
			panic(err)
		}
		scripts[f] = "`" + string(data) + "`"
	}

	tt := template.Must(template.New("js2lib").Parse(tpl))
	tt.Execute(out, scripts)
}

func expandPaths(paths []string) []string {
	res := []string{}
	for _, p := range paths {
		matches, err := filepath.Glob(p)
		if err != nil {
			// We're not really fault tolerant
			panic(err)
		}
		res = append(res, matches...)
	}
	return res
}
