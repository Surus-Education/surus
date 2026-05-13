package jobs

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/erc-pham/surus/api/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LinkRotChecker struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewLinkRotChecker(pool *pgxpool.Pool) *LinkRotChecker {
	return &LinkRotChecker{pool: pool, queries: db.New(pool)}
}

func (c *LinkRotChecker) Run(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	c.check(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.check(ctx)
		}
	}
}

func (c *LinkRotChecker) check(ctx context.Context) {
	videos, err := c.queries.GetAllVideoLessons(ctx)
	if err != nil {
		slog.Error("link rot checker: failed to get videos", "err", err)
		return
	}

	slog.Info("link rot checker: starting", "count", len(videos))

	for _, video := range videos {
		select {
		case <-ctx.Done():
			return
		default:
		}

		broken := c.checkVideo(video.ProviderID)
		if err := c.queries.SetLessonEmbedBroken(ctx, db.SetLessonEmbedBrokenParams{
			ID:          video.LessonID,
			EmbedBroken: broken,
		}); err != nil {
			slog.Error("link rot checker: failed to update lesson", "lesson_id", video.LessonID, "err", err)
		}

		time.Sleep(time.Second)
	}
}

func (c *LinkRotChecker) checkVideo(providerID string) bool {
	url := fmt.Sprintf("https://www.youtube.com/oembed?url=https://www.youtube.com/watch?v=%s&format=json", providerID)
	resp, err := http.Get(url)
	if err != nil {
		return true
	}
	defer resp.Body.Close()
	return resp.StatusCode != http.StatusOK
}
