package dbutil

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
)

var ErrType = errors.New("type error")

// IsNotFound ErrRecordNotFound
func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// prepare for other
type BaseMgr[T any] struct {
	DB  *gorm.DB
	Ctx context.Context
}

// SetCtx set context
func (obj *BaseMgr[T]) SetCtx(c context.Context) {
	if c != nil {
		obj.Ctx = c
	}
}

// SetPreload set preload
func (obj *BaseMgr[T]) SetPreload(preload ...string) *BaseMgr[T] {
	for _, v := range preload {
		obj.DB = obj.DB.Preload(v)
	}
	return obj
}

// GetCtx get context
func (obj *BaseMgr[T]) GetCtx() context.Context {
	return obj.Ctx
}

// GetDB get gorm.DB info
func (obj *BaseMgr[T]) GetDB() *gorm.DB {
	return obj.DB
}

// UpdateDB update gorm.DB info
func (obj *BaseMgr[T]) UpdateDB(db *gorm.DB) {
	obj.DB = db
}

func (obj *BaseMgr[T]) SetSort(order ...OrderItem) *BaseMgr[T] {
	orderString := make([]string, 0)
	for _, v := range order {
		orderString = append(orderString, v.String())
	}
	obj.GetDB().Order(strings.Join(orderString, ","))
	return obj
}

// New new gorm.新gorm,重置条件
func (obj *BaseMgr[T]) New() {
	obj.DB = obj.NewDB()
}

// NewDB new gorm.新gorm
func (obj *BaseMgr[T]) NewDB() *gorm.DB {
	return obj.GetDB().Session(&gorm.Session{NewDB: true, Context: obj.Ctx})
}

// IsNotFound ErrRecordNotFound
func (obj *BaseMgr[T]) IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func (obj *BaseMgr[T]) GetBatchFromCondition(cond *Condition) ([]T, error) {
	return GetBatchFromCondition[T](obj.GetDB(), cond)
}

// GetBatchFromPage 获取分页数据
func (obj *BaseMgr[T]) GetBatchFromPage(size, page int64, cond *Condition, orderItem ...OrderItem) (result []T, total int64, err error) {
	return GetBatchFromPage[T](obj.GetDB(), size, page, cond, orderItem...)
}

// GetFromID 通过id获取内容
func (obj *BaseMgr[T]) GetFromID(id uint) (result T, err error) {
	return GetFromID[T](obj.GetDB(), id)
}

// GetBatchFromID 批量查找
func (obj *BaseMgr[T]) GetBatchFromID(ids []uint) (results []T, err error) {
	return GetBatchFromID[T](obj.GetDB(), ids)
}

// GetBatchFromField 批量查找
func (obj *BaseMgr[T]) GetBatchFromField(name string, v any) (result []T, err error) {
	return GetBatchFromField[T](obj.GetDB(), name, v)
}

// GetFromField 通过指定条件查询获取内容
func (obj *BaseMgr[T]) GetFromField(name string, v any) (result T, err error) {
	return GetFromField[T](obj.GetDB(), name, v)
}

// GetBatchFromFields 批量通过多个条件查询获取内容
func (obj *BaseMgr[T]) GetBatchFromFields(cond D) (result []T, err error) {
	return GetBatchFromFields[T](obj.GetDB(), cond)
}

// GetByOption 功能选项模式获取
func (obj *BaseMgr[T]) GetByOption(opts ...Option) (result T, err error) {
	return GetByOption[T](obj.GetDB(), opts...)
}

// GetByOptions 批量功能选项模式获取
func (obj *BaseMgr[T]) GetByOptions(opts ...Option) (results []T, err error) {
	return GetByOptions[T](obj.GetDB(), opts...)
}
