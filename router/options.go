package router

type Option func(r *chiWrapper)

func OptEnableCORS(r *chiWrapper) {
	r.enableCORS = true
}

func OptHealthCheck(path string, checker APIHandler) Option {
	return func(r *chiWrapper) {
		r.healthEndpoint = path
		r.healthHandler = checker
	}
}

func OptVersionHeader(svcName, version string) Option {
	return func(r *chiWrapper) {
		if version == "" {
			version = "unknown"
		}
		r.version = version
		r.svcName = svcName
	}
}

func OptEnableTracing(svcName string) Option {
	return func(r *chiWrapper) {
		r.svcName = svcName
		r.enableTracing = true
	}
}

func OptRecoverer() Option {
	return func(r *chiWrapper) {
		r.enableRecover = true
	}
}
