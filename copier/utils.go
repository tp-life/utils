package copier

import (
	"errors"
	"strconv"
	"time"

	"github.com/tp-life/utils/diff"
)

// 预置处理器
func CopyOption(src, dist any, op ...TypeConverter) error {

	opts := []TypeConverter{
		{
			SrcType: time.Time{},
			DstType: String,
			Fn: func(src interface{}) (interface{}, error) {
				s, ok := src.(time.Time)

				if !ok {
					return nil, errors.New("src type not matching")
				}

				return s.Format(time.DateTime), nil
			},
		},
		{
			SrcType: String,
			DstType: time.Time{},
			Fn: func(src interface{}) (interface{}, error) {
				s, ok := src.(string)

				if !ok {
					return nil, errors.New("src type not matching")
				}

				location, err := time.ParseInLocation(time.DateTime, s, time.Local)

				if err != nil {
					return nil, errors.New("src type not matching time")
				}

				return location, nil
			},
		},
		{
			SrcType: String,
			DstType: Int,
			Fn: func(src interface{}) (interface{}, error) {
				s, ok := src.(string)
				if !ok {
					return nil, errors.New("src type not matching")
				}

				return strconv.Atoi(s)
			},
		},
	}

	err := CopyWithOption(dist, src, Option{
		DeepCopy:    true,
		IgnoreEmpty: true,
		Converters:  append(opts, op...),
	})
	return err
}

// 复制两个结构 并比较两个结构体的差异
func CopyWithDiff[T any](src, dist T, opts ...TypeConverter) (dr *diff.DiffResult, err error) {
	err = CopyOption(src, dist, opts...)
	if err != nil {
		return
	}

	change, err := diff.Diff(dist, src)
	if err != nil {
		return nil, err
	}

	dr = diff.NewDiffResult(change)
	return
}
