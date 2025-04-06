package repository

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"wechat-robot-client/pkg/gormx"
	"wechat-robot-client/pkg/gtool"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type where = map[string]interface{}

type scope = func(*gorm.DB) *gorm.DB

type Base[T schema.Tabler] struct {
	Ctx         context.Context
	DB          *gorm.DB
	notFoundMsg string
}

func (b *Base[T]) GetByExistsID(id int64, preloads ...string) *T {
	v := b.GetByID(id, preloads...)
	if v == nil {
		b.panicNotFound()
	}
	return v
}

func (b *Base[T]) GetByID(id int64, preloads ...string) *T {
	return b.getOneByWhere(preloads, where{"id": id})
}

func (b *Base[T]) GetByCode(code interface{}, preloads ...string) *T {
	return b.getOneByWhere(preloads, where{"code": code})
}

func (b *Base[T]) Get(condition *T, preloads ...string) *T {
	v := new(T)
	query := b.DB
	for _, preload := range preloads {
		query = query.Preload(preload)
	}
	err := query.Where(condition).First(v).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		b.panicError(err)
	}
	return v
}

func (b *Base[T]) GetByCodes(codes interface{}, preloads ...string) []*T {
	return b.listByWhere(preloads, where{"code": codes})
}

func (b *Base[T]) GetByExistsCode(code interface{}, preloads ...string) *T {
	v := b.GetByCode(code, preloads...)
	if v == nil {
		b.panicNotFound()
	}
	return v
}

func (b *Base[T]) ListAll(preloads ...string) []*T {
	return b.list(preloads)
}

func (b *Base[T]) ListAllOrder(order string, preloads ...string) []*T {
	return b.list(preloads, orderFunc(order))
}

func (b *Base[T]) ListAllUnscoped(preloads ...string) []*T {
	return b.list(preloads, unScopedFunc())
}

func (b *Base[T]) ListByIDs(ids []int64, preloads ...string) []*T {
	return b.list(preloads, idsFunc(ids))
}

func (b *Base[T]) ExistsById(id int64) bool {
	return b.ExistsByWhere(where{"id": id})
}

func (b *Base[T]) ExistsByCode(code interface{}) bool {
	return b.ExistsByWhere(where{"code": code})
}

func (b *Base[T]) ExistsByName(name, objectCode string) bool {
	w := where{"name": name}
	if objectCode != "" {
		w["object_code"] = objectCode
	}
	return b.ExistsByWhere(w)
}

func (b *Base[T]) Update(value *T) {
	b.panicError(b.DB.Updates(value).Error)
}

func (b *Base[T]) Upsert(value []*T) {
	b.panicError(b.DB.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&value).Error)
}

func (b *Base[T]) UpdateInBatches(values []*T) {
	for _, value := range values {
		b.Update(value)
	}
}

func (b *Base[T]) UpdateColumnsByIds(updates interface{}, ids []int64) {
	b.updateColumnsByWhere(updates, where{"id": ids})
}

func (b *Base[T]) Create(value *T) {
	b.panicError(b.DB.Create(value).Error)
}

func (b *Base[T]) CreateOmitAssociations(value *T) {
	b.panicError(b.DB.Omit(clause.Associations).Create(value).Error)
}

func (b *Base[T]) CreateInBatches(values []*T, batchSize int) {
	if len(values) == 0 {
		return
	}
	if batchSize <= 0 {
		batchSize = len(values)
	}
	b.panicError(b.model().CreateInBatches(&values, batchSize).Error)
}

func (b *Base[T]) CreateInBatchesWithClauses(values []*T, batchSize int, clauses ...clause.Expression) {
	if batchSize <= 0 {
		batchSize = len(values)
	}
	b.panicError(b.model().Clauses(clauses...).CreateInBatches(&values, batchSize).Error)
}

func (b *Base[T]) Save(value *T) {
	b.panicError(b.DB.Omit(clause.Associations).Save(value).Error)
}

func (b *Base[T]) SaveInBatches(value []*T) {
	b.panicError(b.DB.Omit(clause.Associations).Save(value).Error)
}

func (b *Base[T]) DeleteById(id int64) {
	b.panicError(b.DB.Delete(new(T), id).Error)
}

func (b *Base[T]) DeleteByIds(ids []int64) {
	b.delete(idsFunc(ids))
}

func (b *Base[T]) DeleteByCode(code interface{}) {
	b.panicError(b.DB.Where("code = ?", code).Delete(new(T)).Error)
}

func (b *Base[T]) DeleteByCodes(codes interface{}) {
	b.delete(codesFunc(codes))
}

func (b *Base[T]) model() *gorm.DB {
	return b.DB.Model(new(T))
}

func (b *Base[T]) Pluck(column string, s ...scope) []*T {
	list := make([]*T, 0)
	query := b.model().Scopes(s...)
	b.panicError(query.Distinct().Pluck(column, &list).Error)
	return list
}

func (b *Base[T]) delete(s ...scope) {
	b.panicError(b.DB.Scopes(s...).Delete(new(T)).Error)
}

func (b *Base[T]) DeleteByWhere(query interface{}, args ...interface{}) {
	b.deleteByWhere(query, args)
}

func (b *Base[T]) deleteByWhere(query interface{}, args ...interface{}) {
	b.panicError(b.DB.Where(query, args...).Delete(new(T)).Error)
}

func (b *Base[T]) List(preloads []string, s ...scope) []*T {
	list := make([]*T, 0)
	query := b.model().Scopes(s...)
	for _, preload := range preloads {
		query = query.Preload(preload)
	}
	b.panicError(query.Find(&list).Error)
	return list
}

func (b *Base[T]) list(preloads []string, s ...scope) []*T {
	list := make([]*T, 0)
	query := b.model().Scopes(s...)
	for _, preload := range preloads {
		query = query.Preload(preload)
	}
	b.panicError(query.Find(&list).Error)
	return list
}

func (b *Base[T]) ListByWhere(preloads []string, query interface{}, args ...interface{}) []*T {
	return b.list(preloads, func(db *gorm.DB) *gorm.DB {
		return db.Where(query, args)
	})
}

func (b *Base[T]) listByWhere(preloads []string, query interface{}, args ...interface{}) []*T {
	return b.list(preloads, func(db *gorm.DB) *gorm.DB {
		return db.Where(query, args)
	})
}

func (b *Base[T]) columnInt64ByWhere(column string, query interface{}, args ...interface{}) []int64 {
	values := make([]int64, 0)
	b.panicError(b.model().Where(query, args...).Pluck(column, &values).Error)
	return values
}

func (b *Base[T]) UpdateColumnsByWhere(updates interface{}, query interface{}, args ...interface{}) {
	b.panicError(b.model().Where(query, args...).UpdateColumns(updates).Error)
}

func (b *Base[T]) updateColumnsByWhere(updates interface{}, query interface{}, args ...interface{}) {
	b.panicError(b.model().Where(query, args...).UpdateColumns(updates).Error)
}

func (b *Base[T]) countByWhere(groupBy string, query interface{}, args ...interface{}) (count int64) {
	q := b.model().Where(query, args)
	if groupBy != "" {
		q = q.Group(groupBy)
	}
	err := q.Count(&count).Error
	if err != nil {
		b.panicError(err)
	}
	return count
}

func (b *Base[T]) countByScope(groupBy string, s ...scope) (count int64) {
	query := b.model().Scopes(s...)
	if groupBy != "" {
		query = query.Group(groupBy)
	}
	b.panicError(query.Count(&count).Error)
	return count
}

func (b *Base[T]) CountByScope(groupBy string, s ...scope) (count int64) {
	return b.countByScope(groupBy, s...)
}

func (b *Base[T]) GetOne(preloads []string, s ...scope) *T {
	return b.getOne(preloads, s...)
}

func (b *Base[T]) getOne(preloads []string, s ...scope) *T {
	v := new(T)
	query := b.DB.Scopes(s...).Limit(1)
	for _, preload := range preloads {
		query = query.Preload(preload)
	}
	if err := query.First(v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		b.panicError(err)
	}
	return v
}

func (b *Base[T]) takeOne(preloads []string, s ...scope) *T {
	v := new(T)
	query := b.DB.Scopes(s...)
	for _, preload := range preloads {
		query = query.Preload(preload)
	}
	if err := query.Take(v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		b.panicError(err)
	}
	return v
}

func (b *Base[T]) getOneByWhere(preloads []string, query interface{}, args ...interface{}) *T {
	return b.getOne(preloads, func(db *gorm.DB) *gorm.DB {
		return db.Where(query, args)
	})
}

func (b *Base[T]) ExistsByWhere(query interface{}, args ...interface{}) bool {
	return b.getOneByWhere(nil, query, args) != nil
}

func (b *Base[T]) exists(s ...scope) bool {
	return b.getOne(nil, s...) != nil
}

func (b *Base[T]) panicError(err error) {
	if err != nil {
		panic(err)
	}
}

type transHandler func(transDB *gorm.DB)

func DoInTrans(ctx context.Context, originDB *gorm.DB, f transHandler) {
	transManager := &gormx.GormUnitOfWork{}
	transDB, err := transManager.BeginTran(gtool.WithOrmContext(ctx, originDB))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := transManager.Rollback(transDB); err != nil && err != sql.ErrTxDone {
			log.Printf("Rollback err:%v", err)
		}
	}()
	f(transDB)
	if err := transManager.Commit(transDB); err != nil {
		panic(err)
	}
}

func pagerFunc(pageIndex, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(pageSize * (pageIndex - 1)).
			Limit(pageSize)
	}
}

func orderFunc(order string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order(order)
	}
}

func idsFunc(ids []int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("id IN (?)", ids)
	}
}

func unScopedFunc() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Unscoped()
	}
}

func codesFunc(codes interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("code IN (?)", codes)
	}
}

func groupByFunc(groupBy string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Group(groupBy)
	}
}

func PagerFunc(pageIndex, pageSize int) func(db *gorm.DB) *gorm.DB {
	return pagerFunc(pageIndex, pageSize)
}

func OrderFunc(order string) func(db *gorm.DB) *gorm.DB {
	return orderFunc(order)
}

func IdsFunc(ids []int64) func(db *gorm.DB) *gorm.DB {
	return idsFunc(ids)
}

func UnScopedFunc() func(db *gorm.DB) *gorm.DB {
	return unScopedFunc()
}

func CodesFunc(codes interface{}) func(db *gorm.DB) *gorm.DB {
	return codesFunc(codes)
}

func GroupByFunc(groupBy string) func(db *gorm.DB) *gorm.DB {
	return groupByFunc(groupBy)
}

func (b *Base[T]) panicNotFound() {
	if b.notFoundMsg != "" {
		panic(errors.New(b.notFoundMsg))
	}
	panic(errors.New("not found"))
}
