package dbutil

import (
	"fmt"
	"strings"
)

type IPage[T any] interface {
	GetRecords() []T              // 获取查询的记录
	SetRecords([]T)               // 设置查询的记录
	GetTotal() int64              // 获取总记录数
	SetTotal(int64)               // 设置总记录数
	GetCurrent() int64            // 获取当前页
	SetCurrent(int64)             // 设置当前页
	GetSize() int64               // 获取每页显示大小
	SetSize(int64)                // 设置每页显示大小
	AddOrderItem(OrderItem)       // 设置排序条件
	AddOrderItems([]OrderItem)    // 批量设置排序条件
	GetOrederItemsString() string // 将排序条件拼接成字符串
	Offset() int64                // 获取偏移量
	GetPages() int64              // 获取总的分页数
}

type Page[T any] struct {
	total   int64       // 总的记录数
	size    int64       // 每页显示的大小
	current int64       // 当前页
	orders  []OrderItem // 排序条件
	Records []T         // 查询数据列表

}

func (page *Page[T]) GetRecords() []T {
	return page.Records
}

func (page *Page[T]) SetRecords(records []T) {
	page.Records = records
}

func (page *Page[T]) GetTotal() int64 {
	return page.total
}

func (page *Page[T]) SetTotal(total int64) {
	page.total = total

}

func (page *Page[T]) GetCurrent() int64 {
	return page.current
}

func (page *Page[T]) SetCurrent(current int64) {
	page.current = current
}

func (page *Page[T]) GetSize() int64 {
	return page.size
}
func (page *Page[T]) SetSize(size int64) {
	page.size = size

}

func (page *Page[T]) AddOrderItem(orderItem OrderItem) {
	page.orders = append(page.orders, orderItem)
}

func (page *Page[T]) AddOrderItems(orderItems []OrderItem) {
	page.orders = append(page.orders, orderItems...)
}

func (page *Page[T]) GetOrederItemsString() string {
	arr := make([]string, 0)
	var order string

	for _, val := range page.orders {
		if val.Asc {
			order = ""
		} else {
			order = "desc"
		}
		arr = append(arr, fmt.Sprintf("%s %s", val.Column, order))
	}
	return strings.Join(arr, ",")
}

func (page *Page[T]) Offset() int64 {
	if page.GetCurrent() > 0 {
		return (page.GetCurrent() - 1) * page.GetSize()
	} else {
		return 0
	}
}

func (page *Page[T]) GetPages() int64 {
	if page.GetSize() == 0 {
		return 0
	}
	pages := page.GetTotal() / page.GetSize()
	if page.GetTotal()%page.size != 0 {
		pages++
	}

	return pages
}

type OrderItem struct {
	Column string // 需要排序的字段
	Asc    bool   // 是否正序排列，默认true
}

func (orderItem OrderItem) String() string {
	if orderItem.Asc {
		return fmt.Sprintf("%s asc", orderItem.Column)
	} else {
		return fmt.Sprintf("%s desc", orderItem.Column)
	}
}

func BuildAsc(column string) OrderItem {
	return OrderItem{Column: column, Asc: true}
}

func BuildDesc(column string) OrderItem {
	return OrderItem{Column: column, Asc: false}
}

func BuildAscs(columns ...string) []OrderItem {
	items := make([]OrderItem, 0)
	for _, val := range columns {
		items = append(items, BuildAsc(val))
	}
	return items
}

func BuildDescs(columns ...string) []OrderItem {
	items := make([]OrderItem, 0)
	for _, val := range columns {
		items = append(items, BuildDesc(val))
	}
	return items
}

func NewPage[T any](size, current int64, orderItems ...OrderItem) *Page[T] {
	return &Page[T]{size: size, current: current, orders: orderItems}
}
