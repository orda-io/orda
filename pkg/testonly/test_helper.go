package testonly

import (
	gocontext "context"
	"encoding/json"
	"github.com/knowhunger/ortoo/pkg/context"
	"github.com/knowhunger/ortoo/pkg/internal/datatypes"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/pkg/types"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
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

func Marshal(t *testing.T, j interface{}) string {
	data, err := json.Marshal(j)
	require.NoError(t, err)
	return string(data)
}

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
