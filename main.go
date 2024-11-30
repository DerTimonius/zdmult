package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
)

var (
	selectedSessions []string
	confirmed        bool
)

func main() {
	sessions, err := getSessions()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	accessible, _ := strconv.ParseBool(os.Getenv("ACCESSIBLE"))
	var options []huh.Option[string]
	for _, session := range sessions {
		options = append(options, huh.NewOption(session, session))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().Options(options...).Title("What sessions do you want to delete?").Value(&selectedSessions),
		),

		huh.NewGroup(huh.NewConfirm().Title("Are you sure you want to delete the selected sessions?").Affirmative("Yes").Negative("No").Value(&confirmed)),
	).WithAccessible(accessible)

	err = form.Run()
	if err != nil {
		fmt.Println("Uh oh:", err)
		os.Exit(1)
	}

	if confirmed {
		runDeletion(accessible)
	}
}

func runDeletion(accessible bool) {
	sessions, err := deleteSessions(selectedSessions)
	if err != nil {
		newForm := huh.NewForm(
			huh.NewGroup(huh.NewConfirm().Title("It appears that normal deletion didn't work. Do you want to force delete the sessions?").Affirmative("Yes").Negative("No").Value(&confirmed)),
		).WithAccessible(accessible)
		err = newForm.Run()
		if err != nil {
			fmt.Println("Uh oh:", err)
			os.Exit(1)
		}

		if confirmed {
			err = forceDeleteSessions(sessions)
			if err != nil {
				fmt.Println("Uh oh:", err)
				os.Exit(1)
			}
		}
	}
}

func getSessions() ([]string, error) {
	cmd := exec.Command("zellij", "ls", "-s", "-r")

	if errors.Is(cmd.Err, exec.ErrDot) {
		cmd.Err = nil
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return []string{}, err
	}
	output := out.String()
	result := strings.Split(output, "\n")
	var sessions []string

	for _, item := range result[1:] {
		if item == "" {
			continue
		}
		sessions = append(sessions, item)
	}

	if len(sessions) == 0 {
		return []string{}, fmt.Errorf("there are no sessions I could delete here")
	}

	return sessions, nil
}

func deleteSessions(sessions []string) ([]string, error) {
	for idx, session := range sessions {
		cmd := exec.Command("zellij", "d", session)

		if errors.Is(cmd.Err, exec.ErrDot) {
			cmd.Err = nil
		}

		err := cmd.Run()
		if err != nil {
			return sessions[idx:], err
		}
	}

	return []string{}, nil
}

func forceDeleteSessions(sessions []string) error {
	for _, session := range sessions {
		cmd := exec.Command("zellij", "d", "-f", session)

		if errors.Is(cmd.Err, exec.ErrDot) {
			cmd.Err = nil
		}

		err := cmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}
