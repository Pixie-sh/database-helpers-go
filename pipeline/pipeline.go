package pipeline

import (
	"github.com/pixie-sh/errors-go/utils"
	"github.com/pixie-sh/logger-go/logger"
	"sync"
)

// Operator something amazing... or not.
type Operator interface {
	Handle(result Result) (Result, error)
	Predicate() bool
	GetType() uint
}

// Result interface to avoid circular imports
type Result interface {
	WithPassable(passable interface{})
	GetPassable() interface{}
	Error() error
}

// Pipeline something amazing
type Pipeline struct {
	passable      Result
	operators     []Operator
	operatorsType uint //one pipeline shall be able to deal with one type of Operator
	once          sync.Once
}

// NewPipeline something amazing, is it?
func NewPipeline() *Pipeline {
	return &Pipeline{once: sync.Once{}}
}

// AddOperator something amazing... who knows....
func (p *Pipeline) AddOperator(operator Operator) *Pipeline {
	if utils.Nil(operator) {
		logger.Logger.Warn(
			"operator not added, nil pointer",
			p.operatorsType,
			operator.GetType())

		return p
	}

	p.once.Do(func() {
		p.operatorsType = operator.GetType()
	})

	if p.operatorsType == operator.GetType() {
		p.operators = append(p.operators, operator)
		return p
	}

	logger.Logger.Warn(
		"operator not added. type mismatch, expected:%d, given:%d",
		p.operatorsType,
		operator.GetType())
	return p
}

// WithPassable something amazing... who knows....
func (p *Pipeline) WithPassable(passable Result) *Pipeline {
	p.passable = passable
	return p
}

// ExecWithPassable something amazing... who knows....
func (p *Pipeline) ExecWithPassable(passable Result) (Result, error) {
	return p.WithPassable(passable).Exec()
}

// Exec something amazing... hopefully?
func (p *Pipeline) Exec() (Result, error) {
	for _, operator := range p.operators {
		if operator.Predicate() {
			result, err := operator.Handle(p.passable)
			if err != nil {
				return nil, err
			}

			p.passable = result
		}
	}

	return p.passable, nil
}
