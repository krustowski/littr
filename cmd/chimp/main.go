package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

func main() {
	findCmd := exec.Command("find", ".", "-wholename", "./pkg/frontend/*/*.go", "-print")

	var out strings.Builder
	findCmd.Stdout = &out

	if err := findCmd.Run(); err != nil {
		fmt.Println(err.Error())
		return
	}

	// Show the stdout.
	//fmt.Printf("%s", out.String())

	var imports = make(map[string]string)

	var importStarted bool

	// Range over split output by newline to operate on each file.
	for _, file := range strings.Split(out.String(), "\n") {
		f, err := os.Open(file)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		rdr := bufio.NewReader(f)

		for {
			line, err := rdr.ReadSlice(byte('\n'))
			if err != nil {
				fmt.Println(err.Error())
			}
		
			if len(line) == 0 {
				continue
			}

			if strings.Contains(string(line), "//") {
				continue
			}

			if strings.Contains(string(line), "import (") {
				importStarted = true
				continue
			}

			if strings.Contains(string(line), ")") {
				importStarted = false
				break
			}

			if importStarted {
				imp := strings.Trim(strings.TrimSpace(string(line)), "\"")
				imports[imp] = ""
			}
		}
	}

	//
	// Results
	//

	fmt.Printf("imports: %d\n", len(imports))
	
	var keys []string
	
	for key, _ := range imports {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	oneport := strings.Join(keys, "\n")

	fmt.Printf("%s", oneport)
}
