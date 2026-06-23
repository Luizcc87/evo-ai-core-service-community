//go:build enterprise

// Package tenantstamp is the enterprise GORM plugin that stamps
// tenant_id on every INSERT into evo_core_* tables, mirroring the
// SQLAlchemy before_flush listener in PY-3 (evo-enterprise-licensing-
// python/src/evo_enterprise_licensing/tenant_stamp.py).
//
// The plugin lives under //go:build enterprise so the community
// release never imports it and the standalone build keeps its
// single-scope behaviour unchanged.
//
// Fail-closed: when runtimecontext.IDFromContext(ctx) returns "" the
// plugin does NOT set the column. The INSERT then carries tenant_id
// = uuid.Nil, which the gem-owned RLS policy
//
//	USING (tenant_id = current_setting('app.current_tenant_id', true)::uuid)
//
// rejects with "new row violates row-level security policy". The Go
// layer never invents a tenant id — Postgres is the source of truth
// for the binding contract.
package tenantstamp

import (
	"context"
	"errors"
	"reflect"

	"evo-ai-core-service/pkg/evoextensions/runtimecontext"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// columnName is the column the gem's migrations add to each
// evo_core_* table. Keeping it as a constant (not a per-model tag
// lookup) lets the plugin stay model-agnostic.
const columnName = "tenant_id"

// ErrScopeUnbound is the fail-closed sentinel raised when a schemaless
// tenant-scoped table (allowlist below) is written with no scope-bound
// connection in context — refusing rather than inserting onto the pool
// (empty GUC → the column DEFAULT would read NULL → NOT NULL violation,
// or worse a row bound to no tenant).
var ErrScopeUnbound = errors.New("tenantstamp: schemaless tenant write with no bound connection")

// schemalessTenantTables: tabelas onde o evo-core ESCREVE via struct mas cujo
// struct community NÃO declara a coluna tenant_id (a migration enterprise do gem
// adicionou-a NOT NULL + RLS no MESMO Postgres do CRM). LookUpField(tenant_id)
// retorna nil → o stamp normal (preencher o VALOR no struct) não tem onde escrever.
//
// SOLUÇÃO (simétrica ao tenantscope dos reads): em vez de carimbar o valor, ROTEAMOS
// o INSERT para a tx GUC-carrying per-request (db.Statement.ConnPool = conn), onde o
// Authorizer enterprise já fez set_config('app.current_tenant_id', tid, is_local). Aí
// o DEFAULT da coluna (migration do gem: tenant_id DEFAULT current_setting(...)) lê o
// tenant correto da tx. struct-create intacto → bot.ID volta via RETURNING. O write
// normalmente roda no pool global com GUC vazio (só os reads eram roteados); isto o
// roteia para tabelas do allowlist. NUNCA tocar o struct community (decisão: tenant_id
// é eixo enterprise). agent_bots é o único caso hoje (os demais structs já declaram
// tenant_id e seguem o caminho de stamp normal).
//
// CUIDADO: estritamente allowlist — não re-rotear writes de outras tabelas (mudaria
// a conexão/atomicidade delas sem motivo).
var schemalessTenantTables = map[string]struct{}{
	"agent_bots": {},
}

// callbackName must be unique across registered Create callbacks.
const callbackName = "evo:tenant_stamp"

// Plugin implements gorm.Plugin.
type Plugin struct{}

// Name returns the plugin identity used by GORM's plugin registry.
func (Plugin) Name() string { return callbackName }

// Initialize registers a Before("gorm:create") callback that stamps
// the tenant_id column on every INSERT when the bound model exposes
// that field.
func (Plugin) Initialize(db *gorm.DB) error {
	return db.Callback().Create().Before("gorm:create").Register(callbackName, stamp)
}

// stamp is the callback body. It is a no-op when:
//   - the statement has no parsed schema (raw SQL / Exec paths),
//   - the model does not declare a tenant_id column,
//   - the caller already set a non-zero tenant_id (seeders, backfill),
//   - no tenant id is bound to the request context (fail-closed).
func stamp(db *gorm.DB) {
	if db.Statement == nil || db.Statement.Schema == nil {
		return
	}

	ctx := db.Statement.Context
	if ctx == nil {
		return
	}

	field := db.Statement.Schema.LookUpField(columnName)
	if field == nil {
		// O struct não declara tenant_id. Se a TABELA é do allowlist schemaless
		// (agent_bots), NÃO podemos preencher o valor no struct — em vez disso
		// ROTEAMOS o INSERT para a tx GUC-carrying per-request (igual o tenantscope
		// faz nos reads), onde o GUC app.current_tenant_id está setado, e o DEFAULT
		// da coluna (migration do gem) preenche o tenant. Fora do allowlist, no-op.
		routeSchemalessTenantWrite(db, ctx)
		return
	}
	tid := runtimecontext.IDFromContext(ctx)
	if tid == "" {
		// Fail-closed: leave tenant_id at uuid.Nil; the RLS policy
		// rejects the INSERT with "new row violates row-level
		// security policy". This is the documented AC for EVO-1624.
		return
	}
	parsed, err := uuid.Parse(tid)
	if err != nil {
		// A bad value in ctx is a programmer error upstream; refusing
		// to guess keeps the RLS rejection signal honest.
		return
	}

	rv := reflect.Indirect(db.Statement.ReflectValue)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < rv.Len(); i++ {
			elem := reflect.Indirect(rv.Index(i))
			if elem.Kind() == reflect.Map {
				stampMap(db, elem, parsed)
				continue
			}
			setIfZero(db, field, elem, parsed)
		}
	case reflect.Struct:
		setIfZero(db, field, rv, parsed)
	case reflect.Map:
		stampMap(db, rv, parsed)
	}
}

// routeSchemalessTenantWrite roteia o INSERT de uma tabela do allowlist schemaless
// (struct sem tenant_id, tabela COM tenant_id NOT NULL + RLS) para a conexão
// scope-bound publicada pela camada enterprise (runtimecontext.ConnFromContext) —
// a tx onde o Authorizer fez set_config('app.current_tenant_id', tid, is_local).
// Nessa tx o DEFAULT da coluna (migration do gem) resolve o tenant correto e o
// struct-create segue intacto (RETURNING id popula bot.ID). É o simétrico de WRITE
// do tenantscope (que roteia os reads). FAIL-CLOSED: tabela do allowlist sem tenant
// no ctx OU sem conn scope-bound → ABORTA (não insere no pool com GUC vazio, o que
// gravaria a row sem tenant ou violaria NOT NULL silenciosamente). Tabelas fora do
// allowlist: no-op (mantém o comportamento anterior).
func routeSchemalessTenantWrite(db *gorm.DB, ctx context.Context) {
	if _, ok := schemalessTenantTables[db.Statement.Table]; !ok {
		return // fora do allowlist → não nos interessa
	}
	// tenant precisa estar bound (igual o tenantscope): sem tenant → fail-closed.
	if runtimecontext.IDFromContext(ctx) == "" {
		_ = db.AddError(ErrScopeUnbound)
		return
	}
	conn, ok := runtimecontext.ConnFromContext(ctx)
	if !ok {
		// tenant bound mas a conn scope-bound não está no ctx (rota que furou o
		// middleware enterprise). Recusar em vez de inserir no pool (GUC vazio →
		// DEFAULT NULL → NOT NULL violation, ou row órfã).
		_ = db.AddError(ErrScopeUnbound)
		return
	}
	// Roteia ESTE INSERT para a tx GUC-carrying. O DEFAULT da coluna lê o GUC dela.
	db.Statement.ConnPool = conn
}

// setIfZero writes parsed into the tenant_id field of elem only when
// the field is at its zero value. field.ValueOf returns (value, isZero);
// we drop the value and branch on isZero so callers that explicitly
// pre-populate tenant_id (seeders, backfill jobs) are not clobbered.
func setIfZero(db *gorm.DB, field *schema.Field, elem reflect.Value, parsed uuid.UUID) {
	if !elem.IsValid() {
		return
	}
	_, isZero := field.ValueOf(db.Statement.Context, elem)
	if !isZero {
		return
	}
	_ = field.Set(db.Statement.Context, elem, parsed)
}

// stampMap handles the map[string]interface{} Create path. GORM allows
// `db.Model(&X{}).Create(map[string]interface{}{...})` for ad-hoc
// inserts; the struct-based stamper above never sees those rows because
// ReflectValue.Kind() is reflect.Map. We mirror setIfZero's "don't
// clobber" rule: only set the key when it's absent or empty.
func stampMap(db *gorm.DB, m reflect.Value, parsed uuid.UUID) {
	if !m.IsValid() || m.IsNil() {
		return
	}
	if m.Type().Key().Kind() != reflect.String {
		return
	}
	// Guard against panic when the map's value type isn't interface{} and
	// isn't directly assignable from uuid.UUID (eg. map[string]string).
	// Such Create patterns are unusual but legal; we'd rather no-op than
	// crash the request.
	elemType := m.Type().Elem()
	if elemType.Kind() != reflect.Interface && !reflect.TypeOf(parsed).AssignableTo(elemType) {
		return
	}
	key := reflect.ValueOf(columnName)
	if existing := m.MapIndex(key); existing.IsValid() {
		v := reflect.ValueOf(existing.Interface())
		if v.IsValid() && !v.IsZero() {
			return
		}
	}
	m.SetMapIndex(key, reflect.ValueOf(parsed))
}
