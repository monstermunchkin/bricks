// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.

package log

import (
	"context"
	"net/http/httptest"
	"testing"
)

func TestLog(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	if RequestID(req) != "" {
		t.Error("Request without set error ID can't have a request id")
	}

	Req(req).Info().Msg("req")

	Ctx(context.Background()).Info().Msg("ctx")

	Stack(context.Background())

	Logger().Info().Msg("log")
}
