package httphandler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/DEEJ4Y/genkitkraft/internal/app/commands"
	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
	"github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/playground"
)

// streamPollInterval is how often the tail loop checks for new chunks. Since
// generation is decoupled from any single HTTP request, both the initial
// connection and any reconnect just tail the same persisted chunk log — there
// is no lower-latency channel to prefer over polling that works uniformly
// across instances.
const streamPollInterval = 200 * time.Millisecond

// streamOrphanGrace is added on top of the generation's own bounded timeout
// (commands.StreamGenerationTimeout) before a tailer gives up on a message
// stuck in "streaming" and finalizes it as failed itself. It can only fire
// after the producer's own context has necessarily already ended, so there is
// no race with a still-running goroutine.
const streamOrphanGrace = 10 * time.Second

// parseLastEventID reads the Last-Event-ID header, defaulting to 0 (replay
// from the start) if absent or invalid.
func parseLastEventID(r *http.Request) int {
	v := r.Header.Get("Last-Event-ID")
	if v == "" {
		return 0
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return 0
	}
	return n
}

// getFlusher writes a "streaming not supported" error and returns ok=false if
// w cannot be flushed incrementally.
func getFlusher(w http.ResponseWriter) (http.Flusher, bool) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeAppError(w, errors.NewAppError(errors.Internal, "streaming not supported"))
		return nil, false
	}
	return flusher, true
}

// pollStreamChunks tails a streaming (or already-finished) assistant message
// starting after sinceSeq, invoking onChunk for each new chunk in order and
// onDone exactly once when the message reaches a terminal status ("complete"
// or "error") — including a status it assigns itself if the message is
// orphaned (stuck "streaming" past its own generation timeout, e.g. because
// the instance that started it crashed). Returns without calling onDone if
// the request context is canceled first — the client disconnected from this
// particular response, but generation (owned by a detached context) keeps
// running independently and a later reconnect can pick it up.
//
// messageID may be empty, meaning "resolve the session's latest message on
// the first poll" (the reconnect endpoints don't know it up front); the
// resolved ID is then reused for every subsequent poll so the tail can't
// jump to a different message if a new one starts mid-poll.
func (h *Handler) pollStreamChunks(ctx context.Context, sessionID, messageID string, sinceSeq int, onChunk func(seq int, content string), onDone func(status string)) {
	deadline := time.Now().Add(commands.StreamGenerationTimeout + streamOrphanGrace)
	seq := sinceSeq
	mid := messageID

	for {
		if ctx.Err() != nil {
			return
		}

		var midPtr *string
		if mid != "" {
			midPtr = &mid
		}

		result, err := h.playgroundApp.Queries.GetStreamChunks.Execute(ctx, queries.GetPlaygroundStreamChunksParams{
			SessionID: sessionID,
			MessageID: midPtr,
			SinceSeq:  seq,
		})
		if err != nil {
			onDone(playground.MessageStatusError)
			return
		}
		mid = result.MessageID

		for _, c := range result.Chunks {
			onChunk(c.Seq, c.Content)
			seq = c.Seq
		}

		if result.Status == playground.MessageStatusComplete || result.Status == playground.MessageStatusError {
			onDone(result.Status)
			return
		}

		if time.Now().After(deadline) {
			_ = h.playgroundApp.Commands.FailStream.Execute(ctx, commands.FailPlaygroundStreamParams{MessageID: messageID})
			onDone(playground.MessageStatusError)
			return
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(streamPollInterval):
		}
	}
}

// streamMessage tails messageID and writes it as raw-text SSE frames, the
// format used by the Playground UI. Used for both the initial connection
// (sinceSeq=0) and a Last-Event-ID reconnect.
func (h *Handler) streamMessage(w http.ResponseWriter, r *http.Request, sessionID, messageID string, sinceSeq int) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := getFlusher(w)
	if !ok {
		return
	}

	h.pollStreamChunks(r.Context(), sessionID, messageID, sinceSeq,
		func(seq int, content string) {
			fmt.Fprintf(w, "id: %d\ndata: %s\n\n", seq, escapeSSEData(content))
			flusher.Flush()
		},
		func(status string) {
			if status == playground.MessageStatusError {
				fmt.Fprintf(w, "data: [ERROR] stream interrupted\n\n")
			} else {
				fmt.Fprintf(w, "data: [DONE]\n\n")
			}
			flusher.Flush()
		},
	)
}

// streamDeployCompletion tails messageID and writes it as OpenAI-compatible
// chat.completion.chunk SSE frames, the format used by the Deploy session
// API. Used for both the initial connection (sinceSeq=0) and a Last-Event-ID
// reconnect — a reconnect gets a fresh completionID/created pair since the
// OpenAI wire format has no notion of resuming a completion object; callers
// are expected to accumulate delta.content, not track id stability.
func (h *Handler) streamDeployCompletion(w http.ResponseWriter, r *http.Request, sessionID, messageID, completionID, model string, created int64, sinceSeq int) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := getFlusher(w)
	if !ok {
		return
	}

	h.pollStreamChunks(r.Context(), sessionID, messageID, sinceSeq,
		func(seq int, content string) {
			chunk := deployChatCompletionChunk{
				ID:      completionID,
				Object:  "chat.completion.chunk",
				Created: created,
				Model:   model,
				Choices: []deployChatCompletionChunkChoice{
					{Index: 0, Delta: deployChatDelta{Content: &content}},
				},
			}
			data, _ := json.Marshal(chunk)
			fmt.Fprintf(w, "id: %d\ndata: %s\n\n", seq, data)
			flusher.Flush()
		},
		func(status string) {
			if status == playground.MessageStatusError {
				errData, _ := json.Marshal(map[string]any{
					"error": map[string]string{
						"message": "stream interrupted",
						"type":    "server_error",
					},
				})
				fmt.Fprintf(w, "data: %s\n\n", errData)
				flusher.Flush()
				return
			}

			stopReason := "stop"
			finishChunk := deployChatCompletionChunk{
				ID:      completionID,
				Object:  "chat.completion.chunk",
				Created: created,
				Model:   model,
				Choices: []deployChatCompletionChunkChoice{
					{Index: 0, Delta: deployChatDelta{}, FinishReason: &stopReason},
				},
			}
			finishData, _ := json.Marshal(finishChunk)
			fmt.Fprintf(w, "data: %s\n\n", finishData)
			fmt.Fprintf(w, "data: [DONE]\n\n")
			flusher.Flush()
		},
	)
}
