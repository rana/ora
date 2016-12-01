// Copyright 2016 Tamás Gulácsi. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

import "context"

const (
	stmtCfgKey = "stmtCfg"
)

// ctxStmtCfg returns the StmtCfg from the context, and
// whether it exist at all.
func ctxStmtCfg(ctx context.Context) (StmtCfg, bool) {
	cfg, ok := ctx.Value(stmtCfgKey).(StmtCfg)
	return cfg, ok
}

// WithStmtCfg returns a new context, with the given cfg that
// can be used to configure several parameters.
//
// WARNING: the StmtCfg must be derived from Cfg(), or NewStmtCfg(),
// as an empty StmtCfg is not usable!
func WithStmtCfg(ctx context.Context, cfg StmtCfg) context.Context {
	return context.WithValue(ctx, stmtCfgKey, cfg)
}
