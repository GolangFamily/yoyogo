package contollers

import (
	"github.com/yoyofx/yoyogo/Examples/SimpleWeb/models"
	"github.com/yoyofx/yoyogo/WebFramework/ActionResult"
	"github.com/yoyofx/yoyogo/WebFramework/Context"
	"github.com/yoyofx/yoyogo/WebFramework/Mvc"
)

type UserController struct {
	Mvc.ApiController
	userAction models.IUserAction
}

func NewUserController(userAction models.IUserAction) *UserController {
	return &UserController{userAction: userAction}
}

type RegisterRequest struct {
	Mvc.RequestBody
	UserName string `param:"UserName"`
	Password string `param:"Password"`
}

func (controller UserController) Register(ctx *Context.HttpContext, request *RegisterRequest) ActionResult.IActionResult {
	result := Mvc.ApiResult{Success: true, Message: "ok", Data: request}

	return ActionResult.Json{Data: result}
}

func (controller UserController) GetUserName(ctx *Context.HttpContext, request *RegisterRequest) ActionResult.IActionResult {
	result := Mvc.ApiResult{Success: true, Message: "ok", Data: request}

	return ActionResult.Json{Data: result}
}

func (controller UserController) PostUserInfo(request *RegisterRequest) ActionResult.IActionResult {

	return ActionResult.Json{Data: Mvc.ApiResult{Success: true, Message: "ok", Data: request}}
}

func (controller UserController) GetInfo() Mvc.ApiResult {

	return controller.OK(controller.userAction.Login("zhang"))
}
