package storage

import (
	"context"
	"database/sql"
	"time"

	"fmt" // Added for debugging

	"github.com/livekit/livekit-server/pkg/streaming"
	"github.com/livekit/protocol/livekit"
)

type RecordingRepository struct {
	db *sql.DB
}

func NewRecordingRepository(db *sql.DB) *RecordingRepository {
	return &RecordingRepository{db: db}
}

func (r *RecordingRepository) CreateRecording(ctx context.Context, rec *streaming.VODRecording) error {
	query := `
	INSERT INTO recordings (
		id, room_name, streamer_id, streamer_name, title, status, 
		video_path, thumbnail_path, created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.db.ExecContext(ctx, query,
		rec.ID,
		rec.RoomName,
		rec.StreamerID,
		rec.StreamerName,
		rec.Title,
		rec.Status,
		rec.VideoURL, // Storing path/url here
		rec.ThumbnailURL,
		rec.RecordedAt,
		time.Now(),
	)
	if err != nil {
		// Log error if insert fails (since this method returns error, caller should log, but we can verify)
		return err
	}
	return nil
}

func (r *RecordingRepository) UpdateRecordingStatus(ctx context.Context, id string, status streaming.VODStatus, duration time.Duration, size int64, videoPath string) error {
	query := `
	UPDATE recordings 
	SET status = $1, duration = $2, file_size = $3, video_path = $4, updated_at = CURRENT_TIMESTAMP 
	WHERE id = $5`

	_, err := r.db.ExecContext(ctx, query, status, duration, size, videoPath, id)
	return err
}

func (r *RecordingRepository) GetRecording(ctx context.Context, id string) (*streaming.VODRecording, error) {
	query := `
	SELECT id, room_name, streamer_id, streamer_name, title, status, 
	       video_path, thumbnail_path, duration, file_size, created_at
	FROM recordings WHERE id = $1`

	rec := &streaming.VODRecording{}
	var videoPath, thumbPath sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rec.ID, &rec.RoomName, &rec.StreamerID, &rec.StreamerName, &rec.Title, &rec.Status,
		&videoPath, &thumbPath, &rec.Duration, &rec.FileSize, &rec.RecordedAt,
	)
	if err != nil {
		return nil, err
	}

	if videoPath.Valid {
		rec.VideoURL = videoPath.String
	}
	if thumbPath.Valid {
		rec.ThumbnailURL = thumbPath.String
	}

	return rec, nil
}

func (r *RecordingRepository) ListRecordings(ctx context.Context, streamerID livekit.ParticipantIdentity) ([]*streaming.VODRecording, error) {
	var rows *sql.Rows
	var err error

	if streamerID != "" {
		query := `
		SELECT id, room_name, streamer_id, streamer_name, title, status, 
			   video_path, thumbnail_path, duration, file_size, created_at
		FROM recordings 
		WHERE streamer_id = $1 
		ORDER BY created_at DESC`
		rows, err = r.db.QueryContext(ctx, query, streamerID)
	} else {
		query := `
		SELECT id, room_name, streamer_id, streamer_name, title, status, 
			   video_path, thumbnail_path, duration, file_size, created_at
		FROM recordings 
		ORDER BY created_at DESC`
		rows, err = r.db.QueryContext(ctx, query)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	recordings := make([]*streaming.VODRecording, 0)
	for rows.Next() {
		rec := &streaming.VODRecording{}
		var videoPath, thumbPath, streamerName, title sql.NullString
		var duration, fileSize sql.NullInt64
		var createdAt sql.NullTime

		if err := rows.Scan(
			&rec.ID, &rec.RoomName, &rec.StreamerID, &streamerName, &title, &rec.Status,
			&videoPath, &thumbPath, &duration, &fileSize, &createdAt,
		); err != nil {
			// Log the error but keep going? No, usually return error.
			// But let's log it to be sure.
			fmt.Printf("Error scanning recording row: %v\n", err)
			return nil, err
		}

		if streamerName.Valid {
			rec.StreamerName = streamerName.String
		}
		if title.Valid {
			rec.Title = title.String
		}
		if videoPath.Valid {
			rec.VideoURL = videoPath.String
		}
		if thumbPath.Valid {
			rec.ThumbnailURL = thumbPath.String
		}
		if duration.Valid {
			rec.Duration = time.Duration(duration.Int64)
		}
		if fileSize.Valid {
			rec.FileSize = fileSize.Int64
		}
		if createdAt.Valid {
			rec.RecordedAt = createdAt.Time
		} else {
			rec.RecordedAt = time.Now() // Fallback
		}
		recordings = append(recordings, rec)
	}
	fmt.Printf("ListRecordings found %d records\n", len(recordings))
	return recordings, nil
}
