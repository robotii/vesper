package vesper

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"
)

const (
	red   = "\033[0;31m"
	green = "\033[0;32m"
	blue  = "\033[0;34m"
	black = "\033[0;0m"
)

type replHandler struct {
	rl   *readline.Instance
	vm   *VM
	line string
	cmds []string
}

func (repl *replHandler) Eval(expr string) (string, bool, error) {
	interrupted = false
	repl.cmds = append(repl.cmds, expr)
	whole := strings.Trim(strings.Join(repl.cmds, " "), " ")

	more, err := repl.checkBalanced("(", ")")
	if more || err != nil {
		return "", more, err
	}
	more, err = repl.checkBalanced("[", "]")
	if more || err != nil {
		return "", more, err
	}
	more, err = repl.checkBalanced("{", "}")
	if more || err != nil {
		return "", more, err
	}

	if whole == "" {
		return "", false, nil
	}

	lexpr, err := repl.vm.Read(String(whole), AnyType)
	repl.cmds = nil
	if err != nil {
		return "", false, err
	}

	val, err := repl.vm.Eval(lexpr)
	if err != nil {
		return "", false, err
	}
	if val == nil {
		return red + " Internal error -> Eval returned nil" + black, false, nil
	}
	return "-> " + Write(val), false, nil
}

func (repl *replHandler) checkBalanced(open string, close string) (bool, error) {
	whole := strings.Trim(strings.Join(repl.cmds, " "), " ")
	opens := len(strings.Split(whole, open))
	closes := len(strings.Split(whole, close))
	if opens > closes {
		return true, nil
	} else if closes > opens {
		repl.cmds = nil
		return false, fmt.Errorf("unbalanced '%s' encountered", close)
	}
	return false, nil
}

func (repl *replHandler) Prompt(more bool) string {
	promptType := "*prompt*"
	if more {
		promptType = "*prompt-cont*"
	}
	prompt := GetGlobal(repl.vm.Intern(promptType))
	if prompt != nil {
		return prompt.String()
	}
	if more {
		return ":| "
	}
	return ":> "
}

// REPL starts a REPL
func REPL() {
	interrupts = make(chan os.Signal, 1)
	signal.Notify(interrupts, os.Interrupt)
	defer signal.Stop(interrupts)
	repl := replHandler{}

	var err error

	repl.vm = NewVM()
	repl.rl, err = readline.NewEx(&readline.Config{
		Prompt:              repl.Prompt(false),
		HistoryFile:         filepath.Join(os.TempDir(), "readline_vesper.tmp"),
		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})
	defer func() { _ = repl.rl.Close() }()

	for {
		err = repl.read()
		if err != nil {
			switch err {
			case io.EOF:
				fmt.Println(repl.line)
				return
			case readline.ErrInterrupt: // Pressing Ctrl-C
				if len(repl.line) == 0 {
					if repl.cmds == nil {
						return
					}
				}
				// Erasing command buffer
				repl.cmds = nil
				continue
			}
		}

		fmt.Println(repl.Prompt(len(repl.cmds) > 1) + repl.line)
		fmt.Printf(blue)
		result, more, err := repl.Eval(repl.line)
		fmt.Printf(black)
		if err != nil {
			fmt.Println(red, "***", err, black)
			repl.cmds = nil
			repl.rl.SetPrompt(repl.Prompt(false))
			continue
		} else if more {
			repl.rl.SetPrompt(repl.Prompt(true))
			continue
		} else {
			fmt.Println(green + result + black)
			repl.cmds = nil
			repl.rl.SetPrompt(repl.Prompt(false))
			continue
		}
	}
}

// filterInput just ignores Ctrl-z.
func filterInput(r rune) (rune, bool) {
	return r, r != readline.CharCtrlZ
}

// read fetches one line from input, with the help of Readline library.
func (repl *replHandler) read() error {
	repl.rl.Config.UniqueEditLine = true // required to update the prompt
	line, err := repl.rl.Readline()
	repl.rl.Config.UniqueEditLine = false
	line = strings.TrimSpace(line)
	repl.line = line
	return err
}
