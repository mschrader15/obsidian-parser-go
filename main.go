package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"github.com/gomarkdown/markdown/parser"
	"gopkg.in/yaml.v2"
)

type T struct {
	P string `yaml:"publish,omitempty"`
	SavePath string `yaml:"relativeSavePath,omitempty"`
}

func main() {
	vaultPathPtr := flag.String("vaultPath", "", "The path to Obsidian Vault's Root Directory")
	outputRootPtr := flag.String("outputRoot", "", "The path to output root")
	removeBracketsPtr := flag.Bool("removeBrackets", false, "Whether or not to remove the brackets from internal links")

	flag.Parse()

	// Create the Parser (not used currently)
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	parser := parser.NewWithExtensions(extensions)

	iterate(*vaultPathPtr, *parser, *removeBracketsPtr, *outputRootPtr)
}

func iterate(rootPath string, parser parser.Parser, removeBrackets bool, outputRoot string) {
	
	filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatalf(err.Error())
		}
		// fmt.Printf("File Name: %s\n", info.Name())
		fmt.Printf("File Path: %s\n", path)
		if filepath.Ext(path) == ".md" {
			fileBytes := openFile(path)
			t := findHeadMatter(fileBytes)
			if t.P == "true" && removeBrackets {
				strippedBytes := stripBrackets(fileBytes, parser)
				
				// copy and replace the file path
				strippedBytes = cpImages(strippedBytes, rootPath, outputRoot)

				file, _ := os.OpenFile(
					filepath.Join(outputRoot, t.SavePath, info.Name()),
					os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
					0666,
				 )
				defer file.Close()
				file.Write(strippedBytes)

			}
		}
		return nil
	})
}

func openFile(path string) []byte {
	fileBytes, _ := ioutil.ReadFile(path)
	return fileBytes
}

func findHeadMatter(fileByte []byte,) T {
	var t T
	re := regexp.MustCompile(`---\n((.*\n)*?)---\n`)
	matches := re.FindAllSubmatch(fileByte, -1)
	if len(matches) > 0 {
		err := yaml.Unmarshal(matches[0][1], &t)
		if err != nil {
			log.Println("cannot unmarshal data: %v", err)
		}
		t.P = strings.ToLower(t.P)
	}
	return t
}

func stripBrackets(fileBytes []byte, p parser.Parser) []byte {
	re := regexp.MustCompile(`[^!]\[\[(.*)\]\]`)
	newBytes := re.ReplaceAll(fileBytes, []byte(`$1`))
	return newBytes
}


func cpImages(fileBytes []byte, inputRootDir string, outputRootDir string) []byte {
	re := regexp.MustCompile(`(!\[\[)(.*)(\]\])`)
	matches := re.FindAllSubmatch(fileBytes, -1)
	var replacements [][][]byte
	for _, match := range(matches){
		println(string(match[2]))
		err := filepath.Walk(inputRootDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Fatalf(err.Error())
			}
			if info.Name() == string(match[2]){
				// source, _ := os.Open(path)
				// defer source.Close()
				// TODO create a way to specify the folder location
				// destination, _ := os.Create(filepath.Join(outputRootDir, "resources", info.Name()))
				// defer destination.Close()
				
				// io.Copy(source, destination)
				// // source.Close()
				newFile := strings.ReplaceAll(info.Name(), " ", "_")
				bytesRead, err := ioutil.ReadFile(path)

				if err != nil {
					log.Fatal(err)
				}
			
				err = ioutil.WriteFile(filepath.Join(outputRootDir, "resources", newFile), bytesRead, 0644)

				// replace the url in the file with a more verbose one.
				var matcher = [][]byte{
					[]byte(`!\[\[` + string(match[2]) + `]]`), 
					[]byte("![" + info.Name() + "](" + "/resources/" + newFile + ")"),
				}

				replacements = append(replacements, matcher)

				// replacements[i][0] = match[0][1] + []byte("resources/" + info.Name())

				return io.EOF
			}
			return nil			
		})

		if err == io.EOF {
			err = nil
		}
	}

	for _, group := range(replacements){
		re := regexp.MustCompile(string(group[0]))
		fileBytes = re.ReplaceAll(fileBytes, group[1])
	}

	return fileBytes
}