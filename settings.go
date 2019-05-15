package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"math"
	"time"

	"github.com/donyori/goctpf"
)

type AiSettings struct {
	AiPiece        Piece         `json:"ai_piece,omitempty"`
	MctsTimeLimit  time.Duration `json:"mcts_time_limit,omitempty"`
	ValidDistThold uint8         `json:"valid_dist_thold,omitempty"`
	UctCmpThold    float64       `json:"uct_cmp_thold,omitempty"`
	UctParamC      float64       `json:"uct_param_c,omitempty"`
}

type BoardPrintSettings struct {
	EmptyChar          string `json:"empty_char,omitempty"`
	BlackChar          string `json:"black_char,omitempty"`
	WhiteChar          string `json:"white_char,omitempty"`
	DoesShowLineNumber bool   `json:"does_show_line_number,omitempty"`
}

type IoSettings struct {
	BoardPrint *BoardPrintSettings `json:"board_print,omitempty"`
}

type Settings struct {
	Rule   Rule                   `json:"rule,omitempty"`
	Ai     *AiSettings            `json:"ai,omitempty"`
	Worker *goctpf.WorkerSettings `json:"worker,omitempty"`
	Io     *IoSettings            `json:"io,omitempty"`
}

func NewSettings() *Settings {
	return &Settings{
		Rule: StandardGomoku,
		Ai: &AiSettings{
			AiPiece:        White,
			MctsTimeLimit:  time.Second * 15,
			ValidDistThold: 1,
			UctCmpThold:    1e-4,
			UctParamC:      math.Sqrt2,
		},
		Worker: goctpf.NewWorkerSettings(),
		Io: &IoSettings{
			BoardPrint: &BoardPrintSettings{
				EmptyChar:          ".",
				BlackChar:          "x",
				WhiteChar:          "o",
				DoesShowLineNumber: true,
			},
		},
	}
}

func LoadSettings() (*Settings, error) {
	data, err := ioutil.ReadFile(SettingsPath)
	if err != nil {
		return nil, err
	}
	settings := NewSettings()
	err = json.Unmarshal(data, settings)
	if err != nil {
		return nil, err
	}
	return settings, nil
}

func StoreSettings(settings *Settings) error {
	if settings == nil {
		panic(errors.New("settings is nil"))
	}
	data, err := json.MarshalIndent(settings, "", "    ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(SettingsPath, data, 0666)
}
