package main

import (
	"flag"
	"github.com/Archs/tfidf"
	"io/ioutil"
	"log"
	"os"
)

var (
	fpath string
	n     int
)

var (
	ex  *tfidf.Extractor
	err error
)

func main() {
	data, err := ioutil.ReadFile(fpath)
	if err != nil {
		log.Fatal(err)
	}
	kws := ex.Keywords(string(data), n)
	for i, v := range kws {
		log.Println(i, v.Word, v.Score)
	}
}

func init() {
	flag.StringVar(&fpath, "f", "", "file to extract keywords")
	flag.IntVar(&n, "n", 30, "# of keyworks")
	flag.Parse()
	if fpath == "" {
		flag.PrintDefaults()
		os.Exit(-1)
	}
	ex, err = tfidf.NewExtractor("../idf.txt", "../dictionary.txt")
	if err != nil {
		log.Fatal(err)
	}
}
