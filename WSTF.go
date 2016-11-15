package readabilty

import (
	"bufio"
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/blevesearch/segment"
	"github.com/neurosnap/sentences"
	"github.com/speedata/hyphenation"
)

type SachTextFormel int

const (
	WSTF1 = iota
	WSTF2
	WSTF3
	WSTF4
)

type Readabilty struct {
	tokenizer *sentences.DefaultSentenceTokenizer
	hyphen    *hyphenation.Lang
	lang      string
}

// Implements the Wiener Sachtextformel according to
// https://de.wikipedia.org/wiki/Lesbarkeitsindex#Wiener_Sachtextformel
// Further reading at https://wwwmatthes.in.tum.de/file/1f2t6qd87twtm/Sebis-Public-Website/-/Comparison-of-Law-Texts-An-Analysis-of-German-and-Austrian-Legislation-regarding-Linguistic-and-Structural-Metrics/Wa15a.pdf
func (r *Readabilty) WienerSachTextFormelType(text string, WSTF_Type SachTextFormel) (float32, error) {
	if r.lang != "german" {
		return 0, errors.New("WienerSachTextFormelType operates only on german text")
	}

	var cnt_sentences, cnt_words, cnt_words_3psyllables, cnt_words_1syllable, cnt_wordlen_p6chars, cnt_words_sentence int
	var meansentencelenght float32

	// split input in sentences
	sentences := r.tokenizer.Tokenize(text)
	for _, val := range sentences {
		cnt_words_sentence = 0
		// split input in words
		scanner := bufio.NewScanner(strings.NewReader(val.Text))
		scanner.Split(segment.SplitWords)

		for scanner.Scan() {

			word := scanner.Text()
			wordlen := utf8.RuneCountInString(word)

			// count words in syllables
			hyp := r.hyphen.Hyphenate(word)

			if len(hyp) >= 3 {
				cnt_words_3psyllables++
			}

			if len(hyp) == 1 {
				cnt_words_1syllable++
			}

			if wordlen > 6 {
				cnt_wordlen_p6chars++
			}

			cnt_words_sentence++
		}
		cnt_words += cnt_words_sentence
		cnt_sentences++
		// calcualte a running average
		// cf. http://stackoverflow.com/a/16757630/433253
		meansentencelenght -= meansentencelenght / float32(cnt_sentences)
		meansentencelenght += float32(cnt_words) / float32(cnt_sentences)

	}

	var MS = float32(cnt_words_3psyllables / cnt_words)
	var SL = meansentencelenght

	var wstfretval float32
	switch WSTF_Type {
	case WSTF1:
		var IW = float32(cnt_wordlen_p6chars / cnt_words)
		var ES = float32(cnt_words_1syllable / cnt_words)
		wstfretval = 0.1935*MS + 0.1672*SL + 0.1297*IW - 0.0327*ES - 0.875
	case WSTF2:
	}

	return wstfretval, nil
}

func (r *Readabilty) WienerSachTextFormel(text string) (float32, error) {
	return r.WienerSachTextFormelType(text, WSTF1)
}

type initalisationfilename struct {
	segmentationfilename, hyphenfileame string
}

var initalisationfilenames = map[string]initalisationfilename{
	"de": initalisationfilename{"data/german.json", "data/hyphen/hyph-de-1996.pat.txt"},
}

func NewReadability(lang string) (*Readabilty, error) {
	var r Readabilty
	// create the default sentence tokenizer

	f, err := os.Open(initalisationfilenames[lang].hyphenfileame)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	training, err := sentences.LoadTraining(b)
	if err != nil {
		return nil, err
	}
	r.tokenizer = sentences.NewSentenceTokenizer(training)

	// create the hyphenation
	f, err = os.Open(initalisationfilenames[lang].hyphenfileame)
	if err != nil {
		return nil, err
	}
	l, err := hyphenation.New(f)
	if err != nil {
		return nil, err
	}
	r.hyphen = l

	r.lang = lang
	return &r, nil
}
func init() {

}
