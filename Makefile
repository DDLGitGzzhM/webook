.PHONY: mock
mock:
	$(shell go env GOPATH)/bin/mockgen -source=webook/internal/service/article.go -package=svcmocks -destination=webook/internal/service/mock/article.mock.go
	$(shell go env GOPATH)/bin/mockgen -source=webook/internal/repository/article/article.go -package=repomocks -destination=webook/internal/repository/mock/article.mock.go
	$(shell go env GOPATH)/bin/mockgen -source=webook/internal/repository/article/article_reader.go -package=repomocks -destination=webook/internal/repository/mock/article_reader.mock.go
	$(shell go env GOPATH)/bin/mockgen -source=webook/internal/repository/article/article_author.go -package=repomocks -destination=webook/internal/repository/mock/article_author.mock.go
