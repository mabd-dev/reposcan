package themeswitcher

import "github.com/mabd-dev/reposcan/internal/theme"

type Model struct {
	theme theme.Theme

	filterFocused bool
	filterQuery   string
}
