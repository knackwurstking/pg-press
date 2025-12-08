package services

import (
	"fmt"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

const TableNameFeeds = "feeds"

type Feeds struct {
	*Base
	broadcaster Broadcaster
}

func NewFeeds(r *Registry) *Feeds {
	return &Feeds{
		Base: NewBase(r),
	}
}

func (f *Feeds) SetBroadcaster(broadcaster Broadcaster) {
	f.broadcaster = broadcaster
}

func (f *Feeds) List() ([]*models.Feed, *errors.MasterError) {
	rows, err := f.DB.Query(SQLListFeeds)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	return ScanRows(rows, ScanFeed)
}

func (f *Feeds) ListRange(offset, limit int) ([]*models.Feed, *errors.MasterError) {
	rows, err := f.DB.Query(SQLListFeedsWithRange, limit, offset)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	return ScanRows(rows, ScanFeed)
}

func (f *Feeds) ListByUser(userID int64, offset, limit int) ([]*models.Feed, *errors.MasterError) {
	rows, err := f.DB.Query(SQLListFeedsByUserIDWithRange, userID, limit, offset)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	return ScanRows(rows, ScanFeed)
}

func (f *Feeds) Get(id models.FeedID) (*models.Feed, *errors.MasterError) {
	feed, err := ScanFeed(f.DB.QueryRow(SQLGetFeed, id))
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}

	return feed, nil
}

func (f *Feeds) Add(title, content string, userID models.TelegramID) *errors.MasterError {
	feed := models.NewFeed(title, content, userID)
	verr := feed.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	result, err := f.DB.Exec(SQLAddFeed, feed.Title, feed.Content, feed.UserID, feed.CreatedAt)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return errors.NewMasterError(err, 0)
	}
	feed.ID = models.FeedID(id)

	if f.broadcaster != nil {
		f.broadcaster.Broadcast()
	}

	return nil
}

// TODO: Continue here with moving SQL statements to constants
func (f *Feeds) Delete(id models.FeedID) *errors.MasterError {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, TableNameFeeds)

	_, err := f.DB.Exec(query, id)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (f *Feeds) DeleteBefore(timestamp int64) (int, *errors.MasterError) {
	query := fmt.Sprintf(`DELETE FROM %s WHERE created_at < ?`, TableNameFeeds)

	result, err := f.DB.Exec(query, timestamp)
	if err != nil {
		return 0, errors.NewMasterError(err, 0)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, errors.NewMasterError(err, 0)
	}

	return int(rowsAffected), nil
}

func (f *Feeds) Count() (int, *errors.MasterError) {
	count, err := f.QueryCount(fmt.Sprintf(`SELECT COUNT(*) FROM %s`, TableNameFeeds))
	if err != nil {
		return 0, errors.NewMasterError(err, 0)
	}

	return count, nil
}

func (f *Feeds) CountByUser(userID int64) (int, *errors.MasterError) {
	count, err := f.QueryCount(
		fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE user_id = ?`, TableNameFeeds),
		userID,
	)
	if err != nil {
		return 0, errors.NewMasterError(err, 0)
	}

	return count, nil
}
