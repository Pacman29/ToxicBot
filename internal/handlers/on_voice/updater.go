package on_voice

import (
	"context"
	"time"
)

func (h *Handler) reloadVoices() error {
	r, err := h.storage.GetVoices()
	if err != nil {
		return err
	}

	r = r.GetEnabled()

	voices := make([]string, 0, len(r))
	for _, dto := range r {
		voices = append(voices, dto.VoiceID)
	}

	h.muVcs.Lock()
	defer h.muVcs.Unlock()
	h.voices = voices

	return nil
}

func (h *Handler) runUpdater(ctx context.Context) {
	t := time.NewTimer(h.cfg.UpdateVoicesPeriod)
	for {
		select {
		case <-t.C:
			if err := h.reloadVoices(); err != nil {
				h.logger.WithError(err).Warn("cannot reload voices")
			}
		case <-ctx.Done():
			return
		}
	}
}
