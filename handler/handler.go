package handler

import (
	"context"
	"log/slog"
)

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Expire1Request(ctx context.Context) error {
	slog.Info("Expire1Request")
	return nil
}

func (h *Handler) Expire2Response(ctx context.Context) error {
	slog.Info("Expire2Response")
	return nil
}
