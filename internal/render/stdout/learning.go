package stdout

// import (
// 	"fmt"
// 	tea "github.com/charmbracelet/bubbletea"
// 	"os"
// 	"strconv"
// )
//
// type model struct {
// 	count    int
// 	quitting bool
// }
//
// func initialModel() model { return model{} }
//
// func (m model) Init() tea.Cmd { return nil }
//
// func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
// 	switch msg := msg.(type) {
// 	case tea.KeyMsg:
// 		switch msg.String() {
// 		case "q", "ctrl+c":
// 			m.quitting = true
// 			return m, tea.Quit
// 		case "k", "up":
// 			m.count++
// 		case "j", "down":
// 			m.count--
// 		case "r":
// 			m.count = 0
// 		}
// 	}
// 	return m, nil
// }
//
// func (m model) View() string {
// 	if m.quitting {
// 		return "Bye!\n"
// 	}
// 	return "Counter:" + strconv.Itoa(m.count) + "\n"
// }
//
// func Start() {
// 	p := tea.NewProgram(initialModel())
// 	if _, err := p.Run(); err != nil {
// 		fmt.Println("error:", err)
// 		os.Exit(1)
// 	}
// }
