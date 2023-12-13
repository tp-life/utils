package fx

import (
	"context"
	"fmt"
	"reflect"
	"testing"
)

func TestPagingAcquisition_Do(t *testing.T) {
	type fields struct {
		handle PageHandle
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "test",
			fields: fields{
				handle: func(page int64) (bool, error) {
					fmt.Println(page)
					if page == 10 {
						return false, nil
					}
					return true, nil
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pg := &PagingAcquisition{
				handle: tt.fields.handle,
			}
			if err := pg.Do(); (err != nil) != tt.wantErr {
				t.Errorf("PagingAcquisition.Do() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

var s1 = []map[string]string{
	{
		"a": "1",
	},
	{
		"a": "2",
	},
	{
		"a": "3",
	},
	{
		"a": "4",
	},
	{
		"a": "5",
	},
	{
		"a": "6",
	},
	{
		"a": "7",
	},
	{
		"a": "8",
	},
	{
		"a": "9",
	},
}

var s2 = []map[string]string{
	{
		"b": "1",
	},
	{
		"b": "2",
	},
	{
		"b": "3",
	},
	{
		"b": "4",
	},
	{
		"b": "5",
	},
	{
		"b": "6",
	},
	{
		"b": "7",
	},
	{
		"b": "8",
	},
	{
		"b": "9",
	},
	{
		"b": "10",
	},
	{
		"b": "11",
	},
	{
		"b": "12",
	},
	{
		"b": "13",
	},
	{
		"b": "14",
	},
	{
		"b": "15",
	},
	{
		"b": "16",
	},
	{
		"b": "17",
	},
	{
		"b": "18",
	},
	{
		"b": "19",
	},
	{
		"b": "20",
	},
	{
		"b": "21",
	},
	{
		"b": "22",
	},
	{
		"b": "23",
	},
	{
		"b": "24",
	},
	{
		"b": "25",
	},
	{
		"b": "26",
	},
	{
		"b": "27",
	},
	{
		"b": "28",
	},
	{
		"b": "29",
	},
	{
		"b": "30",
	},
	{
		"b": "31",
	},
	{
		"b": "32",
	},
	{
		"b": "33",
	},
}

func s1Page(ctx context.Context, page, pageSize, offset int) (PageFnResp, error) {
	of := (page - 1) * pageSize
	if of >= len(s1) {
		return PageFnResp{Total: len(s1)}, nil
	}
	end := of + pageSize
	if end > len(s1) {
		end = len(s1)
	}
	data := make([]interface{}, 0, pageSize)
	for _, v := range s1[of:end] {
		data = append(data, v)
	}
	return PageFnResp{Total: len(s1), Resp: data}, nil
}

func s1Offset(ctx context.Context, page, pageSize, offset int) (PageFnResp, error) {
	if offset >= len(s1) {
		return PageFnResp{Total: len(s1)}, nil
	}
	end := offset + pageSize
	if end > len(s1) {
		end = len(s1)
	}
	data := make([]interface{}, 0, pageSize)
	for _, v := range s1[offset:end] {
		data = append(data, v)
	}
	return PageFnResp{Total: len(s1), Resp: data}, nil
}

func s1Count(ctx context.Context) int {
	return len(s1)
}

func s2Page(ctx context.Context, page, pageSize, offset int) (PageFnResp, error) {
	of := (page - 1) * pageSize
	if of >= len(s2) {
		return PageFnResp{Total: len(s2)}, nil
	}
	end := of + pageSize
	if end > len(s2) {
		end = len(s2)
	}
	data := make([]interface{}, 0, pageSize)
	for _, v := range s2[of:end] {
		data = append(data, v)
	}
	return PageFnResp{Total: len(s2), Resp: data}, nil
}

func s2Offset(ctx context.Context, page, pageSize, offset int) (PageFnResp, error) {
	if offset >= len(s2) {
		return PageFnResp{Total: len(s2)}, nil
	}
	end := offset + pageSize
	if end > len(s2) {
		end = len(s2)
	}
	data := make([]interface{}, 0, pageSize)
	for _, v := range s2[offset:end] {
		data = append(data, v)
	}
	return PageFnResp{Total: len(s2), Resp: data}, nil
}

func s2Count(ctx context.Context) int {
	return len(s2)
}

func TestPageMerge_mergeCallback(t *testing.T) {
	type fields struct {
		page     int
		pageSize int
		source   []PageFunc
		pageMod  bool
	}
	type args struct {
		ctx context.Context
		fns []callback
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantResp int
	}{
		{
			name: "test",
			fields: fields{
				page:     10,
				pageSize: 10,
				source:   []PageFunc{{SourceName: "a", Fn: s1Page, CountFn: s1Count}, {SourceName: "b", Fn: s2Page, CountFn: s2Count}},
				pageMod:  true,
			},
			args: args{},
		},
		// {
		// 	name: "test2",
		// 	fields: fields{
		// 		page:     2,
		// 		pageSize: 10,
		// 		source:   []PageFunc{{SourceName: "a", Fn: s1Page, CountFn: s1Count}, {SourceName: "b", Fn: s2Page, CountFn: s2Count}},
		// 		pageMod:  true,
		// 	},
		// 	args: args{},
		// },
		// {
		// 	name: "test3",
		// 	fields: fields{
		// 		page:     3,
		// 		pageSize: 10,
		// 		source:   []PageFunc{{SourceName: "a", Fn: s1Page, CountFn: s1Count}, {SourceName: "b", Fn: s2Page, CountFn: s2Count}},
		// 		pageMod:  true,
		// 	},
		// 	args: args{},
		// },
		// {
		// 	name: "test",
		// 	fields: fields{
		// 		page:     1,
		// 		pageSize: 10,
		// 		source:   []PageFunc{{SourceName: "a", Fn: s1Offset, CountFn: s1Count}, {SourceName: "b", Fn: s2Offset, CountFn: s2Count}},
		// 		pageMod:  false,
		// 	},
		// 	args: args{},
		// },
		// {
		// 	name: "test2",
		// 	fields: fields{
		// 		page:     2,
		// 		pageSize: 10,
		// 		source:   []PageFunc{{SourceName: "a", Fn: s1Offset, CountFn: s1Count}, {SourceName: "b", Fn: s2Offset, CountFn: s2Count}},
		// 		pageMod:  false,
		// 	},
		// 	args: args{},
		// },
		// {
		// 	name: "test3",
		// 	fields: fields{
		// 		page:     3,
		// 		pageSize: 10,
		// 		source:   []PageFunc{{SourceName: "a", Fn: s1Offset, CountFn: s1Count}, {SourceName: "b", Fn: s2Offset, CountFn: s2Count}},
		// 		pageMod:  false,
		// 	},
		// 	args: args{},
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PageMerge{
				page:     tt.fields.page,
				pageSize: tt.fields.pageSize,
				source:   tt.fields.source,
				pageMod:  tt.fields.pageMod,
			}
			total, gotResp, _ := p.AsyncMerge(tt.args.ctx)
			if !reflect.DeepEqual(total, tt.wantResp) {
				t.Errorf("PageMerge.mergeCallback() = %v, want %v", gotResp, tt.wantResp)
			}
		})
	}
}

func TestPageMerge_SyncMerge(t *testing.T) {
	type fields struct {
		page     int
		pageSize int
		source   []PageFunc
		pageMod  bool
	}
	type args struct {
		ctx context.Context
		fns []callback
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantResp int
	}{
		// 分页模式
		// {
		// 	name: "test",
		// 	fields: fields{
		// 		page:     1,
		// 		pageSize: 10,
		// 		source:   []PageFunc{{SourceName: "a", Fn: s1Page, CountFn: s1Count}, {SourceName: "b", Fn: s2Page, CountFn: s2Count}},
		// 		pageMod:  true,
		// 	},
		// 	args: args{},
		// },
		// {
		// 	name: "test2",
		// 	fields: fields{
		// 		page:     2,
		// 		pageSize: 10,
		// 		source:   []PageFunc{{SourceName: "a", Fn: s1Page, CountFn: s1Count}, {SourceName: "b", Fn: s2Page, CountFn: s2Count}},
		// 		pageMod:  true,
		// 	},
		// 	args: args{},
		// },
		// {
		// 	name: "test3",
		// 	fields: fields{
		// 		page:     3,
		// 		pageSize: 10,
		// 		source:   []PageFunc{{SourceName: "a", Fn: s1Page, CountFn: s1Count}, {SourceName: "b", Fn: s2Page, CountFn: s2Count}},
		// 		pageMod:  true,
		// 	},
		// 	args: args{},
		// },
		// {
		// 	name: "test4",
		// 	fields: fields{
		// 		page:     4,
		// 		pageSize: 10,
		// 		source:   []PageFunc{{SourceName: "a", Fn: s1Page, CountFn: s1Count}, {SourceName: "b", Fn: s2Page, CountFn: s2Count}},
		// 		pageMod:  true,
		// 	},
		// 	args: args{},
		// },
		// {
		// 	name: "test5",
		// 	fields: fields{
		// 		page:     5,
		// 		pageSize: 10,
		// 		source:   []PageFunc{{SourceName: "a", Fn: s1Page, CountFn: s1Count}, {SourceName: "b", Fn: s2Page, CountFn: s2Count}},
		// 		pageMod:  true,
		// 	},
		// 	args: args{},
		// },
		// offset 模式
		{
			name: "test",
			fields: fields{
				page:     10,
				pageSize: 20,
				source:   []PageFunc{{SourceName: "a", Fn: s1Offset, CountFn: s1Count}, {SourceName: "b", Fn: s2Offset, CountFn: s2Count}},
				pageMod:  false,
			},
			args: args{},
		},
		// {
		// 	name: "test2",
		// 	fields: fields{
		// 		page:     2,
		// 		pageSize: 10,
		// 		source:   []PageFunc{{SourceName: "a", Fn: s1Offset, CountFn: s1Count}, {SourceName: "b", Fn: s2Offset, CountFn: s2Count}},
		// 		pageMod:  false,
		// 	},
		// 	args: args{},
		// },
		// {
		// 	name: "test3",
		// 	fields: fields{
		// 		page:     3,
		// 		pageSize: 10,
		// 		source:   []PageFunc{{SourceName: "a", Fn: s1Offset, CountFn: s1Count}, {SourceName: "b", Fn: s2Offset, CountFn: s2Count}},
		// 		pageMod:  false,
		// 	},
		// 	args: args{},
		// },
		// {
		// 	name: "test4",
		// 	fields: fields{
		// 		page:     4,
		// 		pageSize: 10,
		// 		source:   []PageFunc{{SourceName: "a", Fn: s1Page, CountFn: s1Count}, {SourceName: "b", Fn: s2Offset, CountFn: s2Count}},
		// 		pageMod:  false,
		// 	},
		// 	args: args{},
		// },
		// {
		// 	name: "test5",
		// 	fields: fields{
		// 		page:     5,
		// 		pageSize: 10,
		// 		source:   []PageFunc{{SourceName: "a", Fn: s1Page, CountFn: s1Count}, {SourceName: "b", Fn: s2Offset, CountFn: s2Count}},
		// 		pageMod:  false,
		// 	},
		// 	args: args{},
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PageMerge{
				page:     tt.fields.page,
				pageSize: tt.fields.pageSize,
				source:   tt.fields.source,
				pageMod:  tt.fields.pageMod,
			}
			total, _, _ := p.SyncMerge(tt.args.ctx)
			if !reflect.DeepEqual(total, tt.wantResp) {
				t.Errorf("PageMerge.mergeCallback() = %v, want %v", total, tt.wantResp)
			}
		})
	}
}
