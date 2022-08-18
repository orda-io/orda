package testonly

import (
	gocontext "context"
	"encoding/json"
	"github.com/orda-io/orda/client/pkg/context"
	"github.com/orda-io/orda/client/pkg/internal/datatypes"
	"github.com/orda-io/orda/client/pkg/model"
	"github.com/orda-io/orda/client/pkg/types"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// OperationsToString returns a string of an array of operations
func OperationsToString(ops []*model.Operation) string {
	sb := strings.Builder{}
	sb.WriteString("[ ")
	for i, op := range ops {
		sb.WriteString(op.ToString())
		if len(ops)-1 != i {
			sb.WriteString(", ")
		}
	}
	sb.WriteString(" ]")
	return sb.String()
}

// Marshal marshals an interface with checking error
func Marshal(t *testing.T, j interface{}) string {
	data, err := json.Marshal(j)
	require.NoError(t, err)
	return string(data)
}

// NewBase returns a new BaseDatatype
func NewBase(key string, t model.TypeOfDatatype) *datatypes.BaseDatatype {
	cm := &model.Client{
		CUID:       types.NewUID(),
		Alias:      "",
		Collection: "",
		SyncType:   0,
	}
	ctx := context.NewClientContext(gocontext.TODO(), cm)
	return datatypes.NewBaseDatatype(key, t, ctx, model.StateOfDatatype_DUE_TO_CREATE)
}
