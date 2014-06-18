package tfidf

import (
	"bufio"
	"fmt"
	"github.com/huichen/sego"
	"os"
	"sort"
	"strings"
	"unicode/utf8"
)

var (
	stop_words = []string{"the", "of", "is", "and", "to", "in", "that", "we", "for", "an", "are", "by", "be", "as", "on", "with", "can", "if", "from", "which", "you", "it", "this", "then", "at", "have", "all", "not", "one", "has", "or", "that",
		"你们", "要是", "坐在", "没有", "还是", "一样", "不是", "回来", "一句", "一声", "自己", "已经", "这个", "他们", "的话", "一只",
		"那个", "两个", "以后", "地上", "随之", "就是", "咱们", "仍然", "出来", "刚刚", "下来", "屋里", "时候", "说话", "不能", "几乎", "进入", "然后", "觉得", "不要", "那些", "什么", "完全", "走出", "似的", "开始", "这样", "这儿", "三个", "怎么", "整个", "突然", "接着", "听到", "出门", "不敢", "可以", "只是", "不住", "直到", "只有", "之后", "最后", "人们", "坐下", "终于", "十分", "而且", "想到", "无法", "不再", "这回", "发生", "那种", "全都", "更加", "不过", "这种", "而是", "牲畜", "学校", "院里", "不下", "有点", "早已", "重新", "跟前", "今日", "第二天",
		"先生",
	}
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
	Word  string
	Score float64
}

type Words []Word

func (w Words) Len() int {
	return len(w)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (w Words) Less(i, j int) bool {
	return w[i].Score < w[j].Score
}

// Swap swaps the elements with indexes i and j.
func (w Words) Swap(i, j int) {
	w[i], w[j] = w[j], w[i]
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
	// 排除量词
	if utf8.RuneCountInString(w) == 2 && strings.HasPrefix(w, "一") {
		return true
	}
	return false
}

func termFreq(segs []sego.Segment) *Freq {
	words := sego.SegmentsToSlice(segs, true)
	freq := newFreq()
	total := float64(0)
	for _, w := range words {
		if utf8.RuneCountInString(w) < 2 || isStopWord(w) {
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

func (e *Extractor) Keywords(sentence string, topK ...int) []Word {
	segs := e.Segment([]byte(sentence))
	tf := termFreq(segs)
	n := 10
	if len(topK) >= 1 {
		n = topK[0]
	}
	kws := []Word{}
	for word, tfval := range tf.m {
		var w Word
		w.Word = word
		w.Score = tfval * e.IDF.Freq(word)
		// println(word, "\t", tfval, w.Score)
		kws = append(kws, w)
	}
	sort.Sort(sort.Reverse(Words(kws)))
	if len(kws) <= n {
		return kws
	}
	return kws[:n]
}
