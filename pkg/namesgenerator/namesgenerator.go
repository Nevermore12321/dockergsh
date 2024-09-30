package namesgenerator

import (
	"fmt"
	"math/rand"
	"time"
)

/*
namesgenerator 根据著名科学家和黑客的名字，生成容器名称
*/
var (
	left  = [...]string{"happy", "jolly", "dreamy", "sad", "angry", "pensive", "focused", "sleepy", "grave", "distracted", "determined", "stoic", "stupefied", "sharp", "agitated", "cocky", "tender", "goofy", "furious", "desperate", "hopeful", "compassionate", "silly", "lonely", "condescending", "naughty", "kickass", "drunk", "boring", "nostalgic", "ecstatic", "insane", "cranky", "mad", "jovial", "sick", "hungry", "thirsty", "elegant", "backstabbing", "clever", "trusting", "loving", "suspicious", "berserk", "high", "romantic", "prickly", "evil"}
	right = [...]string{"albattani", "almeida", "archimedes", "ardinghelli", "babbage", "bardeen", "bartik", "bell", "blackwell", "bohr", "brattain", "brown", "carson", "colden", "curie", "darwin", "davinci", "einstein", "elion", "engelbart", "euclid", "fermat", "fermi", "feynman", "franklin", "galileo", "goldstine", "goodall", "hawking", "heisenberg", "hoover", "hopper", "hypatia", "jones", "kirch", "kowalevski", "lalande", "leakey", "lovelace", "lumiere", "mayer", "mccarthy", "mcclintock", "mclean", "meitner", "mestorf", "morse", "newton", "nobel", "pare", "pasteur", "perlman", "pike", "poincare", "ptolemy", "ritchie", "rosalind", "sammet", "shockley", "sinoussi", "stallman", "tesla", "thompson", "torvalds", "turing", "wilson", "wozniak", "wright", "yonath"}
)

func GetRandomName(retry int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	name := fmt.Sprintf("%s_%s", left[r.Intn(len(left))], right[r.Intn(len(right))])
	if retry > 0 {
		name = fmt.Sprintf("%s%d", name, rand.Intn(10))
	}
	return name
}
