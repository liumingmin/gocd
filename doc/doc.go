package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func GetCurrPath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	return filepath.Dir(path)
}

func main() {
	var sb strings.Builder

	prefix, _ := ioutil.ReadFile(filepath.Join(GetCurrPath(), "README-BASE.md"))
	sb.WriteString(string(prefix))

	genDocByTestFile(GetCurrPath(), 1, &sb)
	ioutil.WriteFile("README.md", []byte(sb.String()), 0666)

	if err := exec.Command("cmd", "/c", "markdown-toc --maxdepth 2 -i README.md").Run(); err != nil {
		fmt.Println(err)
	}
}

var moduleCnName = map[string]string{
	"cd_server_test.go": "usage",
}

//dir := filepath.Dir(filePath)
func genDocByTestFile(dir string, level int, sb *strings.Builder) {
	files, _ := ioutil.ReadDir(dir)

	nextLevel := level + 1

	for _, file := range files {
		if file.IsDir() {
			if strings.HasPrefix(file.Name(), ".") {
				continue
			}
			genDirLevel(file.Name(), nextLevel, sb)
			genDocByTestFile(filepath.Join(dir, file.Name()), nextLevel, sb)
			continue
		}

		if strings.HasSuffix(file.Name(), "_test.go") {
			codeFilePath := dir + "/" + file.Name()
			bs, err := ioutil.ReadFile(codeFilePath)
			if err != nil {
				continue
			}
			content := string(bs)

			genDirLevel(file.Name(), nextLevel, sb)
			parseTestCode(nextLevel, content, sb)
		}
	}
}

func genDirLevel(dirName string, level int, sb *strings.Builder) {
	prefixSymbol := ""
	for i := 0; i < level; i++ {
		prefixSymbol += "#"
	}

	dirCnName, ok := moduleCnName[dirName]
	if ok {
		dirCnName = dirName + " " + dirCnName
	} else {
		dirCnName = dirName
	}
	fmt.Println(dirName)
	sb.WriteString(fmt.Sprintf("%s %s\n", prefixSymbol, dirCnName))
}

func parseTestCode(level int, content string, sb *strings.Builder) {
	reg, _ := regexp.Compile(`(?U)func (?P<fname>.*)\(t \*testing\.T\) *\{(?P<body>(.|\n)*)\n\}`)
	match := reg.FindAllStringSubmatch(content, -1)

	for _, item := range match {
		genDirLevel(item[1], level+1, sb)

		sb.WriteString("```go\n")
		sb.WriteString(item[2] + "\n")
		sb.WriteString("```\n")
	}
}
