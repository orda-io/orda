package context_test

import (
	gocontext "context"
	"github.com/orda-io/orda/client/pkg/context"
	"github.com/orda-io/orda/client/pkg/log"
	"testing"
)

func TestOrdaContext(t *testing.T) {
	t.Run("Can work with OrdaConext", func(t *testing.T) {
		log.Logger.Infof("print well?")
		ctx1 := context.NewOrdaContext(gocontext.TODO(), "🎪")
		ctx1.UpdateCollectionTags("Hello_Collection", 1)
		ctx1.UpdateClientTags("Hello_Client", "abcdefghijk")
		ctx1.UpdateDatatypeTags("Hello_Datatype", "123456789")
		ctx1.L().Infof("print well?")

		ctx2 := ctx1.CloneWithNewEmoji("👽")
		ctx2.L().Infof("clone print well?")
	})
}
