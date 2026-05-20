package elastic

import (
	"fmt"
	"strings"

	"github.com/pixie-sh/errors-go"
)

type ScriptSection string

const (
	ScriptSectionHelpers ScriptSection = "helpers"
	ScriptSectionSetup   ScriptSection = "setup"
	ScriptSectionScoring ScriptSection = "scoring"
	ScriptSectionReturn  ScriptSection = "return"
)

type ScriptBuilder struct {
	helpers []string
	setup   []string
	scoring []string
	ret     string
	params  map[string]interface{}
}

func NewScriptBuilder() *ScriptBuilder {
	return &ScriptBuilder{
		setup:  []string{"double score = 0.0;"},
		ret:    "return Math.max(score, 0.001);",
		params: make(map[string]interface{}),
	}
}

func (b *ScriptBuilder) Add(section ScriptSection, source string) *ScriptBuilder {
	source = strings.TrimSpace(source)
	if source == "" {
		return b
	}

	switch section {
	case ScriptSectionHelpers:
		b.helpers = append(b.helpers, source)
	case ScriptSectionSetup:
		b.setup = append(b.setup, source)
	case ScriptSectionScoring:
		b.scoring = append(b.scoring, source)
	case ScriptSectionReturn:
		b.ret = source
	}

	return b
}

func (b *ScriptBuilder) AddParams(params map[string]interface{}) error {
	for key, value := range params {
		if _, exists := b.params[key]; exists {
			return errors.New("duplicate script param %s", key)
		}
		b.params[key] = value
	}
	return nil
}

func (b *ScriptBuilder) Params() map[string]interface{} {
	params := make(map[string]interface{}, len(b.params))
	for key, value := range b.params {
		params[key] = value
	}
	return params
}

func (b *ScriptBuilder) Source() string {
	sections := make([]string, 0, 4)
	appendSection := func(parts []string) {
		if len(parts) == 0 {
			return
		}
		sections = append(sections, strings.Join(parts, "\n\n"))
	}

	appendSection(b.helpers)
	appendSection(b.setup)
	appendSection(b.scoring)
	if strings.TrimSpace(b.ret) != "" {
		sections = append(sections, strings.TrimSpace(b.ret))
	}

	return strings.Join(sections, "\n\n")
}

func (b *ScriptBuilder) MustAddParams(params map[string]interface{}) *ScriptBuilder {
	if err := b.AddParams(params); err != nil {
		panic(fmt.Sprintf("unable to add script params: %s", err.Error()))
	}
	return b
}
