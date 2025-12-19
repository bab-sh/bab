package tui

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/bab-sh/bab/internal/babfile"
	"github.com/charmbracelet/huh"
	"golang.org/x/term"
)

var ErrPromptCancelled = errors.New("prompt cancelled")

var ErrNoTTY = errors.New("no TTY available for interactive prompt")

const (
	boolTrue  = "true"
	boolFalse = "false"
)

func RunPrompt(p babfile.PromptRun, message string) (string, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return handleNonInteractive(p)
	}

	switch p.Type {
	case babfile.PromptTypeConfirm:
		return runConfirm(p, message)
	case babfile.PromptTypeInput:
		return runInput(p, message)
	case babfile.PromptTypeSelect:
		return runSelect(p, message)
	case babfile.PromptTypeMultiselect:
		return runMultiselect(p, message)
	case babfile.PromptTypePassword:
		return runPassword(p, message)
	case babfile.PromptTypeNumber:
		return runNumber(p, message)
	default:
		return "", fmt.Errorf("unknown prompt type: %s", p.Type)
	}
}

func handleNonInteractive(p babfile.PromptRun) (string, error) {
	switch p.Type {
	case babfile.PromptTypeConfirm:
		if p.Default != "" {
			return normalizeConfirmDefault(p.Default), nil
		}
		return "", fmt.Errorf("%w: prompt %q requires 'default' for non-interactive mode", ErrNoTTY, p.Prompt)
	case babfile.PromptTypeInput, babfile.PromptTypeSelect, babfile.PromptTypeNumber:
		if p.Default != "" {
			return p.Default, nil
		}
		return "", fmt.Errorf("%w: prompt %q requires 'default' for non-interactive mode", ErrNoTTY, p.Prompt)
	case babfile.PromptTypeMultiselect:
		if len(p.Defaults) > 0 {
			return strings.Join(p.Defaults, ","), nil
		}
		return "", fmt.Errorf("%w: prompt %q requires 'defaults' for non-interactive mode", ErrNoTTY, p.Prompt)
	case babfile.PromptTypePassword:
		return "", fmt.Errorf("%w: password prompts cannot run in non-interactive mode", ErrNoTTY)
	default:
		return "", fmt.Errorf("%w: prompt %q requires interactive input", ErrNoTTY, p.Prompt)
	}
}

func normalizeConfirmDefault(s string) string {
	lower := strings.ToLower(s)
	switch lower {
	case boolTrue, "yes", "y", "1":
		return boolTrue
	default:
		return boolFalse
	}
}

func runConfirm(p babfile.PromptRun, message string) (string, error) {
	var result bool

	if p.Default != "" {
		result = normalizeConfirmDefault(p.Default) == boolTrue
	}

	confirm := huh.NewConfirm().
		Title(message).
		Value(&result)

	form := huh.NewForm(huh.NewGroup(confirm))
	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", ErrPromptCancelled
		}
		return "", err
	}

	return strconv.FormatBool(result), nil
}

func runInput(p babfile.PromptRun, message string) (string, error) {
	var result string
	if p.Default != "" {
		result = p.Default
	}

	input := huh.NewInput().
		Title(message).
		Value(&result)

	if p.Placeholder != "" {
		input.Placeholder(p.Placeholder)
	}

	if p.Validate != "" {
		re := regexp.MustCompile(p.Validate)
		input.Validate(func(s string) error {
			if !re.MatchString(s) {
				return fmt.Errorf("must match pattern: %s", p.Validate)
			}
			return nil
		})
	}

	form := huh.NewForm(huh.NewGroup(input))
	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", ErrPromptCancelled
		}
		return "", err
	}

	return result, nil
}

func runSelect(p babfile.PromptRun, message string) (string, error) {
	var result string
	if p.Default != "" {
		result = p.Default
	}

	options := make([]huh.Option[string], len(p.Options))
	for i, opt := range p.Options {
		options[i] = huh.NewOption(opt, opt)
	}

	sel := huh.NewSelect[string]().
		Title(message).
		Options(options...).
		Value(&result)

	form := huh.NewForm(huh.NewGroup(sel))
	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", ErrPromptCancelled
		}
		return "", err
	}

	return result, nil
}

func runMultiselect(p babfile.PromptRun, message string) (string, error) {
	var result []string
	if len(p.Defaults) > 0 {
		result = make([]string, len(p.Defaults))
		copy(result, p.Defaults)
	}

	options := make([]huh.Option[string], len(p.Options))
	for i, opt := range p.Options {
		options[i] = huh.NewOption(opt, opt).Selected(slices.Contains(p.Defaults, opt))
	}

	multi := huh.NewMultiSelect[string]().
		Title(message).
		Options(options...).
		Value(&result)

	if p.Min != nil || p.Max != nil {
		multi.Validate(func(selected []string) error {
			if p.Min != nil && len(selected) < *p.Min {
				return fmt.Errorf("select at least %d option(s)", *p.Min)
			}
			if p.Max != nil && len(selected) > *p.Max {
				return fmt.Errorf("select at most %d option(s)", *p.Max)
			}
			return nil
		})
	}

	form := huh.NewForm(huh.NewGroup(multi))
	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", ErrPromptCancelled
		}
		return "", err
	}

	return strings.Join(result, ","), nil
}

func runPassword(p babfile.PromptRun, message string) (string, error) {
	var result string

	pw := huh.NewInput().
		Title(message).
		EchoMode(huh.EchoModePassword).
		Value(&result)

	form := huh.NewForm(huh.NewGroup(pw))
	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", ErrPromptCancelled
		}
		return "", err
	}

	if p.Confirm != nil && *p.Confirm {
		var confirmResult string
		confirmPw := huh.NewInput().
			Title("Confirm " + strings.ToLower(message)).
			EchoMode(huh.EchoModePassword).
			Value(&confirmResult).
			Validate(func(s string) error {
				if s != result {
					return errors.New("passwords do not match")
				}
				return nil
			})

		confirmForm := huh.NewForm(huh.NewGroup(confirmPw))
		if err := confirmForm.Run(); err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				return "", ErrPromptCancelled
			}
			return "", err
		}
	}

	return result, nil
}

func runNumber(p babfile.PromptRun, message string) (string, error) {
	var result string
	if p.Default != "" {
		result = p.Default
	}

	input := huh.NewInput().
		Title(message).
		Value(&result).
		Validate(func(s string) error {
			if s == "" {
				return nil
			}
			n, err := strconv.Atoi(s)
			if err != nil {
				return errors.New("must be a number")
			}
			if p.Min != nil && n < *p.Min {
				return fmt.Errorf("must be at least %d", *p.Min)
			}
			if p.Max != nil && n > *p.Max {
				return fmt.Errorf("must be at most %d", *p.Max)
			}
			return nil
		})

	if p.Placeholder != "" {
		input.Placeholder(p.Placeholder)
	}

	form := huh.NewForm(huh.NewGroup(input))
	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", ErrPromptCancelled
		}
		return "", err
	}

	return result, nil
}
