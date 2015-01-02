package spec

import (
	"strings"

	"github.com/casualjim/go-swagger/util"
)

// type operationRef struct {
// 	operation *Operation
// 	parameter *Parameter
// }

// specAnalyzer takes a swagger spec object and turns it into a registry
// with a bunch of utility methods to act on the information in the spec
type specAnalyzer struct {
	spec        *Swagger
	consumes    map[string]struct{}
	produces    map[string]struct{}
	authSchemes map[string]struct{}
	operations  map[string]map[string]*Operation
}

func (s *specAnalyzer) initialize() {
	for _, c := range s.spec.Consumes {
		s.consumes[c] = struct{}{}
	}
	for _, c := range s.spec.Produces {
		s.produces[c] = struct{}{}
	}
	for path, pathItem := range s.AllPaths() {
		s.analyzeOperations(path, &pathItem)
	}
}

func (s *specAnalyzer) analyzeOperations(path string, op *PathItem) {
	s.analyzeOperation("GET", path, op.Get)
	s.analyzeOperation("PUT", path, op.Put)
	s.analyzeOperation("POST", path, op.Post)
	s.analyzeOperation("PATCH", path, op.Patch)
	s.analyzeOperation("DELETE", path, op.Delete)
	s.analyzeOperation("HEAD", path, op.Head)
	s.analyzeOperation("OPTIONS", path, op.Options)
}

func (s *specAnalyzer) analyzeOperation(method, path string, op *Operation) {
	if op != nil {
		for _, c := range op.Consumes {
			s.consumes[c] = struct{}{}
		}
		for _, c := range op.Produces {
			s.produces[c] = struct{}{}
		}
		if _, ok := s.operations[method]; !ok {
			s.operations[method] = make(map[string]*Operation)
		}
		s.operations[method][path] = op
	}
}

// ConsumesFor gets the mediatypes for the operation
func (s *specAnalyzer) ConsumesFor(operation *Operation) []string {
	cons := make(map[string]struct{})
	for k := range s.consumes {
		cons[k] = struct{}{}
	}
	for _, c := range operation.Consumes {
		cons[c] = struct{}{}
	}
	return s.structMapKeys(cons)
}

// ProducesFor gets the mediatypes for the operation
func (s *specAnalyzer) ProducesFor(operation *Operation) []string {
	prod := make(map[string]struct{})
	for k := range s.produces {
		prod[k] = struct{}{}
	}
	for _, c := range operation.Produces {
		prod[c] = struct{}{}
	}
	return s.structMapKeys(prod)
}

func fieldNameFromParam(param *Parameter) string {
	if nm, ok := param.Extensions.GetString("go-name"); ok {
		return nm
	}
	return util.ToGoName(param.Name)
}

func (s *specAnalyzer) paramsAsMap(parameters []Parameter, res map[string]Parameter) {
	for _, param := range parameters {
		res[fieldNameFromParam(&param)] = param
	}
}

func (s *specAnalyzer) ParamsFor(method, path string) map[string]Parameter {
	res := make(map[string]Parameter)
	for _, param := range s.spec.Parameters {
		res[fieldNameFromParam(&param)] = param
	}
	if pi, ok := s.spec.Paths.Paths[path]; ok {
		s.paramsAsMap(pi.Parameters, res)
		s.paramsAsMap(s.operations[strings.ToUpper(method)][path].Parameters, res)
	}
	return res
}

func (s *specAnalyzer) OperationFor(method, path string) (*Operation, bool) {
	if mp, ok := s.operations[strings.ToUpper(method)]; ok {
		op, fn := mp[path]
		return op, fn
	}
	return nil, false
}

func (s *specAnalyzer) Operations() map[string]map[string]*Operation {
	return s.operations
}

func (s *specAnalyzer) structMapKeys(mp map[string]struct{}) []string {
	var result []string
	for k := range mp {
		result = append(result, k)
	}
	return result
}

// AllPaths returns all the paths in the swagger spec
func (s *specAnalyzer) AllPaths() map[string]PathItem {
	return s.spec.Paths.Paths
}

func (s *specAnalyzer) OperationIDs() []string {
	var result []string
	for _, v := range s.operations {
		for _, vv := range v {
			result = append(result, vv.ID)
		}
	}
	return result
}

func (s *specAnalyzer) RequiredConsumes() []string {
	return s.structMapKeys(s.consumes)
}

func (s *specAnalyzer) RequiredProduces() []string {
	return s.structMapKeys(s.produces)
}
