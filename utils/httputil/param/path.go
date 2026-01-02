package param

import "strings"

type PathConfig struct {
	SplitPath string
	VarPrefix string
	VarSuffix string
}

func checkVar(part string, varPrefix, varSuffix string) bool {
	if varPrefix == "" && varSuffix == "" {
		return false
	}
	if varPrefix != "" && varSuffix != "" {
		return strings.HasPrefix(part, varPrefix) && strings.HasSuffix(part, varSuffix)
	}
	if varPrefix != "" {
		return strings.HasPrefix(part, varPrefix)
	}
	return strings.HasSuffix(part, varSuffix)
}
func trimVar(part string, varPrefix, varSuffix string) string {
	if varPrefix == "" && varSuffix == "" {
		return part
	}
	if varPrefix != "" && varSuffix != "" {
		part = strings.TrimPrefix(part, varPrefix)
		part = strings.TrimSuffix(part, varSuffix)
		return part
	}
	if varPrefix != "" {
		return strings.TrimPrefix(part, varPrefix)
	}
	return strings.TrimSuffix(part, varSuffix)
}
func getPathConfig(pc *PathConfig) *PathConfig {
	if pc == nil {
		pc = &PathConfig{}
	}
	if pc.SplitPath == "" {
		pc.SplitPath = "/"
	}
	if pc.VarPrefix == "" {
		pc.VarPrefix = ":"
	}
	return pc
}

func getPathList(requestPath, routePath string, splitStr string) (requestParts, routeParts []string) {
	requestPath = strings.TrimSpace(requestPath)
	routePath = strings.TrimSpace(routePath)

	requestParts = strings.Split(requestPath, splitStr)
	routeParts = strings.Split(routePath, splitStr)
	return requestParts, routeParts
}

func PathMatch(requestPath, routePath string, pcs ...*PathConfig) bool {
	var pc *PathConfig
	if len(pcs) > 0 {
		pc = pcs[0]
	}
	pc = getPathConfig(pc)

	requestParts, routeParts := getPathList(requestPath, routePath, pc.SplitPath)

	if len(requestParts) != len(routeParts) {
		return false
	}

	for i, part := range routeParts {
		if checkVar(part, pc.VarPrefix, pc.VarSuffix) {
			continue
		}
		if part != requestParts[i] {
			return false
		}
	}
	return true
}

func PathParam(requestPath, routePath string, pcs ...*PathConfig) map[string]string {
	var pc *PathConfig
	if len(pcs) > 0 {
		pc = pcs[0]
	}
	pc = getPathConfig(pc)

	requestParts, routeParts := getPathList(requestPath, routePath, pc.SplitPath)

	forLen := len(requestParts)
	if len(routeParts) < len(requestParts) {
		forLen = len(routeParts)
	}
	params := make(map[string]string)
	for i := 0; i < forLen; i++ {
		part := routeParts[i]
		if checkVar(part, pc.VarPrefix, pc.VarSuffix) {
			params[trimVar(part, pc.VarPrefix, pc.VarSuffix)] = requestParts[i]
		}
	}
	return params
}
