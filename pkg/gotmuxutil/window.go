package gotmuxutil

import (
	"strconv"

	gotmux "github.com/jubnzv/go-tmux"
	"github.com/rs/zerolog/log"
)

// ListPanes finds a window by id
func ListPanes(window gotmux.Window) ([]gotmux.Pane, error) {
	log.Warn().Str("test", strconv.Itoa(window.Id)).Msg("yo")

	return gotmux.ListPanes([]string{"-t", strconv.Itoa(window.Id)})
}
