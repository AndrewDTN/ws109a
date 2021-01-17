package main

// mentioned in bleve google group
// https://groups.google.com/forum/#!topic/bleve/-5Q6W3oBizY

import (
	"encoding/json"
	"fmt"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/document"
	"log"
	"os"
	"strconv"
	"time"
)

type Person struct {
	Name string
	Age  int
}

type Occupation struct {
	Institution string
	JobTitle    string
	StartDate   time.Time
	Supervisor  string
}

type SuperPowers struct {
	StrengthLevel int
	SpeedLevel    int
	CanFly        bool
}

type SuperHero struct {
	ID int
	Person
	HeroName string
	DayJob   Occupation
	Powers   SuperPowers
}

func main() {
	// setup
	indexName := "heros.bleve"
	index := makeBleveIndex(indexName)
	id, document, pretty := getData()
	indexData(id, document, index)

	// search for some text
	query := bleve.NewMatchQuery("Superman")
	search := bleve.NewSearchRequest(query)
	searchResults, err := index.Search(search)
	if err != nil {
		log.Fatalln("Trouble with search request!")
	}

	// retrieve document
	BleveSearchDocs := getBleveDocsFromSearchResults(searchResults, index)
	if len(BleveSearchDocs) != 1 {
		log.Fatalln("Trouble retrieving docs from search!")
	}

	OriginalDocs := getOriginalDocsFromSearchResults(searchResults, index)
	if len(OriginalDocs) != 1 {
		log.Fatalln("Trouble retrieving original docs from search!")
	}

	// show before and after
	fmt.Println("Original input document is:\n")
	fmt.Printf("%s\n\n", pretty)

	fmt.Println("Re-created document with the index and search results is:\n")
	fmt.Printf("%s\n\n", BleveSearchDocs[0])

	fmt.Println("Original document retrived from Index.GetInternal is:\n")
	fmt.Printf("%s\n", OriginalDocs[0])

	// clean up
	if err := os.RemoveAll(indexName); err != nil {
		log.Fatalln("Trouble removing index file:", indexName)
	}
}

func getOriginalDocsFromSearchResults(
	results *bleve.SearchResult,
	index bleve.Index,
) [][]byte {
	docs := make([][]byte, 0)

	for _, val := range results.Hits {
		id := val.ID
		raw, err := index.GetInternal([]byte(id))
		if err != nil {
			log.Fatal("Trouble getting internal doc:", err)
		}
		docs = append(docs, raw)
	}
	return docs
}

func getBleveDocsFromSearchResults(
	results *bleve.SearchResult,
	index bleve.Index,
) [][]byte {
	docs := make([][]byte, 0)

	for _, val := range results.Hits {
		id := val.ID
		doc, _ := index.Document(id)

		rv := struct {
			ID     string                 `json:"id"`
			Fields map[string]interface{} `json:"fields"`
		}{
			ID:     id,
			Fields: map[string]interface{}{},
		}
		for _, field := range doc.Fields {
			var newval interface{}
			switch field := field.(type) {
			case *document.TextField:
				newval = string(field.Value())
			case *document.NumericField:
				n, err := field.Number()
				if err == nil {
					newval = n
				}
			case *document.DateTimeField:
				d, err := field.DateTime()
				if err == nil {
					newval = d.Format(time.RFC3339Nano)
				}
			}
			existing, existed := rv.Fields[field.Name()]
			if existed {
				switch existing := existing.(type) {
				case []interface{}:
					rv.Fields[field.Name()] = append(existing, newval)
				case interface{}:
					arr := make([]interface{}, 2)
					arr[0] = existing
					arr[1] = newval
					rv.Fields[field.Name()] = arr
				}
			} else {
				rv.Fields[field.Name()] = newval
			}
		}
		j2, _ := json.MarshalIndent(rv, "", "    ")
		docs = append(docs, j2)
	}

	return docs
}

func makeBleveIndex(indexName string) bleve.Index {
	mapping := bleve.NewIndexMapping()
	index, err := bleve.New(indexName, mapping)
	if err != nil {
		log.Fatalln("Trouble making index!")
	}
	return index
}

func getData() (id int, document []byte, pretty []byte) {
	hero := SuperHero{
		ID:       1,
		HeroName: "Superman",
		Person:   Person{Name: "Clark Kent", Age: 30},
		DayJob: Occupation{
			Institution: "Daily Planet",
			JobTitle:    "news reporter",
			StartDate:   time.Date(1938, time.April, 18, 12, 0, 0, 0, time.UTC),
			Supervisor:  "Perry White",
		},
		Powers: SuperPowers{
			StrengthLevel: 10,
			SpeedLevel:    10,
			CanFly:        true,
		},
	}

	document, err := json.Marshal(hero)
	if err != nil {
		log.Fatalln("Trouble json encoding hero (as document)!")
	}

	pretty, err = json.MarshalIndent(hero, "", "    ")
	if err != nil {
		log.Fatalln("Trouble json encoding hero (as pretty JSON)!")
	}

	return hero.ID, document, pretty
}

func indexData(id int, doc []byte, index bleve.Index) {
	err := index.Index(strconv.Itoa(id), doc)
	if err != nil {
		log.Fatal("Trouble indexing data!")
	}

	err = index.SetInternal([]byte(strconv.Itoa(id)), doc)
	if err != nil {
		log.Fatal("Trouble doing SetInternal!")
	}
}