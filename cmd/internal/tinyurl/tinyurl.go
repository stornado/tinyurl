package tinyurl

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	cuckoo "github.com/seiflotfy/cuckoofilter"
	"github.com/skip2/go-qrcode"
	"github.com/spaolacci/murmur3"
	"github.com/spf13/viper"
)

func Shorten(origin string) (string, error) {
	rawURL, err := url.Parse(origin)
	if err != nil {
		return "", err
	}
	mur32 := murmur3.New32()
	mur32.Write([]byte(rawURL.String()))
	b58 := Base58{big.NewInt(int64(mur32.Sum32()))}
	return b58.Encode(), nil
}

func ShortenURL(origin, baseurl string) (*url.URL, error) {
	if baseurl == "" {
		baseurl = viper.GetString("baseurl")
	}

	baseURL, err := url.Parse(baseurl)
	if err != nil {
		return nil, err
	}

	short, err := Shorten(origin)
	if err != nil {
		return nil, err
	}
	baseURL.Path = path.Join(baseURL.Path, short)
	return baseURL, nil
}

func QrCode(text string) ([]byte, error) {
	return qrcode.Encode(text, qrcode.Highest, 256)
}

func SaveQrCode(text, filename string) error {
	return qrcode.WriteFile(text, qrcode.Highest, 256, filename)
}

func Serve(port int) {
	cf := cuckoo.NewFilter(math.MaxInt16)
	go func(cf *cuckoo.Filter) {
		shortUrls, err := ListAll()
		if err != nil {
			panic(err)
		}
		for _, shortUrl := range shortUrls {
			ok := cf.InsertUnique([]byte(shortUrl.Short))
			fmt.Println(ok, shortUrl)
		}
	}(cf)

	if port <= 0 {
		port = viper.GetInt("port")
	}
	r := gin.Default()

	r.Any("/:short", func(c *gin.Context) {
		short := c.Params.ByName("short")
		found := cf.Lookup([]byte(short))
		if found {
			origin, err := CacheGet(short)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"err": err.Error()})
				return
			}
			if origin != "" {
				c.Redirect(http.StatusTemporaryRedirect, origin)
				return
			}

			tiny, err := PersistGet(short)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"err": err.Error()})
				return
			}
			c.Redirect(http.StatusTemporaryRedirect, tiny.Origin)
		} else {
			c.JSON(http.StatusNotFound, gin.H{"short": short})
		}
	})

	r.POST("/shorturl/", func(c *gin.Context) {
		var data struct {
			Origin string `json:"origin" binding:"required"`
		}
		err := c.ShouldBindJSON(&data)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"err": err.Error()})
			return
		}
		short, err := Shorten(data.Origin)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"origin": data.Origin, "err": err.Error()})
			return
		}
		if cf.Lookup([]byte(short)) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"err": short + " already exists"})
			return
		}

		cf.InsertUnique([]byte(short))
		tiny := Tinyurl{Short: short, Origin: data.Origin}
		_, err = tiny.Add()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
			return
		}
		err = Cache(tiny.Short, tiny.Origin)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"origin": data.Origin, "short": short, "base": viper.GetString("baseurl")})
	})

	r.PUT("/shorturl/:short", func(c *gin.Context) {
		var data struct {
			Origin string `json:"origin" binding:"required"`
			Base   string `json:"base" binding:"required"`
		}
		err := c.ShouldBindJSON(&data)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"err": err.Error()})
			return
		}
		short, err := Shorten(data.Origin)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"origin": data.Origin, "err": err.Error()})
			return
		}

		if !cf.Lookup([]byte(short)) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"err": "not found"})
			return
		}
		if short != c.Params.ByName("short") {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"err": "short cannot modified"})
			return
		}
		baseURL, err := url.Parse(data.Base)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"base": data.Base, "err": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"origin": data.Origin, "short": path.Join(baseURL.String(), short), "base": baseURL.String()})
		cf.Insert([]byte(short))
	})

	r.GET("/shorturl/:short", func(c *gin.Context) {
		short := c.Params.ByName("short")
		if !cf.Lookup([]byte(short)) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"err": "not found"})
			return
		}
		if short != c.Params.ByName("short") {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"err": "short cannot modified"})
			return
		}
		baseURL, err := url.Parse(viper.GetString("baseurl"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"base": baseURL.String(), "err": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"short": path.Join(baseURL.String(), short), "base": baseURL.String()})
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
