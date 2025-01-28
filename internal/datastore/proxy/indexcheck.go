package proxy

import (
	"context"
	"fmt"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"

	"github.com/authzed/spicedb/pkg/datastore"
	"github.com/authzed/spicedb/pkg/datastore/options"
	"github.com/authzed/spicedb/pkg/datastore/queryshape"
	core "github.com/authzed/spicedb/pkg/proto/core/v1"
	"github.com/authzed/spicedb/pkg/tuple"
)

// NewIndexCheckingDatastoreProxy returns a datastore proxy that runs EXPLAIN ANALYZE on all
// relationships queries and ensures that the index(es) used within match those defined in the
// schema for the datastore.
func NewIndexCheckingDatastoreProxy(d datastore.Datastore) datastore.Datastore {
	return &indexcheckingProxy{delegate: d}
}

type indexcheckingProxy struct{ delegate datastore.Datastore }

func (p *indexcheckingProxy) SnapshotReader(rev datastore.Revision) datastore.Reader {
	delegateReader := p.delegate.SnapshotReader(rev)
	return &indexcheckingReader{delegateReader}
}

func (p *indexcheckingProxy) ReadWriteTx(
	ctx context.Context,
	f datastore.TxUserFunc,
	opts ...options.RWTOptionsOption,
) (datastore.Revision, error) {
	return p.delegate.ReadWriteTx(ctx, func(ctx context.Context, delegateRWT datastore.ReadWriteTransaction) error {
		return f(ctx, &indexcheckingRWT{&indexcheckingReader{delegateRWT}, delegateRWT})
	}, opts...)
}

func (p *indexcheckingProxy) OptimizedRevision(ctx context.Context) (datastore.Revision, error) {
	return p.delegate.OptimizedRevision(ctx)
}

func (p *indexcheckingProxy) CheckRevision(ctx context.Context, revision datastore.Revision) error {
	return p.delegate.CheckRevision(ctx, revision)
}

func (p *indexcheckingProxy) HeadRevision(ctx context.Context) (datastore.Revision, error) {
	return p.delegate.HeadRevision(ctx)
}

func (p *indexcheckingProxy) RevisionFromString(serialized string) (datastore.Revision, error) {
	return p.delegate.RevisionFromString(serialized)
}

func (p *indexcheckingProxy) Watch(ctx context.Context, afterRevision datastore.Revision, options datastore.WatchOptions) (<-chan *datastore.RevisionChanges, <-chan error) {
	return p.delegate.Watch(ctx, afterRevision, options)
}

func (p *indexcheckingProxy) Features(ctx context.Context) (*datastore.Features, error) {
	return p.delegate.Features(ctx)
}

func (p *indexcheckingProxy) OfflineFeatures() (*datastore.Features, error) {
	return p.delegate.OfflineFeatures()
}

func (p *indexcheckingProxy) Statistics(ctx context.Context) (datastore.Stats, error) {
	return p.delegate.Statistics(ctx)
}

func (p *indexcheckingProxy) Unwrap() datastore.Datastore {
	return p.delegate
}

func (p *indexcheckingProxy) ReadyState(ctx context.Context) (datastore.ReadyState, error) {
	return p.delegate.ReadyState(ctx)
}

func (p *indexcheckingProxy) Close() error { return p.delegate.Close() }

type indexcheckingReader struct{ delegate datastore.Reader }

func (r *indexcheckingReader) CountRelationships(ctx context.Context, name string) (int, error) {
	return r.delegate.CountRelationships(ctx, name)
}

func (r *indexcheckingReader) LookupCounters(ctx context.Context) ([]datastore.RelationshipCounter, error) {
	return r.delegate.LookupCounters(ctx)
}

func (r *indexcheckingReader) ReadCaveatByName(ctx context.Context, name string) (*core.CaveatDefinition, datastore.Revision, error) {
	return r.delegate.ReadCaveatByName(ctx, name)
}

func (r *indexcheckingReader) LookupCaveatsWithNames(ctx context.Context, caveatNames []string) ([]datastore.RevisionedCaveat, error) {
	return r.delegate.LookupCaveatsWithNames(ctx, caveatNames)
}

func (r *indexcheckingReader) ListAllCaveats(ctx context.Context) ([]datastore.RevisionedCaveat, error) {
	return r.delegate.ListAllCaveats(ctx)
}

func (r *indexcheckingReader) ListAllNamespaces(ctx context.Context) ([]datastore.RevisionedNamespace, error) {
	return r.delegate.ListAllNamespaces(ctx)
}

func (r *indexcheckingReader) LookupNamespacesWithNames(ctx context.Context, nsNames []string) ([]datastore.RevisionedNamespace, error) {
	return r.delegate.LookupNamespacesWithNames(ctx, nsNames)
}

func (r *indexcheckingReader) ReadNamespaceByName(ctx context.Context, nsName string) (*core.NamespaceDefinition, datastore.Revision, error) {
	return r.delegate.ReadNamespaceByName(ctx, nsName)
}

func (r *indexcheckingReader) QueryRelationships(ctx context.Context, filter datastore.RelationshipsFilter, opts ...options.QueryOptionsOption) (datastore.RelationshipIterator, error) {
	opts = append(opts, options.WithSQLExplainCallback(func(ctx context.Context, sql string, args []any, shape queryshape.Shape, explain string, expectedIndexes options.SQLIndexInformation) {
		fmt.Println("SQL:", sql)
		fmt.Println(explain)
	}))
	return r.delegate.QueryRelationships(ctx, filter, opts...)
}

func (r *indexcheckingReader) ReverseQueryRelationships(ctx context.Context, subjectsFilter datastore.SubjectsFilter, opts ...options.ReverseQueryOptionsOption) (datastore.RelationshipIterator, error) {
	opts = append(opts, options.WithSQLExplainCallbackForReverse(func(ctx context.Context, sql string, args []any, shape queryshape.Shape, explain string, expectedIndexes options.SQLIndexInformation) {
		fmt.Println("SQL:", sql)
		fmt.Println(explain)
	}))
	return r.delegate.ReverseQueryRelationships(ctx, subjectsFilter, opts...)
}

type indexcheckingRWT struct {
	*indexcheckingReader
	delegate datastore.ReadWriteTransaction
}

func (rwt *indexcheckingRWT) RegisterCounter(ctx context.Context, name string, filter *core.RelationshipFilter) error {
	return rwt.delegate.RegisterCounter(ctx, name, filter)
}

func (rwt *indexcheckingRWT) UnregisterCounter(ctx context.Context, name string) error {
	return rwt.delegate.UnregisterCounter(ctx, name)
}

func (rwt *indexcheckingRWT) StoreCounterValue(ctx context.Context, name string, value int, computedAtRevision datastore.Revision) error {
	return rwt.delegate.StoreCounterValue(ctx, name, value, computedAtRevision)
}

func (rwt *indexcheckingRWT) WriteCaveats(ctx context.Context, caveats []*core.CaveatDefinition) error {
	return rwt.delegate.WriteCaveats(ctx, caveats)
}

func (rwt *indexcheckingRWT) DeleteCaveats(ctx context.Context, names []string) error {
	return rwt.delegate.DeleteCaveats(ctx, names)
}

func (rwt *indexcheckingRWT) WriteRelationships(ctx context.Context, mutations []tuple.RelationshipUpdate) error {
	return rwt.delegate.WriteRelationships(ctx, mutations)
}

func (rwt *indexcheckingRWT) WriteNamespaces(ctx context.Context, newConfigs ...*core.NamespaceDefinition) error {
	return rwt.delegate.WriteNamespaces(ctx, newConfigs...)
}

func (rwt *indexcheckingRWT) DeleteNamespaces(ctx context.Context, nsNames ...string) error {
	return rwt.delegate.DeleteNamespaces(ctx, nsNames...)
}

func (rwt *indexcheckingRWT) DeleteRelationships(ctx context.Context, filter *v1.RelationshipFilter, options ...options.DeleteOptionsOption) (bool, error) {
	return rwt.delegate.DeleteRelationships(ctx, filter, options...)
}

func (rwt *indexcheckingRWT) BulkLoad(ctx context.Context, iter datastore.BulkWriteRelationshipSource) (uint64, error) {
	return rwt.delegate.BulkLoad(ctx, iter)
}

var (
	_ datastore.Datastore            = (*indexcheckingProxy)(nil)
	_ datastore.Reader               = (*indexcheckingReader)(nil)
	_ datastore.ReadWriteTransaction = (*indexcheckingRWT)(nil)
)
