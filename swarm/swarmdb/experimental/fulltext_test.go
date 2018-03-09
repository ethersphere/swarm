package swarmdb_test

import (
	"fmt"
	"github.com/ethereum/go-ethereum/swarmdb"
	"testing"
)

func PrintDocs(docs [][]byte) {
}

func getSwarmDB(t *testing.T) (a swarmdb.SwarmDB) {
	swarmdb := swarmdb.NewSwarmDB()
	return *swarmdb
}

func TestPutString(t *testing.T) {
	fmt.Printf("---- TestPutString: generate 20 strings and enumerate them\n")

	hashid := make([]byte, 32)
	r := swarmdb.NewFullTextIndex(getSwarmDB(t), hashid)

	r.StartBuffer()
	k := []byte("game of thrones")
	v := []byte("gameofthronesse08.mp4")
	r.Put(k, v)

	k = []byte("star wars the last jedi")
	v = []byte("starwarsthelastjedi.mp4")
	r.Put(k, v)

	k = []byte("star wars the last jedi")
	v = []byte("starwarsthelastjedi.mp4")
	r.Put(k, v)

	k = []byte("star wars return of the jedi")
	v = []byte("starwarsreturnofthejedi.mp4")
	r.Put(k, v)

	k = []byte("star trek nemesis")
	v = []byte("nemesis2002.mp4")
	r.Put(k, v)

	k = []byte("imitation game")
	v = []byte("imitation_game.mp4")
	r.Put(k, v)

	k = []byte("hunger games")
	v = []byte("hunger_games.mp4")
	r.Put(k, v)

	k = []byte("war games")
	v = []byte("war_games.mp4")
	r.Put(k, v)

	r.FlushBuffer()
	hashid, _ = r.GetRootHash()

	s := swarmdb.NewFullTextIndex(getSwarmDB(t), hashid)

	var words1 []string
	words1 = append(words1, "game")
	docs1, err := s.GetDocs(words1)
	if err != nil {
	} else {
		PrintDocs(docs1)
	}
	var words2 []string
	words2 = append(words2, "star")
	docs2, err := s.GetDocs(words2)
	if err != nil {
	} else {
		PrintDocs(docs2)
	}

	var words3 []string
	words3 = append(words3, "games")
	docs3, err := s.GetDocs(words3)
	if err != nil {
	} else {
		PrintDocs(docs3)
	}

	var words4 []string
	words4 = append(words4, "star")
	words4 = append(words4, "games")
	docs4, err := s.GetDocs(words4)
	if err != nil {
	} else {
		PrintDocs(docs4)
	}

	var words5 []string
	words5 = append(words5, "war")
	words5 = append(words5, "games")
	docs5, err := s.GetDocs(words5)
	if err != nil {
	} else {
		PrintDocs(docs5)
	}

	var words6 []string
	words6 = append(words6, "jedi")
	words6 = append(words6, "star")
	docs6, err := s.GetDocs(words6)
	if err != nil {
	} else {
		PrintDocs(docs6)
	}

}
