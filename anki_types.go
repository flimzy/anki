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
	Decks          Decks                 `db:"decks"`  // JSON array of json objects containing decks
	DeckConfig     DeckConfigs           `db:"dconf"`  // JSON blob containing deck configuration options
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
	newMap := make(map[ID]Model)
	for _, v := range tmp {
		newMap[v.ID] = v
	}
	*m = Models(newMap)
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
	RTL      bool   `json:"rtl"`    // boolean to indicate if this field uses Right-to-Left script
	Ordinal  int    `json:"ord"`    // Ordinal of the field. Goes from 0 to num fields -1.
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

type Decks map[ID]Deck

func (d *Decks) Scan(src interface{}) error {
	return scanJSON(src, d)
}

func (d *Decks) UnmarshalJSON(src []byte) error {
	tmp := make(map[string]Deck)
	if err := json.Unmarshal(src, &tmp); err != nil {
		return err
	}
	newMap := make(map[ID]Deck)
	for _, v := range tmp {
		newMap[v.ID] = v
	}
	*d = Decks(newMap)
	return nil
}

type Deck struct {
	ID                      ID               `json:"id"`               // Deck ID
	Name                    string           `json:"name"`             // Deck name
	Description             string           `json:"desc"`             // Deck description
	Modified                TimestampSeconds `json:"mod"`              // Last modification time in seconds
	UpdateSequence          int              `json:"usn"`              // Update sequence number. Used in the same way as the other USN values
	Collapsed               bool             `json:"collapsed"`        // True when the deck is collapsed
	BrowserCollapsed        bool             `json:"browserCollapsed"` // True when the deck is collapsed in the browser
	ExtendedNewCardLimit    int              `json:"extendedNew"`      // Extended new card limit for custom study
	ExtendedReviewCardLimit int              `json:"extendedRev"`      // Extended review card limit for custom study
	Dynamic                 BoolInt          `json:"dyn"`              // True for a dynamic (aka filtered) deck
	ConfID                  int              `json:"conf"`             // ID of option group from dconf in `col` table
	NewToday                [2]int           `json:"newToday"`         // two number array used somehow for custom study
	ReviewsToday            [2]int           `json:"revToday"`         // two number array used somehow for custom study
	LearnToday              [2]int           `json:"lrnToday"`         // two number array used somehow for custom study
	TimeToday               [2]int           `json:"timeToday"`        // two number array used somehow for custom study (in ms)
}

type DeckConfigs map[ID]DeckConfig

func (dc *DeckConfigs) Scan(src interface{}) error {
	return scanJSON(src, dc)
}

func (dc *DeckConfigs) UnmarshalJSON(src []byte) error {
	tmp := make(map[string]DeckConfig)
	if err := json.Unmarshal(src, &tmp); err != nil {
		return err
	}
	newMap := make(map[ID]DeckConfig)
	for _, v := range tmp {
		newMap[v.ID] = v
	}
	*dc = DeckConfigs(newMap)
	return nil
}

// DeckConfig
//
// Excluded from this definition is the `minSpace` field from Reviews, as it is no longer used.
type DeckConfig struct {
	ID               ID               `json:"id"`       // Deck ID
	Name             string           `json:"name"`     // Deck Name
	ReplayAudio      bool             `json:"replayq"`  // When answer shown, replay both question and answer audio
	ShowTimer        BoolInt          `json:"timer"`    // Show answer timer
	MaxAnswerSeconds int              `json:"maxTaken"` // Ignore answers that take longer than this many seconds
	Modified         TimestampSeconds `json:"mod"`      // Modified timestamp
	AutoPlay         bool             `json:"autoplay"` // Automatically play audio
	Lapses           struct {
		LeechFails      int               `json:"leechFails"`  // Leech threshold
		MinimumInterval DurationDays      `json:"minInt"`      // Minimum interval in days
		LeechAction     LeechAction       `json:"leechAction"` // Leech action: Suspend or Tag Only
		Delays          []DurationMinutes `json:"delays"`      // Steps in minutes
		NewInterval     float32           `json:"mult"`        // New Interval Multiplier
	} `json:"lapse"`
	Reviews struct {
		PerDay           int          `json:"perDay"` // Maximum reviews per day
		Fuzz             float32      `json:"fuzz"`   // Apparently not used?
		IntervalModifier float32      `json:"ivlFct"` // Interval modifier (fraction)
		MaxInterval      DurationDays `json:"maxIvl"` // Maximum interval in days
		EasyBonus        float32      `json:ease4"`   // Easy bonus
		Bury             bool         `json:"bury"`   // Bury related reviews until next day
	} `json:"rev"`
	New struct {
		PerDay        int               `json:"perDay"`        // Maximum new cards per day
		Delays        []DurationMinutes `json:"delays"`        // Steps in minutes
		Bury          bool              `json:"bury"`          // Bury related cards until the next day
		Separate      bool              `json:"separate"`      // Unused??
		Intervals     [3]DurationDays   `json:"ints"`          // Intervals??
		InitialFactor float32           `json:"initialFactor"` // Starting Ease
		Order         NewCardOrder      `json:"order"`         // New card order: Random, or order added
	} `json:"new"`
}

type LeechAction int

const (
	LeechActionSuspendCard LeechAction = iota
	LeechActoinTagOnly
)

type NewCardOrder int

const (
	NewCardOrderOrderAdded NewCardOrder = iota
	NewCardOrderRandomOrder
)
