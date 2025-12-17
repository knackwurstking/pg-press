package helper

import (
	"fmt"

	"github.com/knackwurstking/pg-press/internal/common"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
)

func ListNotesForLinked(db *common.DB, linked string, id shared.EntityID) ([]*shared.Note, *errors.MasterError) {
	notes, merr := db.Note.Note.List()
	if merr != nil {
		return nil, merr
	}

	n := 0
	matchingLinked := fmt.Sprintf("%s_%d", linked, id)
	for _, note := range notes {
		if note.Linked != matchingLinked {
			continue
		}

		notes[n] = note
		n++
	}
	return notes[:n], nil
}
