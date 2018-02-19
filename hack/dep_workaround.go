package hack

import (
	// ugly workaround to get dep to currently import the code-generator
	//
	// will be addressed in: https://github.com/golang/dep/issues/1306
	//_ "k8s.io/code-generator/pkg/util"
	//
	// Another problem is: the code-generator has its own vendor folder,
	// which will never correctly work (gets not included during import),
	// we might need to switch to a different import method.
	//
	_ "k8s.io/code-generator/cmd/client-gen/generators"
	_ "k8s.io/gengo/examples/deepcopy-gen/generators"
	_ "k8s.io/gengo/examples/defaulter-gen/generators"
	_ "k8s.io/kube-openapi/pkg/generators"
)
