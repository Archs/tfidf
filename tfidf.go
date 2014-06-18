package tfidf

import (
	"bufio"
	"fmt"
	"github.com/huichen/sego"
	"os"
	"sort"
	"strings"
)

var (
	stop_words = []string{"the", "of", "is", "and", "to", "in", "that", "we", "for", "an", "are", "by", "be", "as", "on", "with", "can", "if", "from", "which", "you", "it", "this", "then", "at", "have", "all", "not", "one", "has", "or", "that",
		"的", "之", "和", "在", "我", "一", "是", "那"}
)

// Extractor关键字提取器
type Extractor struct {
	*sego.Segmenter
	IDF *Freq
}

// NewExtractor 通过idf文件和字典文件构造关键字提取器
func NewExtractor(idfpath, dictfpaths string) (*Extractor, error) {
	var err error
	ex := new(Extractor)
	ex.IDF, err = ReadIdf(idfpath)
	if err != nil {
		return nil, err
	}
	ex.Segmenter = new(sego.Segmenter)
	ex.LoadDictionary(dictfpaths)
	return ex, nil
}

type Word struct {
	string
	Score float64
}

type Words []Word

func (w Words) Len() int {
	return len(w)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (w Words) Less(i, j int) bool {
	if w[i].Score <= w[j].Score {
		return true
	}
	return false
}

// Swap swaps the elements with indexes i and j.
func (w Words) Swap(i, j int) {
	w[j], w[i] = w[i], w[j]
}

// Freq maintains a idf of the universe
type Freq struct {
	m      map[string]float64
	median float64
}

func newFreq() *Freq {
	f := new(Freq)
	f.m = make(map[string]float64)
	return f
}

func (i *Freq) Freq(word string) float64 {
	if val, ok := i.m[word]; ok {
		return val
	}
	return i.median
}

func (f *Freq) Words() Words {
	w := []Word{}
	for k, v := range f.m {
		w = append(w, Word{k, v})
	}
	return w
}

// NewIdf read fpath to get the idf
func ReadIdf(fpath string) (*Freq, error) {
	r, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	i := newFreq()
	scanner := bufio.NewScanner(r)
	scoreArray := []float64{}
	for scanner.Scan() {
		var w string
		var score float64
		_, err = fmt.Sscanf(scanner.Text(), "%s %f", &w, &score)
		if err != nil {
			return nil, err
		}
		i.m[w] = score
		scoreArray = append(scoreArray, score)
	}
	sort.Float64s(scoreArray)
	i.median = scoreArray[len(scoreArray)/2]
	return i, nil
}

func isStopWord(w string) bool {
	w = strings.ToLower(w)
	for _, s := range stop_words {
		if s == w {
			return true
		}
	}
	return false
}

func termFreq(segs []sego.Segment) *Freq {
	words := sego.SegmentsToSlice(segs, false)
	freq := newFreq()
	total := float64(0)
	for _, w := range words {
		if isStopWord(w) || len(w) < 2 {
			continue
		}
		// 此时freq.Freq(w)默认是0
		freq.m[w] = freq.Freq(w) + 1
		total += 1
	}
	for k, v := range freq.m {
		freq.m[k] = v / total
	}
	return freq
}

func (e *Extractor) Keywords(sentence string, topK ...int) Words {
	segs := e.Segment([]byte(sentence))
	tf := termFreq(segs)
	n := 10
	if len(topK) >= 1 {
		n = topK[0]
	}
	kws := Words([]Word{})
	for word, tfval := range tf.m {
		var w Word
		w.string = word
		w.Score = tfval * e.IDF.Freq(word)
		kws = append(kws, w)
	}
	sort.Reverse(kws)
	if len(kws) <= 10 {
		return kws
	}
	return kws[:n]
}
