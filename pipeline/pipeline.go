package pipeline

import (
	"context"
	"github.com/pixie-sh/database-helpers-go/pipeline/operators"
	"github.com/pixie-sh/errors-go/utils"
	"github.com/pixie-sh/logger-go/logger"
	"sync"
)

// Pipeline something amazing
type Pipeline struct {
	passable      operators.Result
	operators     []operators.Operator
	operatorsType uint //one pipeline shall be able to deal with one type of operators.Operator
	once          sync.Once
	log           logger.Interface
}

// NewPipeline something amazing, is it?
func NewPipeline(log logger.Interface) *Pipeline {
	return &Pipeline{once: sync.Once{}, log: log}
}

// AddOperator if already existing operators type mismatch error will be returned
func (p *Pipeline) AddOperator(operator ...operators.Operator) *Pipeline {
	if utils.Nil(operator) || len(operator) == 0 {
		p.log.Warn("operator not added, nil pointer or empty")
		return p
	}

	for _, o := range operator {
		if utils.Nil(o) {
			p.log.Warn(
				"operator not added, nil pointer or empty; ", p.operatorsType)
			continue
		}

		p.once.Do(func() {
			p.operatorsType = o.GetType()
		})

		if p.operatorsType == o.GetType() {
			p.operators = append(p.operators, o)
			return p
		}

		p.log.Warn(
			"operator not added. type mismatch, expected:%d, given:%d",
			p.operatorsType,
			o.GetType())
	}

	return p
}

// WithPassable something amazing... who knows....
func (p *Pipeline) WithPassable(passable operators.Result) *Pipeline {
	p.passable = passable
	return p
}

// ExecWithPassable something amazing... who knows....
func (p *Pipeline) ExecWithPassable(ctx context.Context, passable operators.Result) (operators.Result, error) {
	return p.WithPassable(passable).Exec(ctx)
}

// Exec something amazing... hopefully?
func (p *Pipeline) Exec(ctx context.Context) (operators.Result, error) {
	for _, operator := range p.operators {
		if operator.Predicate(ctx, false) {
			result, err := operator.Handle(ctx, p.passable)
			if err != nil {
				p.log.With("error", err).Error("error iterating operator")

				return nil, err
			}

			p.passable = result
		}
	}

	return p.passable, nil
}
