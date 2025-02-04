package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var (
	path   string
	revert *bool
)

func main() {
	revert = flag.Bool("r", false, "revert the last numbering")
	fhelp := flag.Bool("h", false, "help on command")
	revertMap := make(map[string]string)

	flag.Parse()

	if *fhelp {
		flag.Usage()
		os.Exit(1)
	}

	path = flag.Args()[0]

	log.Printf(`using path : "%s"`, path)

	recoverFile := filepath.Join(path, ".recover.json")

	if _, err := os.Stat(recoverFile); !errors.Is(err, os.ErrNotExist) && !*revert {
		log.Fatalf("path already processed, only reverting possible.")
		os.Exit(-1)
	}

	fs, err := os.ReadDir(path)
	if err != nil {
		log.Fatalf("error reading dir: %v", err)
	}

	if *revert {
		log.Println("doing a revert!")
		js, err := os.ReadFile(recoverFile)
		if err != nil {
			log.Fatalf("error reading reverting file: %v", err)
			os.Exit(-1)
		}
		err = json.Unmarshal(js, &revertMap)
		if err != nil {
			log.Fatalf("error unmarshalling reverting file: %v", err)
			os.Exit(-1)
		}
		for _, f := range fs {
			if strings.HasPrefix(f.Name(), ".") {
				log.Printf("%s ignored", f.Name())
				continue
			}
			src := filepath.Join(path, f.Name())
			dstName := revertMap[f.Name()]
			dst := filepath.Join(path, dstName)
			err = os.Rename(src, dst)
			if err != nil {
				log.Printf("error in renaming: %v\r\n", err)
			}
			log.Printf("%s -> %s", src, dst)
		}
		os.Remove(recoverFile)
		os.Exit(1)
	}
	log.Printf("found %d files\r\n", len(fs))
	sort.Slice(fs, func(a, b int) bool {
		ai, err := fs[a].Info()
		if err != nil {
			log.Fatalf("error in sorting: %v", err)
		}
		bi, err := fs[b].Info()
		if err != nil {
			log.Fatalf("error in sorting: %v", err)
		}
		return ai.ModTime().Before(bi.ModTime())
	})

	idx := 1
	for _, f := range fs {
		if strings.HasPrefix(f.Name(), ".") {
			log.Printf("%s ignored", f.Name())
			continue
		}
		src := filepath.Join(path, f.Name())
		ext := filepath.Ext(f.Name())
		dstName := fmt.Sprintf("%.04d%s", idx, ext)
		dst := filepath.Join(path, dstName)
		err = os.Rename(src, dst)
		if err != nil {
			log.Printf("error in renaming: %v\r\n", err)
		}
		revertMap[dstName] = f.Name()
		log.Printf("%s -> %s", src, dst)
		idx++
	}
	js, err := json.Marshal(revertMap)
	if err != nil {
		log.Fatalf("error in converting to js: %v", err)
	}
	os.WriteFile(recoverFile, js, os.ModePerm)
}
