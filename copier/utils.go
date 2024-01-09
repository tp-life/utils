package copier

import (
	"errors"
	"strconv"
	"time"
)

// 预置处理器
func CopyOption[T any](src, dist T, op ...TypeConverter) error {
	err := CopyWithOption(dist, src, Option{
		DeepCopy:    true,
		IgnoreEmpty: true,
		Converters: []TypeConverter{
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
		},
	})
	return err
}
