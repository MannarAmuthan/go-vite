package main

import (
	"github.com/cross-cpm/go-shutil"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func processTemplate(tmplPath string) *template.Template {
	tmplFile := filepath.Join("templates", tmplPath)
	tmpl, err := template.ParseFiles(tmplFile)
	if err != nil {
		printErr(err)
	}

	return tmpl
}

func handleMd(mdPath string) {
	content, err := ioutil.ReadFile(mdPath)
	if err != nil {
		printErr(err)
	}

	restContent, fm := parseFrontmatter(content)
	bodyHtml := mdRender(restContent)
	relPath, _ := filepath.Rel("pages/", mdPath)

	var buildPath string
	if strings.HasSuffix(relPath, "_index.md") {
		dir, _ := filepath.Split(relPath)
		buildPath = filepath.Join("build", dir)
	} else {
		buildPath = filepath.Join(
			"build",
			strings.TrimSuffix(relPath, filepath.Ext(relPath)),
		)
	}

	os.MkdirAll(buildPath, 0755)

	cfg := parseConfig()
	fm.Body = string(bodyHtml)

	// combine config and matter structs
	combined := struct {
		Cfg Config
		Fm Matter
	}{cfg, fm}

	htmlFile, err := os.Create(filepath.Join(buildPath, "index.html"))
	if err != nil {
		printErr(err)
		return
	}
	if fm.Template == "" {
		fm.Template = "text.html"
	}
	tmpl := processTemplate(fm.Template)
	err = tmpl.Execute(htmlFile, combined)
	if err != nil {
		printErr(err)
		return
	}
	htmlFile.Close()
}

func viteBuild() {
	err := filepath.Walk("./pages", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			printErr(err)
			return err
		}
		if filepath.Ext(path) == ".md" {
			handleMd(path)
		} else {
			f, err := os.Stat(path)
			if err != nil {
				printErr(err)
			}
			mode := f.Mode()
			if mode.IsRegular() {
				options := shutil.CopyOptions{}
				relPath, _ := filepath.Rel("pages/", path)
				options.FollowSymlinks = true
				shutil.CopyFile(
					path,
					filepath.Join("build", relPath),
					&options,
				)
			}
		}
		return nil
	})
	if err != nil {
		printErr(err)
	}

	_, err = shutil.CopyTree("static", filepath.Join("build", "static"), nil)
	if err != nil {
		printErr(err)
	}
}
