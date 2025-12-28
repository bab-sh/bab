package babfile

import "github.com/invopop/jsonschema"

type RunItem interface {
	isRunItem()
	ShouldRunOnPlatform(platform string) bool
	GetWhen() string
}

func RunItemSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		OneOf: []*jsonschema.Schema{
			CommandRunSchema(),
			TaskRunSchema(),
			LogRunSchema(),
			PromptRunSchema(),
		},
	}
}

func matchesPlatform(platforms []Platform, platform string) bool {
	if len(platforms) == 0 {
		return true
	}
	for _, p := range platforms {
		if string(p) == platform {
			return true
		}
	}
	return false
}
