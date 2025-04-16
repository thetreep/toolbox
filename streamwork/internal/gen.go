package main

import (
	_ "embed"
	"os"
	"os/exec"
	"path"
	"runtime"
	"text/template"
)

//go:embed stream_gen.go.tpl
var streamGenTpl string

func main() {
	tmpl := template.Must(
		template.New("stream").Funcs(
			template.FuncMap{
				"iterate": func(start int, end int) []int {
					values := make([]int, 0, 1+end-start)
					for i := start; i <= end; i++ {
						values = append(values, i)
					}
					return values
				},
				"add": func(a int, b int) int {
					return a + b
				},
				"sub": func(a int, b int) int {
					return a - b
				},
				"mod": func(a int, b int) int {
					return a % b
				},
			},
		).Parse(streamGenTpl),
	)
	_, file, _, _ := runtime.Caller(0)

	targetFile := path.Join(path.Dir(file), "..", "stream_gen.go")
	f, err := os.Create(targetFile)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(f, nil)
	if err != nil {
		panic(err)
	}
	err = f.Close()
	if err != nil {
		panic(err)
	}
	cmd := exec.Command("gofmt", "-w", targetFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
}
