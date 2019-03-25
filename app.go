package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/singleflight"
)

var (
	numberOfRequest int64
	numberOfShared  int64
	group           singleflight.Group
)

// Service struct hold dependency
type Service struct {
	group singleflight.Group
}

func main() {
	var (
		r = gin.Default()
		s = Service{
			group: singleflight.Group{},
		}
	)

	gin.SetMode(gin.ReleaseMode)

	r.GET("/plain/:id", func(c *gin.Context) {
		id := c.Param("id")
		atomic.AddInt64(&numberOfRequest, 1)
		showIter(id)
	})

	r.GET("/outside/:id", func(c *gin.Context) {
		atomic.AddInt64(&numberOfRequest, 1)
		id := c.Param("id")
		v, _, shared := group.Do(id, func() (interface{}, error) {
			return showIter(id), nil
		})

		if shared {
			atomic.AddInt64(&numberOfShared, 1)
		}

		resp := v.(string)

		log.Println(resp)
	})

	r.GET("/stat", func(c *gin.Context) {

		stat := struct {
			NumberOfReq    int64
			NumberOfShared int64
		}{
			NumberOfReq:    atomic.LoadInt64(&numberOfRequest),
			NumberOfShared: atomic.LoadInt64(&numberOfShared),
		}

		log.Println("counter reset")

		numberOfRequest = 0
		numberOfShared = 0

		c.JSON(http.StatusOK, stat)
	})

	r.GET("/inside/:uid", func(c *gin.Context) {
		id := c.Param("uid")

		r := s.GroupInside(id)
		log.Println(r)
	})

	r.Run(":9090")

}

func showIter(id string) string {
	log.Println("handle request for id :", id)
	log.Println("iter: ", numberOfRequest)

	return fmt.Sprintf("handle request id : %s", id)
}

func (s *Service) GroupInside(id string) string {

	atomic.AddInt64(&numberOfRequest, 1)

	v, _, shared := s.group.Do(id, func() (interface{}, error) {
		return showIter(id), nil
	})

	res := v.(string)

	if shared {
		atomic.AddInt64(&numberOfShared, 1)
	}

	return res
}
