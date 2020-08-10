package main

import (
	"fmt"
	"github.com/yoyofx/yoyogo/Abstractions"
	"github.com/yoyofx/yoyogo/DependencyInjection"
	"github.com/yoyofx/yoyogo/Examples/SimpleWeb/contollers"
	"github.com/yoyofx/yoyogo/Examples/SimpleWeb/models"
	"github.com/yoyofx/yoyogo/WebFramework"
	"github.com/yoyofx/yoyogo/WebFramework/Context"
	"github.com/yoyofx/yoyogo/WebFramework/Endpoints"
	"github.com/yoyofx/yoyogo/WebFramework/Mvc"
	"github.com/yoyofx/yoyogo/WebFramework/Router"
)

func main() {

	//webHost := YoyoGo.CreateDefaultBuilder(registerEndpointRouterConfig).Build()
	webHost := CreateCustomBuilder().Build()
	webHost.Run()
}

//* Create the builder of Web host
func CreateCustomBuilder() *Abstractions.HostBuilder {
	return YoyoGo.NewWebHostBuilder().
		SetEnvironment(Context.Prod).
		UseFastHttp().
		Configure(func(app *YoyoGo.WebApplicationBuilder) {
			app.UseStatic("Static")
			app.UseEndpoints(registerEndpointRouterConfig)
			app.UseMvc(func(builder *Mvc.ControllerBuilder) {
				builder.AddController(contollers.NewUserController)
				builder.AddFilter("/v1/user/info", &contollers.TestActionFilter{})
			})
		}).
		ConfigureServices(func(serviceCollection *DependencyInjection.ServiceCollection) {
			serviceCollection.AddTransientByImplements(models.NewUserAction, new(models.IUserAction))
		}).
		OnApplicationLifeEvent(getApplicationLifeEvent)
}

//*/

//region router config function
func registerEndpointRouterConfig(router Router.IRouterBuilder) {
	Endpoints.UseHealth(router)
	Endpoints.UseViz(router)
	//Endpoints.UsePprof(router)

	router.GET("/error", func(ctx *Context.HttpContext) {
		panic("http get error")
	})

	router.POST("/info/:id", PostInfo)

	router.Group("/v1/api", func(router *Router.RouterGroup) {
		router.GET("/info", GetInfo)
	})

	router.GET("/info", GetInfo)
	router.GET("/ioc", GetInfoByIOC)
}

//endregion

//region Http Request Methods

type UserInfo struct {
	UserName string `param:"username"`
	Number   string `param:"q1"`
	Id       string `param:"id"`
}

//HttpGet request: /info  or /v1/api/info
//bind UserInfo for id,q1,username
func GetInfo(ctx *Context.HttpContext) {
	ctx.JSON(200, Context.M{"info": "ok"})
}

func GetInfoByIOC(ctx *Context.HttpContext) {
	var userAction models.IUserAction
	_ = ctx.RequiredServices.GetService(&userAction)
	ctx.JSON(200, Context.M{"info": "ok " + userAction.Login("zhang")})
}

//HttpPost request: /info/:id ?q1=abc&username=123
func PostInfo(ctx *Context.HttpContext) {
	qs_q1 := ctx.Input.Query("q1")
	pd_name := ctx.Input.Param("username")
	id := ctx.Input.Param("id")
	userInfo := &UserInfo{}
	_ = ctx.Bind(userInfo)

	strResult := fmt.Sprintf("Name:%s , Q1:%s , bind: %s , routeData id:%s", pd_name, qs_q1, userInfo, id)

	ctx.JSON(200, Context.M{"info": "hello world", "result": strResult})
}

func getApplicationLifeEvent(life *Abstractions.ApplicationLife) {
	printDataEvent := func(event Abstractions.ApplicationEvent) {
		fmt.Printf("[yoyogo] Topic: %s; Event: %v\n", event.Topic, event.Data)
	}

	for {
		select {
		case ev := <-life.ApplicationStarted:
			go printDataEvent(ev)
		case ev := <-life.ApplicationStopped:
			go printDataEvent(ev)
			break
		}
	}
}

//endregion
