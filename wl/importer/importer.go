package importer

import (
	"weblang/wl/types"
)

func Default() types.Importer {
	return &importer{}
}

type importer struct {

}

func (i *importer) Import(path string) (*types.Package, error) {
	return nil, nil
}