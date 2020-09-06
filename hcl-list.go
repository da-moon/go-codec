package codec

import (
	"fmt"
	"sort"

	multierror "github.com/hashicorp/go-multierror"
	ast "github.com/hashicorp/hcl/hcl/ast"
)

// ListParser is the interface used for
// a slice of objects in HCL format
type ListParser interface {
	Empty() ObjectParser
	Append(v interface{}) error
}

// ObjectParser is the interface that represents
// an object in HCL format
type ObjectParser interface {
	GetID(parent ...string) string
	ParseAST(raw *ast.ObjectItem) error
}

func ParseList(lists []*ast.ObjectList, key string, parser ListParser) (err error) {
	var res1 []ObjectParser
	err = &multierror.Error{}

	for _, list := range lists {
		for _, obj := range list.Filter(key).Items {
			objParser := parser.Empty()
			if parseErr := objParser.ParseAST(obj); parseErr != nil {
				multierror.Append(err, parseErr)
				continue
			}
			res1 = append(res1, objParser)
		}
	}
	sort.Slice(res1, func(i, j int) bool {
		return res1[i].GetID() < res1[j].GetID()
	})
	var lastId = ""
	for _, obj := range res1 {
		if id := obj.GetID(); lastId != id {
			lastId = id
			if appendErr := parser.Append(obj); appendErr != nil {
				multierror.Append(err, appendErr)
			}
		} else {
			multierror.Append(err, fmt.Errorf(`%s with %s already defined`, key, id))
		}
	}
	err = err.(*multierror.Error).ErrorOrNil()
	return
}
