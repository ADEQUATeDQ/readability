package readability

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/blevesearch/segment"
	"github.com/neurosnap/sentences"
	"github.com/speedata/hyphenation"
)

type CompareType int

const (
	_ = iota
	WSTF1
	WSTF2
	WSTF3
	WSTF4
)

type Readability struct {
	tokenizer *sentences.DefaultSentenceTokenizer
	hyphen    *hyphenation.Lang
	lang      string
}

// Implements the Wiener Sachtextformel according to
// https://de.wikipedia.org/wiki/Lesbarkeitsindex#Wiener_Sachtextformel
// Further reading at https://wwwmatthes.in.tum.de/file/1f2t6qd87twtm/Sebis-Public-Website/-/Comparison-of-Law-Texts-An-Analysis-of-German-and-Austrian-Legislation-regarding-Linguistic-and-Structural-Metrics/Wa15a.pdf
func (r *Readability) WienerSachTextFormelType(text string, WSTF_Type CompareType) (float32, error) {

	if !(WSTF_Type == WSTF1 || WSTF_Type == WSTF2 || WSTF_Type == WSTF3 || WSTF_Type == WSTF4) {
		return 0, errors.New(fmt.Sprintf("Unknown compare type provided to WienerSachTextFormelType: %d", WSTF_Type))
	}

	if r.lang != "de" {
		return 0, errors.New("WienerSachTextFormelType operates only on german text")
	}

	var cnt_sentences, cnt_words, cnt_words_3psyllables, cnt_words_1syllable, cnt_wordlen_p6chars int
	// split input in sentences
	sentences := r.tokenizer.Tokenize(text)
	for _, val := range sentences {

		// split sentences into words
		segmenter := segment.NewWordSegmenter(strings.NewReader(val.Text))

		for segmenter.Segment() {

			if segmenter.Type() != segment.Letter {
				continue
			}
			word := segmenter.Text()
			wordlen := utf8.RuneCountInString(word)

			// count syllables in words
			hyp := r.hyphen.Hyphenate(word)

			if len(hyp) >= 3 {
				cnt_words_3psyllables++
			} else if len(hyp) == 1 {
				cnt_words_1syllable++
			}

			if wordlen > 6 {
				cnt_wordlen_p6chars++
			}
			cnt_words++

		}

		cnt_sentences++

	}

	var MS = float32(cnt_words_3psyllables) / float32(cnt_words) * 100
	var SL = float32(cnt_words) / float32(cnt_sentences)
	var wstfretval float32

	switch WSTF_Type {
	case WSTF1:
		var IW = 0.1297 * float32(cnt_wordlen_p6chars) / float32(cnt_words) * 100
		var ES = 0.0327 * float32(cnt_words_1syllable) / float32(cnt_words) * 100
		wstfretval = 0.1935*MS + 0.1672*SL + IW - ES - 0.875
	case WSTF2:
		var IW = 0.1373 * float32(cnt_wordlen_p6chars) / float32(cnt_words) * 100
		wstfretval = 0.2007*MS + 0.1682*SL + IW - 2.779
	case WSTF3:
		wstfretval = 0.2963*MS + 0.1905*SL - 1.1144
	case WSTF4:
		wstfretval = 0.2656*SL + 0.2744*MS - 1.693
	}

	return wstfretval, nil
}

// Returns the readability of a text according to the Wiener Sachtextformel.
// The type is fixed to Type1
// cf. https://de.wikipedia.org/wiki/Lesbarkeitsindex#Wiener_Sachtextformel
func (r *Readability) WienerSachTextFormel(text string) (float32, error) {
	return r.WienerSachTextFormelType(text, WSTF1)
}

type initalisationfilename struct {
	segmentationfilename, hyphenfileame string
}

var initalisationfilenames = map[string]initalisationfilename{
	"de": initalisationfilename{"data/german.json", "data/hyphen/hyph-de-1996.pat.txt"},
}

// Initializes the Readability Engine by reading language-specific hypenation patterns and sentence training data.
// Returns a Readability-Object or error, if initialisation failes
func NewReadability(lang string) (*Readability, error) {
	var r Readability
	// create the default sentence tokenizer

	f, err := os.Open(initalisationfilenames[lang].segmentationfilename)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	training, err := sentences.LoadTraining(b)
	if err != nil {
		return nil, errors.New("NewReadabily.LoadTraining failed: " + err.Error())
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
