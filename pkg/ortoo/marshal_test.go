package ortoo

import (
	"encoding/json"
	"fmt"
	"github.com/knowhunger/ortoo/pkg/log"
	"github.com/stretchr/testify/require"
	"testing"
)

type HelloIface interface {
	json.Marshaler
	json.Unmarshaler
}

type Hello1 struct {
	json string
}

type Hello2 struct {
	json int
}

type World struct {
	Hellos map[string]HelloIface
	Size   int
}

type World2 struct {
	Hellos map[string]interface{}
	Size   int
}

func (w World) UnmarshalJSON(bytes []byte) error {
	fakeWorld := &struct {
		Hellos map[string]string
		Size   int
	}{}
	if err := json.Unmarshal(bytes, fakeWorld); err != nil {
		return err
	}
	w.Size = fakeWorld.Size
	// w.Hellos
	return nil
}

func (h Hello1) UnmarshalJSON(bytes []byte) error {
	fmt.Printf("unmarshal:%s", string(bytes))
	return json.Unmarshal(bytes, &h)
}

func (h Hello1) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("{ \"json\":\"%s\" }", h.json)), nil
}

func (h Hello2) UnmarshalJSON(bytes []byte) error {
	fmt.Printf("unmarshal:%s", string(bytes))
	return json.Unmarshal(bytes, &h)
}

func (h Hello2) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("{ \"json\":%d }", h.json)), nil
}

func TestMarshaling(t *testing.T) {
	t.Run("Can test marshalling and unmarshalling", func(t *testing.T) {
		w := &World{
			Hellos: make(map[string]HelloIface),
			Size:   0,
		}
		w.Hellos["a"] = &Hello1{
			json: "hello1",
		}
		w.Hellos["b"] = &Hello2{
			json: 1234,
		}
		j, err := json.Marshal(w)
		require.NoError(t, err)
		log.Logger.Infof("Marshaled: %v", string(j))

		m := &World2{}
		err = json.Unmarshal(j, m)
		require.NoError(t, err)
		log.Logger.Infof("Unmarshaled: %+v", m)
		// h:= Hello1{
		// 	json: "eeee",
		// }
		// j, err := json.Marshal(h)
		// require.NoError(t, err)
		// log.Logger.Infof("%v", string(j))
		//
		// m1 := make(map[string]string)
		// m1["a"]="b"
		// m1["b"]="c"
		// j2, err := json.Marshal(m1)
		// require.NoError(t, err)
		// log.Logger.Infof("%v", string(j2))
		//
		// m2 := make(map[string]HelloIface)
		// m2["a"]= h
		// m2["b"]= h
		// j3, err := json.Marshal(m2)
		// require.NoError(t, err)
		// log.Logger.Infof("%v", string(j3))
		//
		// m2_copy := make(map[string]HelloIface)
		//
		// err = json.Unmarshal(j3, &m2_copy)
		// require.NoError(t, err)
	})
}
