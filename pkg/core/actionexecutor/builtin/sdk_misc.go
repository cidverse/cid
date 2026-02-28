package builtin

import (
	"fmt"
	"log/slog"

	"github.com/cidverse/cid/pkg/core/actionsdk"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func (sdk ActionSDK) HealthV1() (actionsdk.HealthV1Response, error) {
	return actionsdk.HealthV1Response{Status: "up"}, nil
}

func (sdk ActionSDK) LogV1(req actionsdk.LogV1Request) error {
	// get level
	lvl := zerolog.DebugLevel
	if req.Level == "info" {
		lvl = zerolog.InfoLevel
	} else if req.Level == "warn" {
		lvl = zerolog.WarnLevel
	} else if req.Level == "error" {
		lvl = zerolog.ErrorLevel
	}

	// log message with context
	ev := log.WithLevel(lvl)
	if req.Context != nil {
		for k, v := range req.Context {
			ev.Interface(k, v)
		}
	}

	msgPrefix := ""
	if sdk.CurrentAction != nil {
		msgPrefix = fmt.Sprintf("[%s] ", sdk.CurrentAction.Metadata.Name)
	}
	ev.Msg(msgPrefix + req.Message)

	return nil
}

// UUIDV4 generates a new UUID string.
func (sdk ActionSDK) UUIDV4() string {
	u, err := uuid.NewUUID()
	if err != nil {
		slog.With("err", err).Error("failed to generate UUIDV4")
	}
	return u.String()
}
