package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/HFunction2013/coj-cli/internal/api"
	"github.com/HFunction2013/coj-cli/internal/iostreams"
	"github.com/HFunction2013/coj-cli/internal/prompter"
)

// Action is the outcome of a "What's next?" confirmation, mirroring
// gh's pkg/cmd/pr/shared Action enum.
type Action int

const (
	SubmitAction Action = iota
	PreviewAction
	CancelAction
)

const (
	submitLabel  = "Submit"
	previewLabel = "Continue in browser"
	cancelLabel  = "Cancel"
)

var errCancelled = fmt.Errorf("cancelled")

// confirmSubmission asks "What's next?" after an interactive run.
//
// gh uses this so that a command filled in by prompts never fires blindly:
// the user gets Submit / Continue in browser / Cancel. We offer Preview only
// when a web page exists for this endpoint.
func confirmSubmission(io *iostreams.IOStreams, p prompter.Prompter, allowPreview bool) (Action, error) {
	options := []string{submitLabel}
	if allowPreview {
		options = append(options, previewLabel)
	}
	options = append(options, cancelLabel)

	result, err := p.Select("What's next?", submitLabel, options)
	if err != nil {
		return -1, fmt.Errorf("could not prompt: %w", err)
	}
	if result < 0 || result >= len(options) {
		return -1, fmt.Errorf("invalid index: %d", result)
	}
	switch options[result] {
	case submitLabel:
		return SubmitAction, nil
	case previewLabel:
		return PreviewAction, nil
	default:
		return CancelAction, nil
	}
}

// destructivePattern matches verbs that destroy or overwrite data.
var destructivePattern = regexp.MustCompile(`(?i)^(del|delete|remove|hide|reset|one-key-reset-pwd|clear)`)

// isWriteAction reports whether an endpoint mutates state.
func isWriteAction(ep *api.Endpoint) bool {
	if ep == nil {
		return false
	}
	action := ep.Action
	if destructivePattern.MatchString(action) {
		return true
	}
	// get*/list*/can*/search* are reads; anything else is treated as a write.
	return !regexp.MustCompile(`(?i)^(get|list|can|search|check|show|export|student-get)`).MatchString(action)
}

// isDestructive reports whether an endpoint destroys or overwrites data.
func isDestructive(ep *api.Endpoint) bool {
	return ep != nil && destructivePattern.MatchString(ep.Action)
}

// confirmDestructive applies gh's ConfirmDeletion to destructive verbs:
// the user must type the target value, so a stray Enter cannot wipe a class.
//
// It only triggers when the user actually went through prompts, or when the
// ids came from flags but look like a batch (comma-separated) — scripted
// single-id calls keep working untouched.
func confirmDestructive(io *iostreams.IOStreams, p prompter.Prompter, ep *api.Endpoint, params *api.Params) error {
	if !isDestructive(ep) || !io.CanPrompt() {
		return nil
	}
	target := ""
	for _, key := range []string{"F_ID", "F_MatchID", "F_ClassID", "F_SubjectID", "F_UserID", "F_HomeworkID", "F_CourseID"} {
		if v, ok := params.Fields[key]; ok && v != "" {
			target = v
			break
		}
	}
	if target == "" {
		return nil
	}
	// Single, explicitly-provided id: the user already named the target.
	if !strings.Contains(target, ",") {
		return nil
	}
	fmt.Fprintf(io.ErrOut, "About to run %s on %d ids.\n", ep.Path, len(strings.Split(target, ",")))
	return p.ConfirmDeletion(target)
}

// resolvedByPrompting reports whether this run filled any field by asking.
//
// gh gates its follow-up "What's next?" on interactivity so that scripted
// runs stay deterministic; we additionally require that we really asked
// something, otherwise every flagged write would prompt.
func resolvedByPrompting(ep *api.Endpoint, params *api.Params) bool {
	if ep == nil || params == nil {
		return false
	}
	return len(params.PromptedFields()) > 0
}
