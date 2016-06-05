package anki

import (
	"encoding/json"
	"errors"
)

type Collection struct {
	ID             ID                    `db:"id"`     // Primary key; should always be 1, as there's only ever one collection per *.apkg file
	Created        TimestampSeconds      `db:"crt"`    // Created timestamp (seconds)
	Modified       TimestampMilliseconds `db:"mod"`    // Last modified timestamp (milliseconds)
	SchemaModified TimestampMilliseconds `db:"scm"`    // Schema modification time (milliseconds)
	Version        int                   `db:"ver"`    // Version?
	Dirty          int                   `db:"dty"`    // Dirty? No longer used. See https://github.com/dae/anki/blob/master/anki/collection.py#L90
	UpdateSequence int                   `db:"usn"`    // update sequence number. used to figure out diffs when syncing
	LastSync       TimestampMilliseconds `db:"ls"`     // Last sync time (milliseconds)
	Config         Config                `db:"conf"`   // JSON blob containing configuration options
	Models         Models                `db:"models"` // JSON array of json objects containing the models (aka Note types)
	Decks          string                `db:"decks"`  // JSON array of json objects containing decks
	DeckConfig     string                `db:"dconf"`  // JSON blob containing deck configuration options
	Tags           string                `db:"tags"`   // a cache of tags used in the collection
}

type Config struct {
	NextPos       int             `json:"nextPos"`
	EstimateTimes bool            `json:"estTimes"`
	ActiveDecks   []ID            `json:"activeDecks"` // Array of active decks(?)
	SortType      string          `json:"sortType"`
	TimeLimit     DurationSeconds `json:"timeLimit"`
	SortBackwards bool            `json:"sortBackwards"`
	AddToCurrent  bool            `json:"addToCur"` // Add new cards to current deck(?)
	CurrentDeck   ID              `json:"curDeck"`
	NewBury       bool            `json:"newBury"`
	NewSpread     int             `json:"newSpread"`
	DueCounts     bool            `json:"dueCounts"`
	CurrentModel  ID              `json:"curModel,string"`
	CollapseTime  int             `json:"collapseTime"`
}

func scanJSON(src interface{}, target interface{}) error {
	var blob []byte
	switch src.(type) {
	case []byte:
		blob = src.([]byte)
	case string:
		blob = []byte(src.(string))
	default:
		return errors.New("Incompatible type for Config")
	}
	return json.Unmarshal(blob, target)
}

func (c *Config) Scan(src interface{}) error {
	return scanJSON(src, c)
}

type Models map[ID]Model

func (m *Models) Scan(src interface{}) error {
	return scanJSON(src, m)
}

func (m *Models) UnmarshalJSON(src []byte) error {
	tmp := make(map[string]Model)
	if err := json.Unmarshal(src, &tmp); err != nil {
		return err
	}
	newM := make(map[ID]Model)
	for _, v := range tmp {
		newM[v.ID] = v
	}
	*m = Models(newM)
	return nil
}

// Model aka Note Type
//
// Excluded from this definition is the `vers` field, which is no longer used by Anki.
type Model struct {
	ID             ID               `json:"id"`    // Model ID
	Name           string           `json:"name"`  // Model name
	Tags           []string         `json:"tags"`  // Anki saves the tags of the last added note to the current model
	DeckID         ID               `json:"did"`   // Deck ID of deck where cards are added by default
	Fields         []Field          `json:"flds"`  // Array of Field objects
	SortField      int              `json:"sortf"` // Integer specifying which field is used for sorting in the browser
	Templates      []Template       `json:"tmpls"`
	Type           ModelType        `json:"type"`      // Model type: Standard or Cloze
	LatexPre       string           `json:"latexPre"`  // preamble for LaTeX expressions
	LatexPost      string           `json:"latexPost"` // String added to end of LaTeX expressions (usually \\end{document})
	CSS            string           `json:"css"`       // CSS, shared for all templates
	Modified       TimestampSeconds `json:"mod"`       // Modification time in seconds
	RequiredFields []CardConstraint `json:"req"`       // Array of card constraints describing which fields are required for each card to be generated
	UpdateSequence int              `json:"usn"`       // Update sequence number: used in same way as other usn vales in db
}

type ModelType int

const (
	ModelTypeStandard ModelType = iota
	ModelTypeCloze
)

// Field
//
// Excluded from this definition is the `media` field, which appears to no longer be used.
type Field struct {
	Name     string `json:"name"`   // Field name
	Sticky   bool   `json:"sticky"` // Sticky fields retain the value that was last added when adding new notes
	RTL      bool   `json:"rtl"`    // Boolean to indicate if this field uses Right-to-Left script
	Ord      int    `json:"ord"`    // Ordinal of the field. Goes from 0 to num fields -1.
	Font     string `json:"font"`   // Display font
	FontSize int    `json:"size"`   // Font size
}

type CardConstraint struct {
	Index     int    // Card index
	MatchType string // "any" or "all"
	Fields    []int  // Array of fields which must exist
}

func (c *CardConstraint) UnmarshalJSON(src []byte) error {
	tmp := make([]interface{}, 3)
	if err := json.Unmarshal(src, &tmp); err != nil {
		return err
	}
	c.Index = int(tmp[0].(float64))
	c.MatchType = tmp[1].(string)
	tmpAry := tmp[2].([]interface{})
	c.Fields = make([]int, len(tmpAry))
	for i, v := range tmpAry {
		c.Fields[i] = int(v.(float64))
	}
	return nil
}

type Template struct {
	Name                  string `json:"name"`  // Template name
	Ordinal               int    `json:"ord"`   // Template number
	QuestionFormat        string `json:"qfmt"`  // Question format
	AnswerFormat          string `json:"afmt"`  // Answer format
	BrowserQuestionFormat string `json:"bqfmt"` // Browser question format
	BrowserAnswerFormat   string `json:"bafmt"` // Browser answer format
	DeckOverride          ID     `json:"did"`   // Deck override (null by default) (??)

}
