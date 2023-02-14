package dendrite

import (
	"net/http"

	"github.com/laminatedio/dendrite/internal/pkg/backend"
	"github.com/laminatedio/dendrite/internal/pkg/dendrite/dto"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

type Error struct {
	Message string `json:"message"`
}

type DendriteController struct {
	dendriteService *DendriteService
	logger          *zap.SugaredLogger
	config          *backend.Config
}

func NewDendriteController(dendriteService *DendriteService, logger *zap.SugaredLogger, config *backend.Config) *DendriteController {
	return &DendriteController{
		dendriteService: dendriteService,
		logger:          logger,
		config:          config,
	}
}

func (c *DendriteController) Query(ctx *gin.Context) {
	json := &dto.QueryInput{}
	err := ctx.BindJSON(json)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, Error{
			Message: "failed to parse body, please check whether the request body is valid",
		})
	} else {
		object, err := c.dendriteService.Query(ctx, json.Query)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, Error{
				Message: "failed to query: " + err.Error(),
			})
		} else {
			ctx.JSON(http.StatusOK, object)
		}
	}
}

func (c *DendriteController) GetCurrent(ctx *gin.Context) {
	json := &dto.GetCurrentInput{}
	err := ctx.BindJSON(json)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, Error{
			Message: "failed to parse body, please check whether the request body is valid",
		})
	} else {
		value, err := c.dendriteService.backend.GetCurrent(ctx, json.Path)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, Error{
				Message: err.Error(),
			})
		} else {
			ctx.JSON(http.StatusOK, map[string]string{
				"value": value,
			})
		}
	}
}

func (c *DendriteController) Get(ctx *gin.Context) {
	json := &dto.GetInput{}
	err := ctx.BindJSON(json)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, Error{
			Message: "failed to parse body, please check whether the request body is valid",
		})
	} else {
		value, err := c.dendriteService.backend.Get(ctx, json.Path, json.Version)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, Error{
				Message: err.Error(),
			})
		} else {
			ctx.JSON(http.StatusOK, map[string]string{
				"value": value,
			})
		}
	}
}

func (c *DendriteController) GetManyCurrent(ctx *gin.Context) {
	json := &dto.GetCurrentInput{}
	err := ctx.BindJSON(json)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, Error{
			Message: "failed to parse body, please check whether the request body is valid",
		})
	} else {
		values, err := c.dendriteService.backend.GetManyCurrent(ctx, json.Path)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, Error{
				Message: err.Error(),
			})
		} else {
			ctx.JSON(http.StatusOK, map[string][]string{
				"values": values,
			})
		}
	}
}

func (c *DendriteController) GetMany(ctx *gin.Context) {
	json := &dto.GetInput{}
	err := ctx.BindJSON(json)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, Error{
			Message: "failed to parse body, please check whether the request body is valid",
		})
	} else {
		values, err := c.dendriteService.backend.GetMany(ctx, json.Path, json.Version)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, Error{
				Message: err.Error(),
			})
		} else {
			ctx.JSON(http.StatusOK, map[string][]string{
				"values": values,
			})
		}
	}
}

func (c *DendriteController) Set(ctx *gin.Context) {
	json := &dto.SetInput{}
	err := ctx.BindJSON(json)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, Error{
			Message: "failed to parse body, please check whether the request body is valid",
		})
	} else {
		object, err := c.dendriteService.backend.Set(ctx, json.Path, json.Value, backend.SetOptions{
			KeepCurrent: json.KeepCurrent,
		})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, Error{
				Message: err.Error(),
			})
		} else {
			c.logger.Infof("(From %v) Created kv with path: %v with value: %v, backend: %v", ctx.ClientIP(), json.Path, json.Value, c.config.Type)
			ctx.JSON(http.StatusCreated, object)
		}
	}
}

func (c *DendriteController) SetMany(ctx *gin.Context) {
	json := &dto.SetManyInput{}
	err := ctx.BindJSON(json)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, Error{
			Message: "failed to parse body, please check whether the request body is valid",
		})
	} else {
		object, err := c.dendriteService.backend.SetMany(ctx, json.Path, json.Values, backend.SetOptions{
			KeepCurrent: json.KeepCurrent,
		})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, Error{
				Message: err.Error(),
			})
		} else {
			c.logger.Infof("(From %v) Created kv with path: %v with values: %v, backend: %v", ctx.ClientIP(), json.Path, json.Values, c.config.Type)
			ctx.JSON(http.StatusCreated, object)
		}
	}
}

func (c *DendriteController) RoutePattern() string {
	return "/"
}

func (c *DendriteController) RegisterControllerRoutes(rg *gin.RouterGroup) {
	rg.POST("/query", c.Query)
	rg.POST("/get", c.Get)
	rg.POST("/getMany", c.GetMany)
	rg.POST("/getCurrent", c.GetCurrent)
	rg.POST("/getManyCurrent", c.GetManyCurrent)
	rg.POST("/set", c.Set)
	rg.POST("/setMany", c.SetMany)
}
