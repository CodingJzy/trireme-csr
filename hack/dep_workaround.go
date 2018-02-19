package hack

import (
	// ugly workaround to get dep to currently import the code-generator
	// will be addressed in: https://github.com/golang/dep/issues/1306
	_ "k8s.io/code-generator/pkg/util"
)
