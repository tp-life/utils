package diff

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/tp-life/utils/copier"
)

type DiffResult struct {
	Change    Changelog
	changeKey map[string]string
}

func NewDiffResult(cg Changelog) *DiffResult {
	dr := &DiffResult{
		Change: cg,
	}
	dr.changeKey = dr.getAll()
	return dr
}

func (dc *DiffResult) GetChangeKey(deep int) (keys map[string]string) {
	keys = make(map[string]string)
	for _, v := range dc.Change {
		if deep > len(v.Path) || deep == 0 {
			deep = len(v.Path)
		}
		k := strings.Join(v.Path[0:deep], ".")
		keys[k] = v.Type
	}
	return
}

func (dc *DiffResult) GetAllKey() (keys map[string]string) {
	if len(dc.changeKey) == 0 {
		return dc.getAll()
	}
	return dc.changeKey
}

func (dc *DiffResult) getAll() (keys map[string]string) {
	keys = make(map[string]string)
	for _, v := range dc.Change {
		for i := range v.Path {
			k := strings.Join(v.Path[0:i+1], ".")
			if _, ok := keys[k]; ok || k == "" {
				continue
			}
			keys[k] = v.Type
		}
	}
	return
}

func (dc *DiffResult) Match(key string) bool {
	if _, ok := dc.changeKey[key]; ok {
		return true
	}
	return false
}

func (dc *DiffResult) MatchOr(key ...string) bool {
	for _, k := range key {
		if _, ok := dc.changeKey[k]; ok {
			return true
		}
	}
	return false
}

func (dc *DiffResult) MatchAnd(key ...string) bool {
	for _, k := range key {
		if _, ok := dc.changeKey[k]; !ok {
			return false
		}
	}
	return true
}

func (dc *DiffResult) MatchDeep(key string) bool {
	for k := range dc.changeKey {
		if strings.HasPrefix(key, k) {
			return true
		}
	}
	return false
}

func (dc *DiffResult) Path(dist any) (ms map[string]any, err error) {
	log := Patch(dc.Change, dist)

	if log.HasErrors() {
		err = errors.New("patch error")
		return
	}

	if log.Applied() {
		ms = make(map[string]any)
		byt, err := json.Marshal(dist)
		if err != nil {
			return ms, err
		}
		if err := json.Unmarshal(byt, &ms); err != nil {
			return ms, err
		}
	}
	return
}

// 指定类型变更的字段
func Diff4Empty[T any](dist T) (*DiffResult, error) {
	src := new(T)
	changelog, err := Diff(*src, dist)
	if err != nil {
		return nil, err
	}
	return NewDiffResult(changelog), nil
}

// 指定类型变更的字段
func Diff4Tag[T any](dist T, tag string) (*DiffResult, error) {
	src := new(T)
	changelog, err := Diff(*src, dist, TagName(tag))
	if err != nil {
		return nil, err
	}
	return NewDiffResult(changelog), nil
}

// 复制两个结构 并比较两个结构体的差异
func CopyWithDiff[T any](src, dist T, opts ...copier.TypeConverter) (dr *DiffResult, err error) {
	err = copier.CopyOption(src, dist, opts...)
	if err != nil {
		return
	}

	change, err := Diff(dist, src)
	if err != nil {
		return nil, err
	}

	dr = NewDiffResult(change)
	return
}
