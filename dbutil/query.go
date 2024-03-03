package dbutil

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type A []any

type D []A

// =======condition========

// 自定义sql查询
type Condition struct {
	list []*conditionInfo
}

func (c *Condition) AndWithCondition(condition bool, column string, cases string, value any) *Condition {
	if condition {
		c.list = append(c.list, &conditionInfo{
			andor:  "and",
			column: column, // 列名
			case_:  cases,  // 条件(and,or,in,>=,<=)
			value:  value,
		})
	}
	return c
}

// And a Condition by and .and 一个条件
func (c *Condition) And(column string, cases string, value interface{}) *Condition {
	return c.AndWithCondition(true, column, cases, value)
}

func (c *Condition) OrWithCondition(condition bool, column string, cases string, value interface{}) *Condition {
	if condition {
		c.list = append(c.list, &conditionInfo{
			andor:  "or",
			column: column, // 列名
			case_:  cases,  // 条件(and,or,in,>=,<=, =)
			value:  value,
		})
	}
	return c
}

// Or a Condition by or .or 一个条件
func (c *Condition) Or(column string, cases string, value interface{}) *Condition {
	return c.OrWithCondition(true, column, cases, value)
}

func (c *Condition) Get() (where string, out []interface{}) {
	firstAnd := -1
	for i := 0; i < len(c.list); i++ { // 查找第一个and
		if c.list[i].andor == "and" {
			where = fmt.Sprintf("`%v` %v ?", c.list[i].column, c.list[i].case_)
			out = append(out, c.list[i].value)
			firstAnd = i
			break
		}
	}

	if firstAnd < 0 && len(c.list) > 0 { // 补刀
		where = fmt.Sprintf("`%v` %v ?", c.list[0].column, c.list[0].case_)
		out = append(out, c.list[0].value)
		firstAnd = 0
	}

	for i := 0; i < len(c.list); i++ { // 添加剩余的
		if firstAnd != i {
			where += fmt.Sprintf(" %v `%v` %v ?", c.list[i].andor, c.list[i].column, c.list[i].case_)
			out = append(out, c.list[i].value)
		}
	}

	return
}

type conditionInfo struct {
	andor  string
	column string // 列名
	case_  string // 条件(in,>=,<=)
	value  interface{}
}

// ====== option ========

type Options struct {
	Query map[string]interface{}
}

// Option overrides behavior of Connect.
type Option interface {
	Apply(*Options)
}

type OptionFunc func(*Options)

func (f OptionFunc) Apply(o *Options) {
	f(o)
}

// WithName name获取
func WithFileName(name string, val any) Option {
	return OptionFunc(func(o *Options) { o.Query[name] = val })
}

// GetByOption 功能选项模式获取
func GetByOption[T any](db *gorm.DB, opts ...Option) (result T, err error) {
	options := Options{
		Query: make(map[string]interface{}, len(opts)),
	}
	for _, o := range opts {
		o.Apply(&options)
	}

	r := new(T)
	err = db.Where(options.Query).First(r).Error
	result = *r
	return
}

// GetByOptions 批量功能选项模式获取
func GetByOptions[T any](db *gorm.DB, opts ...Option) (results []T, err error) {
	options := Options{
		Query: make(map[string]interface{}, len(opts)),
	}
	for _, o := range opts {
		o.Apply(&options)
	}

	err = db.Where(options.Query).Find(&results).Error

	return
}

// SelectPage 分页查询
func SelectPage[T any](db *gorm.DB, page IPage[T], cond *Condition, opts ...Option) (resultPage IPage[T], err error) {

	resultPage = page
	results := make([]T, 0)

	options := Options{
		Query: make(map[string]interface{}, len(opts)),
	}
	for _, o := range opts {
		o.Apply(&options)
	}

	var count int64 // 统计总的记录数
	if cond != nil {
		s, c := cond.Get()
		db = db.Where(s, c...)
	}
	query := db.Where(options.Query)
	if err = query.Count(&count).Error; err != nil {
		return
	}

	if (count) == 0 {
		return
	}

	resultPage.SetTotal(count)
	if len(page.GetOrederItemsString()) > 0 {
		query = query.Order(page.GetOrederItemsString())
	}

	err = query.Limit(int(page.GetSize())).Offset(int(page.Offset())).Find(&results).Error

	resultPage.SetRecords(results)
	return
}

func SliceWhere[T any](db *gorm.DB, field string, val []T) *gorm.DB {
	if len(val) > 0 {
		if len(val) == 1 {
			db = db.Where(fmt.Sprintf("%s = ?", field), val[0])
		} else {
			db = db.Where(fmt.Sprintf("%s IN ?", field), val)
		}
	}
	return db
}

func SliceCondition[T any](cond *Condition, field string, val []T) *Condition {
	if cond == nil {
		return cond
	}
	if len(val) > 0 {
		if len(val) == 1 {
			cond.And(field, "=", val[0])
		} else {
			cond.And(field, "IN", val)
		}
	}

	return cond
}

// GetFromID 通过id获取内容
func GetFromID[T any](db *gorm.DB, id uint) (result T, err error) {
	r := new(T)
	err = db.Where("`id` = ?", id).First(r).Error
	result = *r
	return
}

// GetBatchFromID 批量查找
func GetBatchFromID[T any](db *gorm.DB, ids []uint) (results []T, err error) {
	err = db.Where("`id` IN (?)", ids).Find(&results).Error
	return
}

// GetFromID 通过id获取内容
func GetFromField[T, F any](db *gorm.DB, name string, v F) (result T, err error) {
	if name == "" {
		return result, errors.New("name is empty")
	}
	r := new(T)
	err = db.Where("`"+name+"` = ?", v).First(r).Error
	result = *r
	return
}

// GetBatchFromField 批量查找
func GetBatchFromField[T, F any](db *gorm.DB, name string, v ...F) (results []T, err error) {
	db = SliceWhere[F](db, name, v)
	err = db.Find(&results).Error
	return
}

// GetBatchFromFields 批量查找
func GetBatchFromFields[T any](db *gorm.DB, sd D) (results []T, err error) {
	for _, v := range sd {
		if len(v) < 2 {
			return nil, errors.New("len(v) != 2")
		}
		name, ok := v[0].(string)
		if !ok {
			return nil, errors.New("search key is not string")
		}
		db = SliceWhere(db, name, v[1:])
	}
	err = db.Find(&results).Error
	return
}

// GetBatchFromPage 获取分页数据
func GetBatchFromPage[T any](db *gorm.DB, size, page int64, cond *Condition, orderItem ...OrderItem) (result []T, total int64, err error) {
	p := NewPage[T](size, page, orderItem...)

	ipg, err := SelectPage(db, p, cond)
	if err != nil {
		return
	}
	opg := ipg.GetRecords()
	result = opg
	total = ipg.GetTotal()
	return
}

// GetBatchFromPage 获取分页数据
func GetBatchFromCondition[T any](db *gorm.DB, cond *Condition) (result []T, err error) {
	if cond != nil {
		s, c := cond.Get()
		db = db.Where(s, c...)
	}
	err = db.Find(&result).Error
	return
}
