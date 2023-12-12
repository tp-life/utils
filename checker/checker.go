package checker

import (
	"context"
	"fmt"
	"github.com/tp-life/utils/dag"
	"log/slog"

	"github.com/samber/lo"
)

// PlaceCheckHandler 下单检测handler
type PlaceCheckHandler func(ctx context.Context, o any) (bool, error)

const initCheckParams = "checkParams"

// checkOp Operator
type checkOp int8

const (
	// opAnd and
	opAnd checkOp = iota
	// opOr or
	opOr
)

type checkOpCol struct {
	flag checkOp
	cols []string
}

// Checker 校验器
// NOTE:: 不要使用单例模式实例化
type Checker struct {
	dag  *dag.FxDag
	cols []checkOpCol
}

// NewChecker 初始化
func NewChecker(params any) *Checker {
	dg := dag.New()
	if params == nil {
		params = struct{}{}
	}
	_ = dg.InitParamsByName(initCheckParams, params)
	return &Checker{dag: dg}
}

// CheckAndResult 校验
func (ck *Checker) CheckAndResult(ctx context.Context) (b bool, err error) {
	err = ck.Check(ctx)
	if err != nil {
		return
	}
	b = ck.Result()
	if !b {
		slog.InfoContext(ctx, "CheckAndResult Checker check fail", slog.Any("checker", ck.cols))
	}
	return
}

// Check 校验
func (ck *Checker) Check(ctx context.Context) (err error) {
	err = ck.dag.DrawAndExec(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Checker check fail", err)
	}
	return
}

// And 并且
func (ck *Checker) And(handlers ...PlaceCheckHandler) error {
	ansOp, err := ck.genCheckOp("and", handlers...)
	if err != nil {
		return err
	}
	ck.cols = append(ck.cols, checkOpCol{flag: opAnd, cols: ansOp})
	return nil
}

// Or 或者
func (ck *Checker) Or(handlers ...PlaceCheckHandler) error {
	ansOp, err := ck.genCheckOp("or", handlers...)
	if err != nil {
		return err
	}
	ck.cols = append(ck.cols, checkOpCol{flag: opOr, cols: ansOp})
	return nil
}

func (ck *Checker) genCheckOp(name string, handlers ...PlaceCheckHandler) (
	ansOp []string,
	err error,
) {

	for i, v := range handlers {
		v := v
		n := fmt.Sprintf("%s_%d_%d", name, len(ck.cols), i)
		err = ck.RegisterHandler(n, func(ctx context.Context, dag *dag.FxDag) (bool, error) {
			r, _ := dag.Load(initCheckParams)
			return v(ctx, r)
		})
		if err != nil {
			return
		}
		ansOp = append(ansOp, n)
	}

	return
}

// RegisterHandler 注册handler
func (ck *Checker) RegisterHandler(name string, handler dag.FxHandler[bool], des ...string) error {
	return dag.ProvideByName[bool](
		ck.dag,
		name,
		handler,
		des...,
	)
}

// OneOfResult 任一验证不通过则返回
func (ck *Checker) OneOfResult() bool {
	val := ck.dag.LoadAll()
	for _, v := range val {
		r, ok := v.(bool)
		if !ok || !r {
			return false
		}
	}
	return true
}

// Result 获取结果集
func (ck *Checker) Result() (result bool) {

	for _, cols := range ck.cols {
		s := ck.getVal(cols.cols)
		switch cols.flag {
		case opAnd:
			result = !lo.Contains(s, false)
		case opOr:
			result = lo.Contains(s, true)
		}
		if !result {
			break
		}
	}

	return
}

func (ck *Checker) getVal(names []string) []bool {
	var result []bool

	for _, v := range names {
		r, ok := ck.dag.Load(v)
		if !ok {
			result = append(result, false)
			break
		}
		b, ok := r.(bool)
		if !ok {
			result = append(result, false)

			break
		}
		result = append(result, b)
	}

	return result
}
