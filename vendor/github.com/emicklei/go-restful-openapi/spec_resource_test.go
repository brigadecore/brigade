package restfulspec

import (
	"testing"

	restful "github.com/emicklei/go-restful"
)

func TestBuildSwagger(t *testing.T) {
	path := "/testPath"

	ws1 := new(restful.WebService)
	ws1.Path(path)
	ws1.Route(ws1.GET("").To(dummy))

	ws2 := new(restful.WebService)
	ws2.Path(path)
	ws2.Route(ws2.DELETE("").To(dummy))

	c := Config{}
	c.WebServices = []*restful.WebService{ws1, ws2}
	s := BuildSwagger(c)

	if !(s.Paths.Paths[path].Get != nil && s.Paths.Paths[path].Delete != nil) {
		t.Errorf("Swagger spec should have methods for GET and DELETE")
	}

}
