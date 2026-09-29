package expirer

import (
	"fmt"
	"time"

	"github.com/kiwiirc/plugin-fileuploader/shardedfilestore"
	"github.com/rs/zerolog"
)

// expireBatchSize bounds how many uploads are loaded per query, so that a
// large backlog is never loaded into memory in one go.
const expireBatchSize = 1000

type Expirer struct {
	ticker        *time.Ticker
	checkInterval time.Duration
	store         *shardedfilestore.ShardedFileStore
	quitChan      chan struct{} // closes when ticker has been stopped
	log           *zerolog.Logger
}

func New(store *shardedfilestore.ShardedFileStore, checkInterval time.Duration, log *zerolog.Logger) *Expirer {
	expirer := &Expirer{
		ticker:        time.NewTicker(checkInterval),
		checkInterval: checkInterval,
		store:         store,
		quitChan:      make(chan struct{}),
		log:           log,
	}

	go func() {
		for {
			select {

			// tick
			case t := <-expirer.ticker.C:
				expirer.gc(t)

			// ticker stopped, exit the goroutine
			case _, ok := <-expirer.quitChan:
				if !ok {
					return
				}

			}
		}
	}()

	return expirer
}

// Stop turns off an Expirer. No more Filestore garbage collection cycles will start.
func (expirer *Expirer) Stop() {
	expirer.ticker.Stop()
	close(expirer.quitChan)
}

func (expirer *Expirer) gc(t time.Time) {
	expirer.log.Debug().
		Str("event", "gc_tick").
		Msg("Filestore GC tick")

	// Spend at most one check interval per pass: with a large backlog, files
	// that expire in the meantime are picked up by the next pass instead of
	// waiting for the whole backlog to drain.
	deadline := time.Now().Add(expirer.checkInterval)
	now := time.Now().Unix()

	terminated, failed, done := expirer.expireNewestFirst("expires_at", "expires_at IS NOT NULL", now, deadline)
	if done {
		// uploads without expires_at (never finished) expire a day after creation
		legacyTerminated, legacyFailed, legacyDone := expirer.expireNewestFirst("created_at", "expires_at IS NULL", now-86400, deadline)
		terminated += legacyTerminated
		failed += legacyFailed
		done = legacyDone
	}

	if terminated > 0 || failed > 0 || !done {
		expirer.log.Info().
			Str("event", "gc_done").
			Int("terminated", terminated).
			Int("failed", failed).
			Bool("backlog_remaining", !done).
			Dur("duration", time.Since(t)).
			Msg("Filestore GC pass finished")
	}
}

// expireNewestFirst terminates the uploads matching cond whose column is at
// most upTo, most recent first: those files are the likeliest to still be on
// disk. Keyset pagination moves past uploads that fail to terminate instead of
// fetching them again. done is false if the deadline stopped the pass early.
func (expirer *Expirer) expireNewestFirst(column, cond string, upTo int64, deadline time.Time) (terminated, failed int, done bool) {
	query := fmt.Sprintf(`
		SELECT id, %[1]s AS ts
		FROM uploads
		WHERE deleted = 0 AND %[2]s AND (
			%[1]s < ? OR (%[1]s = ? AND id < ?)
		)
		ORDER BY %[1]s DESC, id DESC
		LIMIT ?`, column, cond)

	// start just past upTo: "< upTo+1" covers "<= upTo", and no id is < ""
	cursorTs, cursorID := upTo+1, ""

	for time.Now().Before(deadline) {
		var batch []struct {
			ID string `db:"id"`
			Ts int64  `db:"ts"`
		}
		err := expirer.store.DBConn.DB.Select(&batch, query, cursorTs, cursorTs, cursorID, expireBatchSize)
		if err != nil {
			expirer.log.Error().
				Err(err).
				Msg("Failed to enumerate expired uploads")
			return terminated, failed, false
		}

		for _, upload := range batch {
			err = expirer.store.Terminate(upload.ID)
			if err != nil {
				failed++
				expirer.log.Error().
					Err(err).
					Str("id", upload.ID).
					Msg("Failed to terminate expired upload")
				continue
			}
			terminated++
			expirer.log.Info().
				Str("event", "expired").
				Str("id", upload.ID).
				Msg("Terminated upload id")
		}

		if len(batch) < expireBatchSize {
			return terminated, failed, true
		}
		last := batch[len(batch)-1]
		cursorTs, cursorID = last.Ts, last.ID
	}

	return terminated, failed, false
}
