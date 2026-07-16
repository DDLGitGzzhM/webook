package main

func main() {
	//db := initDB()
	//rbd := initRedis()
	//u := initUser(db, rbd)
	//u.RegisterRoutes(server)

	server := InitWebServer()
	//server := gin.Default()
	//server.GET("/hello", func(ctx *gin.Context) {
	//	ctx.JSON(200, "你好 k8s")
	//})
	server.Run(":8080")
}
