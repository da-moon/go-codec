package codec

import (
	"bytes"
	"io"

	multierror "github.com/hashicorp/go-multierror"
	hcl "github.com/hashicorp/hcl"
	ast "github.com/hashicorp/hcl/hcl/ast"
	stacktrace "github.com/palantir/stacktrace"
)

// CheckHCLKeys checks whether the keys in the AST list contains any of the valid keys provided.
func CheckHCLKeys(node ast.Node, valid []string) error {

	var list *ast.ObjectList
	switch n := node.(type) {
	case *ast.ObjectList:
		list = n
	case *ast.ObjectType:
		list = n.List
	default:
		return stacktrace.NewError("cannot check HCL keys of type %T", n)
	}

	validMap := make(map[string]struct{}, len(valid))
	for _, v := range valid {
		validMap[v] = struct{}{}
	}

	var result error
	for _, item := range list.Items {
		key := item.Keys[0].Token.Value().(string)
		if _, ok := validMap[key]; !ok {
			result = multierror.Append(result, stacktrace.NewError("invalid key %q on line %d", key, item.Assign.Line))
		}
	}

	return result
}

func ParseHCLMerge(readers ...io.Reader) (roots *ast.ObjectList, err error) {
	err = &multierror.Error{}
	var err1 error
	contents := &bytes.Buffer{}
	var ok bool
	for _, reader := range readers {
		var buf bytes.Buffer
		if _, err1 = io.Copy(&buf, reader); err1 != nil {
			err = multierror.Append(err, err1)
			continue
		}
		var rawRoot *ast.File
		if rawRoot, err1 = hcl.Parse(buf.String()); err1 != nil {
			err = multierror.Append(err, err1)
			continue
		}
		_, ok := rawRoot.Node.(*ast.ObjectList)
		if !ok {
			err = multierror.Append(err, stacktrace.NewError("error parsing: root should be an object"))
			continue
		}
		contents.Write(buf.Bytes())
		contents.WriteString("\n")
	}
	var rawRoot *ast.File
	if rawRoot, err1 = hcl.Parse(contents.String()); err1 != nil {
		err = multierror.Append(err, err1)
		return
	}
	roots, ok = rawRoot.Node.(*ast.ObjectList)
	if !ok {
		err = multierror.Append(err, stacktrace.NewError("error parsing: root should be an object"))
	}
	err = err.(*multierror.Error).ErrorOrNil()
	return
}

func ParseHCL(readers ...io.Reader) (lists []*ast.ObjectList, err error) {
	err = &multierror.Error{}
	var err1 error
	var ok bool
	for _, reader := range readers {
		var buf bytes.Buffer
		if _, err1 = io.Copy(&buf, reader); err1 != nil {
			err = multierror.Append(err, err1)
			continue
		}
		var rawRoot *ast.File
		if rawRoot, err1 = hcl.Parse(buf.String()); err1 != nil {
			err = multierror.Append(err, err1)
			continue
		}
		var list *ast.ObjectList
		list, ok = rawRoot.Node.(*ast.ObjectList)
		if !ok {
			err = multierror.Append(err, stacktrace.NewError("error parsing: root should be an object"))
			continue
		}
		lists = append(lists, list)
	}
	err = err.(*multierror.Error).ErrorOrNil()
	return
}
