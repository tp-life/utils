package fx

import (
	"context"
	"sort"

	"golang.org/x/sync/errgroup"
)

type PageHandle func(page int64) (bool, error)

// PagingAcquisition 分页获取数据
// 通过页码获取数据
type PagingAcquisition struct {
	handle PageHandle
}

// NewPagingAcquisition 初始化
func NewPagingAcquisition(handle PageHandle) *PagingAcquisition {
	return &PagingAcquisition{
		handle: handle,
	}
}

// 这段代码定义了一个名为 Do() 的函数，用于执行分页获取。该函数首先将当前页数设置为1，然后不断调用一个名为 handle() 的函数，并将当前页数作为参数传入，直到 handle() 函数返回 false 为止。如果出现错误，函数将返回一个错误；否则返回 nil。
func (pg *PagingAcquisition) Do() error {
	var page int64 = 1
	for {
		b, err := pg.handle(page)
		if err != nil {
			return err
		}
		if !b {
			break
		}
		page += 1
	}
	return nil
}

type (
	// callback 返回格式
	PageFnResp struct {
		Total int
		Resp  []interface{}
	}
	// PageFunc 列表分页函数
	// NOTE: Fn 函数处理分页，除第一数据源以外，其他数据源需要使用offset 进行数据分页
	PageFunc struct {
		SourceName string
		Fn         func(ctx context.Context, page, pageSize, offset int) (PageFnResp, error)
		CountFn    func(context.Context) int
	}

	// Result 返回结果
	Result struct {
		Resp       interface{}
		SourceName string
	}
)

type (
	callback struct {
		fn       PageFunc
		offset   int
		page     int
		pageSize int
		mod      int
		source   string
	}
	mergeResult struct {
		index int
		data  []*Result
	}
)

// PageMerge 合并分页
type PageMerge struct {
	page     int
	pageSize int
	source   []PageFunc
	pageMod  bool // 是否为页码模式
}

func NewPageMerge(page int, pageSize int, source []PageFunc, pageMod ...bool) *PageMerge {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 1
	}
	pm := false
	if len(pageMod) > 0 {
		pm = pageMod[0]
	}
	return &PageMerge{page: page, pageSize: pageSize, source: source, pageMod: pm}
}

// SyncMerge 同步获取数据 仅支持offset 模式
func (p *PageMerge) SyncMerge(ctx context.Context) (int, []*Result, error) {
	page, pageSize := p.page, p.pageSize
	offset := (page - 1) * pageSize
	finalOffset := offset
	total := 0
	resp := make([]*Result, 0, pageSize)
	for _, fn := range p.source {

		mod, ps := 0, pageSize
		pg := offset/ps + 1
		if p.pageMod {
			mod = offset % p.pageSize // 起始数据位 需要补充的数据
			ps = mod + pageSize       // 增加每页查询的数据，防止在上一数据源数据不足，进行了数据补位后，后续内容按页查询时数据重复。
			pg = offset/ps + 1
			if (offset-pageSize) > 0 && (offset-pageSize)%ps == 0 {
				pg += 1
			}
		}

		result, err := fn.Fn(ctx, pg, ps, offset)
		if err != nil {
			return 0, nil, err
		}
		total += result.Total
		if p.pageSize > len(resp) {
			for i, v := range result.Resp {
				if i < mod && pg == 1 {
					continue
				}
				resp = append(resp, &Result{SourceName: fn.SourceName, Resp: v})
			}
		}
		limit := p.pageSize - len(resp)
		if limit > 0 {
			offset = p.page*pageSize - total - limit // 当前页 结束时 offset 位置
			if offset < 0 {
				offset = 0
			}
		} else {
			offset = 0
		}

	}
	if total < p.page*p.pageSize && total > finalOffset && len(resp)-(total-finalOffset) > 0 {
		resp = resp[len(resp)-(total-finalOffset):]
	}
	if len(resp) > p.pageSize {
		resp = resp[0:p.pageSize]
	}
	return total, resp, nil
}

// AsyncMerge 异步查询合并数据
func (p *PageMerge) AsyncMerge(ctx context.Context) (int, []*Result, error) {

	type (
		countRe struct {
			Index int
			Count int
		}
	)

	page, pageSize := p.page, p.pageSize
	offset := (page - 1) * pageSize
	total := 0
	resp := make([]*Result, 0, pageSize)
	wait := errgroup.Group{}

	ch := make(chan countRe, len(p.source))
	// 计算count 总数
	for i, fn := range p.source {
		ind := i
		f := fn.CountFn
		wait.Go(func() error {
			t := f(ctx)
			ch <- countRe{Index: ind, Count: t}
			return nil
		})
	}
	if err := wait.Wait(); err != nil {
		return total, resp, err
	}
	close(ch)
	sourceCount := make(map[int]int)
	for v := range ch {
		sourceCount[v.Index] = v.Count
		total += v.Count
	}

	excFn := make([]callback, 0, len(p.source))
	dataCount := 0
	// 查询指定的几个数据源
	for i, v := range p.source {
		c := sourceCount[i]
		if c == 0 {
			continue
		}
		dataCount += c
		if dataCount > offset {
			// 计算offset 偏移量，总的偏移量 减去 截止到上一数据源总的数据条数
			of := offset - (dataCount - c)
			if of < 0 {
				of = 0
			}

			mod, ps := 0, p.pageSize
			if p.pageMod {
				mod = of % p.pageSize
				ps = mod + p.pageSize // 增加每页查询的数据，防止在上一数据源数据不足，进行了数据补位后，后续内容按页查询时数据重复。
			}
			pg := of/ps + 1
			excFn = append(excFn, callback{offset: of, fn: v, page: pg, pageSize: ps, mod: mod, source: v.SourceName})
		}
		if dataCount >= page*pageSize {
			break
		}
	}
	resp, err := p.mergeCallback(ctx, excFn)
	return total, resp, err

}

func (p *PageMerge) mergeCallback(ctx context.Context, fns []callback) (resp []*Result, err error) {
	if len(fns) == 0 {
		return
	}

	erGo := errgroup.Group{}
	ch := make(chan mergeResult, len(fns))
	for i, v := range fns {
		cb := v.fn.Fn
		p, s, o, mod, source := v.page, v.pageSize, v.offset, v.mod, v.source
		index := i
		erGo.Go(func() error {
			r, err := cb(ctx, p, s, o)
			if err != nil {
				return err
			}
			result := make([]*Result, 0, len(r.Resp))
			for i := range r.Resp {
				if i < mod && p == 1 {
					continue
				}
				result = append(result, &Result{SourceName: source, Resp: r.Resp[i]})
			}
			ch <- mergeResult{index: index, data: result}
			return nil
		})
	}
	if err := erGo.Wait(); err != nil {
		return nil, err
	}
	close(ch)
	mr := make([]mergeResult, 0)

	for v := range ch {
		mr = append(mr, v)
	}

	sort.Slice(mr, func(i, j int) bool {
		return mr[i].index < mr[j].index
	})

	for _, v := range mr {
		resp = append(resp, v.data...)
	}

	if len(resp) > p.pageSize {
		resp = resp[0:p.pageSize]
	}

	return

}

type (
	// PreNextCallback 上一条 下一条数据
	PreNextCallback func(ctx context.Context, curId string) (PreNextResp, bool, error)

	// PreNextResp 上一条 下一条 返回数据
	PreNextResp struct {
		Pre  string
		Next string
	}
)

// PreAndNext 上一条下一条
func (p *PageMerge) PreAndNext(ctx context.Context, curId string, source []PreNextCallback) (result PreNextResp, err error) {
	if len(source) == 0 {
		return
	}

	hasCurrent := false

	for _, fn := range source {
		if hasCurrent {
			curId = ""
		}
		pnr, has, err := fn(ctx, curId)
		if err != nil {
			return result, err
		}
		hasCurrent = has
		if pnr.Pre == "" && pnr.Next == "" {
			continue
		}
		if pnr.Pre != "" {
			result.Pre = pnr.Pre
		}

		if pnr.Next != "" {
			result.Next = pnr.Next
			break
		}
	}
	return
}
