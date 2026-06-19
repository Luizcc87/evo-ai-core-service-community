// Package runtimecontext is the public runtime-scope extension point of
// the community release. See EXTENSION_POINTS.md at the repository root.
package runtimecontext

import (
	"context"
	"database/sql"
)

// Scope resolves the runtime scope identifier bound to a given request
// or background job. Implementations must be safe for concurrent use.
//
// The returned string is opaque to the community release; an empty
// string means "no scope bound", which is the standalone case.
type Scope interface {
	CurrentID(ctx context.Context) string
}

type noop struct{}

func (noop) CurrentID(context.Context) string { return "" }

// Default returns the no-op scope used when no extension is installed.
// It always reports the empty string, preserving the community
// release's single-scope behaviour.
func Default() Scope { return noop{} }

type ctxKey int

const idKey ctxKey = 0

// WithID returns a copy of ctx carrying the given runtime-scope id. The
// community release does not call this itself; the helper exists so the
// enterprise build can bridge its tenant binding into a context value
// that downstream community code can read without importing the
// enterprise SDK.
//
// Nil-ctx contract: we return the same nil rather than fabricating a
// Background. A caller that passes nil is already broken — their next
// ctx.Value()/ctx.Done() call panics regardless — so we keep the bug
// at the original call site instead of papering over it with a
// synthetic root.
func WithID(ctx context.Context, id string) context.Context {
	if ctx == nil {
		return ctx
	}
	return context.WithValue(ctx, idKey, id)
}

// IDFromContext returns the runtime-scope id bound to ctx via WithID,
// or "" when no id is bound. Use this in community code paths that
// need to propagate the active scope (eg. server-to-server header
// propagation) without coupling to any specific Scope implementation.
func IDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v, ok := ctx.Value(idKey).(string); ok {
		return v
	}
	return ""
}

const connKey ctxKey = 1

// ScopedConn is the minimal connection contract a scope-bound read must
// run on. It is declared HERE (community, neutral) — NOT imported from
// any enterprise SDK — so community read adapters can route a statement
// onto a scope-bound connection without coupling to the enterprise
// binding implementation. *sql.Tx (the enterprise per-request,
// GUC-carrying tx) satisfies this, and it is exactly gorm.ConnPool, so a
// GORM statement's ConnPool can be set to it directly.
type ScopedConn interface {
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// WithConn returns a copy of ctx carrying a scope-bound connection. Like
// WithID, the community release does not call this itself; the
// enterprise build bridges its per-request GUC-carrying tx here so
// community read adapters (the tenant_scope GORM plugin) can route reads
// onto it WITHOUT importing the enterprise SDK. nil-ctx contract matches
// WithID (return the same nil; a nil ctx is already a caller bug).
func WithConn(ctx context.Context, conn ScopedConn) context.Context {
	if ctx == nil {
		return ctx
	}
	return context.WithValue(ctx, connKey, conn)
}

// ConnFromContext returns the scope-bound connection set via WithConn,
// or (nil, false) when none is bound. Community read adapters use this
// to fail-closed (refuse the read) on a scoped table when no bound
// connection is present, rather than falling through to an unscoped
// pool.
func ConnFromContext(ctx context.Context) (ScopedConn, bool) {
	if ctx == nil {
		return nil, false
	}
	c, ok := ctx.Value(connKey).(ScopedConn)
	return c, ok && c != nil
}
