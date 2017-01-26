// Copyright: Jonathan Hall
// License: GNU AGPL, Version 3 or later; http://www.gnu.org/licenses/agpl.html

package anki

import (
	"encoding/json"
	"errors"
	"sort"
	"strings"
)

// Collection is an Anki Collection, stored in the `col` table.
type Collection struct {
	ID             ID                     `db:"id"`     // Primary key; should always be 1, as there's only ever one collection per *.apkg file
	Created        *TimestampSeconds      `db:"crt"`    // Created timestamp (seconds)
	Modified       *TimestampMilliseconds `db:"mod"`    // Last modified timestamp (milliseconds)
	SchemaModified *TimestampMilliseconds `db:"scm"`    // Schema modification time (milliseconds)
	Version        int                    `db:"ver"`    // Version?
	Dirty          BoolInt                `db:"dty"`    // Dirty? No longer used. See https://github.com/dae/anki/blob/master/anki/collection.py#L90
	UpdateSequence int                    `db:"usn"`    // update sequence number. used to figure out diffs when syncing
	LastSync       *TimestampMilliseconds `db:"ls"`     // Last sync time (milliseconds)
	Config         Config                 `db:"conf"`   // JSON blob containing configuration options
	Models         Models                 `db:"models"` // JSON array of json objects containing the models (aka Note types)
	Decks          Decks                  `db:"decks"`  // JSON array of json objects containing decks
	DeckConfigs    DeckConfigs            `db:"dconf"`  // JSON blob containing deck configuration options
	Tags           string                 `db:"tags"`   // a cache of tags used in the collection
	apkg           *Apkg
}

// Config represents basic global configuration for the Anki client.
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

// Scan implements the sql.Scanner interface for the Config type.
func (c *Config) Scan(src interface{}) error {
	return scanJSON(src, c)
}

// Models is a collection of Models (aka note types), stored as JSON in the `models`
// column of the `col` table.
type Models map[ID]*Model

// Scan implements the sql.Scanner interface for the Models type.
func (m *Models) Scan(src interface{}) error {
	return scanJSON(src, m)
}

// UnmarshalJSON implements the json.Unmarshaler interface for the Models type.
func (m *Models) UnmarshalJSON(src []byte) error {
	tmp := make(map[string]*Model)
	if err := json.Unmarshal(src, &tmp); err != nil {
		return err
	}
	newMap := make(map[ID]*Model)
	for _, v := range tmp {
		newMap[v.ID] = v
	}
	*m = Models(newMap)
	return nil
}

// Model (aka Note Type)
//
// Excluded from this definition is the `vers` field, which is no longer used by Anki.
type Model struct {
	ID             ID                `json:"id"`    // Model ID
	Name           string            `json:"name"`  // Model name
	Tags           []string          `json:"tags"`  // Anki saves the tags of the last added note to the current model
	DeckID         ID                `json:"did"`   // Deck ID of deck where cards are added by default
	Fields         []*Field          `json:"flds"`  // Array of Field objects
	SortField      int               `json:"sortf"` // Integer specifying which field is used for sorting in the browser
	Templates      []*Template       `json:"tmpls"`
	Type           ModelType         `json:"type"`      // Model type: Standard or Cloze
	LatexPre       string            `json:"latexPre"`  // preamble for LaTeX expressions
	LatexPost      string            `json:"latexPost"` // String added to end of LaTeX expressions (usually \\end{document})
	CSS            string            `json:"css"`       // CSS, shared for all templates
	Modified       *TimestampSeconds `json:"mod"`       // Modification time in seconds
	RequiredFields []*CardConstraint `json:"req"`       // Array of card constraints describing which fields are required for each card to be generated
	UpdateSequence int               `json:"usn"`       // Update sequence number: used in same way as other usn vales in db
}

// Enum representing the available Note Type Types (confusing, eh?)
type ModelType int

const (
	// ModelTypeStandard indicates an Anki Basic note type
	ModelTypeStandard ModelType = iota
	// ModelTypeCloze indicates an Anki Cloze note type
	ModelTypeCloze
)

// A field of a model
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

// A card constraint defines which fields are necessary for a particular card
// type to be generated. This is (apparently) auto-calculated whenever a note
// is created or modified.
type CardConstraint struct {
	Index     int    // Card index
	MatchType string // "any" or "all"
	Fields    []int  // Array of fields which must exist
}

// UnmarshalJSON implements the json.Unmarshaler interface for the
// CardConstraint type
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

// A Template definition. A template definition represents a single card type,
// and is stored as part of a Model.
type Template struct {
	Name                  string `json:"name"`  // Template name
	Ordinal               int    `json:"ord"`   // Template number
	QuestionFormat        string `json:"qfmt"`  // Question format
	AnswerFormat          string `json:"afmt"`  // Answer format
	BrowserQuestionFormat string `json:"bqfmt"` // Browser question format
	BrowserAnswerFormat   string `json:"bafmt"` // Browser answer format
	DeckOverride          ID     `json:"did"`   // Deck override (null by default) (??)
}

// A collection of Decks
type Decks map[ID]*Deck

// Scan implements the sql.Scanner interface for the Decks type.
func (d *Decks) Scan(src interface{}) error {
	return scanJSON(src, d)
}

// UnmarshalJSON implements the json.Unmarshaler interface for the Decks type.
func (d *Decks) UnmarshalJSON(src []byte) error {
	tmp := make(map[string]*Deck)
	if err := json.Unmarshal(src, &tmp); err != nil {
		return err
	}
	newMap := make(map[ID]*Deck)
	for _, v := range tmp {
		newMap[v.ID] = v
	}
	*d = Decks(newMap)
	return nil
}

// A Deck definition
type Deck struct {
	ID                      ID                `json:"id"`               // Deck ID
	Name                    string            `json:"name"`             // Deck name
	Description             string            `json:"desc"`             // Deck description
	Modified                *TimestampSeconds `json:"mod"`              // Last modification time in seconds
	UpdateSequence          int               `json:"usn"`              // Update sequence number. Used in the same way as the other USN values
	Collapsed               bool              `json:"collapsed"`        // True when the deck is collapsed
	BrowserCollapsed        bool              `json:"browserCollapsed"` // True when the deck is collapsed in the browser
	ExtendedNewCardLimit    int               `json:"extendedNew"`      // Extended new card limit for custom study
	ExtendedReviewCardLimit int               `json:"extendedRev"`      // Extended review card limit for custom study
	Dynamic                 BoolInt           `json:"dyn"`              // True for a dynamic (aka filtered) deck
	ConfigID                ID                `json:"conf"`             // ID of option group from dconf in `col` table
	NewToday                [2]int            `json:"newToday"`         // two number array used somehow for custom study
	ReviewsToday            [2]int            `json:"revToday"`         // two number array used somehow for custom study
	LearnToday              [2]int            `json:"lrnToday"`         // two number array used somehow for custom study
	TimeToday               [2]int            `json:"timeToday"`        // two number array used somehow for custom study (in ms)
	Config                  *DeckConfig       `json:"-"`
}

// Collection of per-deck configurations
type DeckConfigs map[ID]*DeckConfig

// Scan implements the sql.Scanner interface for the DeckConfigs type.
func (dc *DeckConfigs) Scan(src interface{}) error {
	return scanJSON(src, dc)
}

// UnmarshalJSON implements the json.Unmarshaler interface for the DeckConfigs
// type.
func (dc *DeckConfigs) UnmarshalJSON(src []byte) error {
	tmp := make(map[string]*DeckConfig)
	if err := json.Unmarshal(src, &tmp); err != nil {
		return err
	}
	newMap := make(map[ID]*DeckConfig)
	for _, v := range tmp {
		newMap[v.ID] = v
	}
	*dc = DeckConfigs(newMap)
	return nil
}

// Per-Deck configuration options.
//
// Excluded from this definition is the `minSpace` field from Reviews, as it is no longer used.
type DeckConfig struct {
	ID               ID                `json:"id"`       // Deck ID
	Name             string            `json:"name"`     // Deck Name
	ReplayAudio      bool              `json:"replayq"`  // When answer shown, replay both question and answer audio
	ShowTimer        BoolInt           `json:"timer"`    // Show answer timer
	MaxAnswerSeconds int               `json:"maxTaken"` // Ignore answers that take longer than this many seconds
	Modified         *TimestampSeconds `json:"mod"`      // Modified timestamp
	AutoPlay         bool              `json:"autoplay"` // Automatically play audio
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

// Enum of available leech actions
type LeechAction int

const (
	LeechActionSuspendCard LeechAction = iota
	LeechActoinTagOnly
)

// Enum of new card order options
type NewCardOrder int

const (
	NewCardOrderOrderAdded NewCardOrder = iota
	NewCardOrderRandomOrder
)

// Note definition
//
// Excludes the `flags` and `data` columns, which are no longer used
type Note struct {
	ID             ID                `db:"id"`   // Primary key
	GUID           string            `db:"guid"` // globally unique id, almost certainly used for syncing
	ModelID        ID                `db:"mid"`  // Model ID
	Modified       *TimestampSeconds `db:"mod"`  // Last modified time
	UpdateSequence int               `db:"usn"`  // Update sequence number (no longer used?)
	Tags           string            `db:"tags"` // List of the note's tags
	FieldValues    FieldValues       `db:"flds"` // Values for the note's fields
	UniqueField    string            `db:"sfld"` // The text of the first field, used for Anki's simplistic uniqueness checking
	Checksum       int64             `db:"csum"` // Field checksum used for duplicate check. Integer representation of first 8 digits of sha1 hash of the first field
}

// The Tags type represents an array of tags for a note.
type Tags []string

// Scan implements the sql.Scanner interface for the Tags type.
func (t *Tags) Scan(src interface{}) error {
	var tmp string
	switch src.(type) {
	case []byte:
		tmp = string(src.([]byte))
	case string:
		tmp = src.(string)
	default:
		return errors.New("Incompatible type for Tags")
	}
	tags := Tags(strings.Split(tmp, " "))
	sort.Strings(tags)
	*t = tags
	return nil
}

type FieldValues []string

// Scan implements the sql.Scanner interface for the FieldValues type.
func (fv *FieldValues) Scan(src interface{}) error {
	var tmp string
	switch src.(type) {
	case []byte:
		tmp = string(src.([]byte))
	case string:
		tmp = src.(string)
	default:
		return errors.New("Incompatible type for FieldValues")
	}
	*fv = FieldValues(strings.Split(tmp, "\x1f"))
	return nil
}

// Card definition
//
// This definition excludes the `flags` and `data` fields, which are no longer
// used. Additionally, this definition modifies the original senses of `due`,
// `odue`, and `ivl` by converting them to a consistent representation.
// `Specifically
//
// `due` and `odue` are stored in one of three states:
//  - For card type 0 (new), the due time is ignored. Here we convert it to 0.
//  - For card type 1 (learning), the due time is stored as seconds since epoch.
//    We leave this as-is.
//  - For card type 2 (due), the due time is stored as days since the collection
//    was created. We convert this to seconds since epoch.
//
// `ivl` is stored either as negative seconds, or as positive days. We convert
// both to positive seconds.
type Card struct {
	ID             ID                `db:"id"`     // Primary key
	NoteID         ID                `db:"nid"`    // Foreign Key to a Note
	DeckID         ID                `db:"did"`    // Foreign key to a Deck
	TemplateID     int               `db:"ord"`    // The Template ID, within the Model, to which this card corresponds.
	Modified       *TimestampSeconds `db:"mod"`    // Last modified time
	UpdateSequence int               `db:"usn"`    // Update sequence number
	Type           CardType          `db:"type"`   // Card type: new, learning, due
	Queue          CardQueue         `db:"queue"`  // Queue: suspended, user buried, sched buried
	Due            *TimestampSeconds `db:"due"`    // Time when the card is next due
	Interval       *DurationSeconds  `db:"ivl"`    // SRS interval in seconds
	Factor         float32           `db:"factor"` // SRS factor
	ReviewCount    int               `db:"reps"`   // Number of reviews
	Lapses         int               `db:"lapses"` // Number of times card went from "answered correctly" to "answered incorrectly" state
	Left           int               `db:"left"`   // Reviews remaining until graduation
	OriginalDue    *TimestampSeconds `db:"odue"`   // Original due time. Only used when card is in filtered deck.
	OriginalDeckID ID                `db:"odid"`   // Original Deck ID. Only used when card is in filtered deck.
}

type CardType int

const (
	CardTypeNew CardType = iota
	CardTypeLearning
	CardTypeReview
)

type CardQueue int

// CardQueue specifies the card's queue type
//
// See https://github.com/dae/anki/blob/master/anki/sched.py#L17
// and https://github.com/dae/anki/blob/master/anki/cards.py#L14
const (
	CardQueueSchedBuried CardQueue = -3 // Sched Buried (??, possibly unused)
	CardQueueBuried      CardQueue = -2 // Buried
	CardQueueSuspended   CardQueue = -1 // Suspended
	CardQueueNew         CardQueue = 0  // New/Cram
	CardQueueLearning    CardQueue = 1  // Learning
	CardQueueReview      CardQueue = 2  // Review
	CardQueueRelearning  CardQueue = 3  // Day Learn (Relearn?)
)

// Review definition
//
// `ivl` is stored either as negative seconds, or as positive days. We convert
// both to positive seconds.
type Review struct {
	Timestamp      *TimestampSeconds    `db:"id"`      // Times when the review was done
	CardID         ID                   `db:"cid"`     // Foreign key to a Card
	UpdateSequence int                  `db:"usn"`     // Update sequence number
	Ease           ReviewEase           `db:"ease"`    // Button pushed to score recall: wrong, hard, ok, easy
	Interval       DurationSeconds      `db:"ivl"`     // SRS interval in seconds
	LastInterval   DurationSeconds      `db:"lastIvl"` // Prevoius SRS interval in seconds
	Factor         float32              `db:"factor"`  // SRS factor
	ReviewTime     DurationMilliseconds `db:"time"`    // Time spent on the review
	Type           ReviewType           `db:"type"`    // Review type: learn, review, relearn, cram
}

type ReviewEase int

const (
	ReviewEaseWrong ReviewEase = 1
	ReviewEaseHard  ReviewEase = 2
	ReviewEaseOK    ReviewEase = 3
	ReviewEaseEasy  ReviewEase = 4
)

type ReviewType int

const (
	ReviewTypeLearn ReviewType = iota
	ReviewTypeReview
	ReviewTypeRelearn
	ReviewTypeCram
)
