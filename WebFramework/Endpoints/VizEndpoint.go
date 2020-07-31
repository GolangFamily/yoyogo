package Endpoints

import (
	"github.com/yoyofx/yoyogo/WebFramework/Context"
	"github.com/yoyofx/yoyogo/WebFramework/Router"
	"html/template"
)

const (
	panicText = `<script src="http://mnur-prod-public.oss-cn-beijing.aliyuncs.com/0/tech/viz.js"></script>
<script src="http://mnur-prod-public.oss-cn-beijing.aliyuncs.com/0/tech/full.render.js"></script>
<script type="dot" id="dotscript">
{{.}}
</script>
<script>
  window.onload=function(e){
    var viz = new Viz();
    viz.renderSVGElement(document.getElementById('dotscript').innerText)
    .then(function(element) {
      document.body.appendChild(element);
    })
    .catch(error => {
      // Create a new Viz instance (@see Caveats page for more info)
      viz = new Viz();
      // Possibly display the error
      console.error(error);
    });
  }
</script>
`
)

var panicHTMLTemplate = template.Must(template.New("PanicPage").Parse(panicText))

func UseViz(router Router.IRouterBuilder) {
	router.GET("/actuator/graph", func(ctx *Context.HttpContext) {
		graphType := ctx.QueryStringOrDefault("type", "data")
		graphString := ctx.RequiredServices.GetGraph()

		if graphType == "data" {
			ctx.Text(200, graphString)
		} else {
			ctx.Header(Context.HeaderContentType, Context.MIMETextHTMLCharsetUTF8)
			ctx.Response.WriteHeader(200)
			_ = panicHTMLTemplate.Execute(ctx.Response.ResponseWriter, template.HTML(graphString))
		}
	})
}