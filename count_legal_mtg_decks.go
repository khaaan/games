// https://twitter.com/jordancurve/status/920773458031792129
// Calculate the number of legal decks (60 main + 15 sideboard) in various Magic the Gathering formats.
// Results (as of 2017-07-28 mtgjson data):
// Standard: 1.89e+152 (189345355916230985373072200536169947947295794716089748222356848897535008894697835577214506567305987637419359573332299963631250585641452157281030004526776)
// Modern: 2.47e+209 (246511459455625348732139446965857761235921626159567784697436049301569552137240823762776689418734538645168096497645751307281466847703823941861869181432462231383059014175889995268515798746671138992528629285395774)
// Legacy: 9.71e+222 (9711422830638704141259812921405089710335676405917350613072183705149034130461377032700261245864254868125958229056507427492174381630697827240946680622430281166393400388562121015248005082352931408613565265695867679151960622256)
// Vintage: 1.21e+223 (12063272679040923314177308539650193007692645139519724704069486829832479473447140142201268049499776311015674384983815936045615943597154054628579727750462487506874852277403770068593267760078598877771190130846684438528997450774)
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"strings"
)

type Legality struct {
	Format, Legality string
}

type Card struct {
	Name       string
	Type       string
	Legalities []Legality
}

func main() {
	limits := FormatLimits("AllCards-x.json") // from https://mtgjson.com/json/AllCards-x.json.zip
	formats := []string{"Standard", "Modern", "Legacy", "Vintage"}
	for _, f := range formats {
		c := CountDecks(60, 15, limits[f])
		fmt.Printf("%8s: %.3g (%v)\n", f, new(big.Float).SetInt(c), c)
	}
}

func FormatLimits(mtgJsonFile string) map[string][]int {
	mtgJson, err := ioutil.ReadFile(mtgJsonFile)
	if err != nil {
		panic(err)
	}
	var cards map[string]Card
	if err := json.Unmarshal(mtgJson, &cards); err != nil {
		panic(err)
	}
	limits := map[string][]int{}
	for _, c := range cards {
		for _, leg := range c.Legalities {
			f := leg.Format
			if _, ok := limits[f]; !ok {
				limits[f] = []int{}
			}
			lim := 0
			if leg.Legality == "Legal" {
				if strings.HasPrefix(c.Type, "Basic Land") {
					lim = 1000
				} else {
					lim = 4
				}
			} else if leg.Legality == "Restricted" {
				lim = 1
			}
			if lim > 0 {
				limits[f] = append(limits[f], lim)
			}
		}
	}
	return limits
}

// Cache key, used to speed up LimitedMultiChooose.
type key struct {
	main, side, numCards int
}

// CountDecks(M, S, L) returns the number of ways to make a deck with M
// cards in the main deck and S cards in the sideboard where there are len(L)
// cards to choose from, and there can be at most L[I] copies of card I in your
// mainboard and sideboard combined (0 < I < len(L)).
// Examples (mainboard/sideboard):
//   CountDecks(3, 0, []int{1,2,3})=6 (abb abc acc bbc bcc ccd)
//   CountDecks(3, 3, []int{1,2,3})=6 (abb/ccc abc/bcc acc/bbc bbc/acc bcc/abc ccc/abb)
//   CountDecks(3, 1, []int{1,2,3})=12 (abb/c abc/b abc/c acc/b acc/c bbc/a bbc/c bcc/a bcc/b bcc/c ccc/a ccc/b)
//   CountDecks(4, 0, []int{1,2,3})=5 (abbc abcc accc bbcc bccc)
//   CountDecks(4, 1, []int{1,2,3})=8 (abbc/c abcc/b abcc/b accc/b bbcc/a bbcc/c bccc/a bccc/b)
//   CountDecks(4, 2, []int{1,2,3})=5 (abbc/cc abcc/bc accc/bb bbcc/ac bccc/ab)
//   CountDecks(60, 15, []int{75})=1 (the "all islands" example)
func CountDecks(numMain, numSide int, limit []int) *big.Int {
	return _countDecks(numMain, numSide, limit, map[key]*big.Int{})
}

func _countDecks(numMain, numSide int, limit []int, cache map[key]*big.Int) *big.Int {
	if numMain+numSide == 0 {
		return big.NewInt(1)
	}
	if len(limit) == 0 {
		return big.NewInt(0)
	}
	key := key{numMain, numSide, len(limit)}
	if val, ok := cache[key]; ok {
		return val
	}
	sum := big.NewInt(0)
	for m := 0; m <= numMain && m <= limit[0]; m++ {
		for s := 0; s <= numSide && m+s <= limit[0]; s++ {
			sum.Add(sum, _countDecks(numMain-m, numSide-s, limit[1:], cache))
		}
	}
	cache[key] = sum
	return sum
}
