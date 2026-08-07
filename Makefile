.PHONY: mock
mock:
	$(shell go env GOPATH)/bin/mockgen -source=webook/internal/service/article.go -package=svcmocks -destination=webook/internal/service/mock/article.mock.go
	$(shell go env GOPATH)/bin/mockgen -source=webook/internal/repository/article/article.go -package=repomocks -destination=webook/internal/repository/mock/article.mock.go
	$(shell go env GOPATH)/bin/mockgen -source=webook/internal/repository/article/article_reader.go -package=repomocks -destination=webook/internal/repository/mock/article_reader.mock.go
	$(shell go env GOPATH)/bin/mockgen -source=webook/internal/repository/article/article_author.go -package=repomocks -destination=webook/internal/repository/mock/article_author.mock.go

.PHONY: grpc
grpc:
	@protoc --go_out=webook/api/proto/gen --go_opt=paths=source_relative \
		--go-grpc_out=webook/api/proto/gen --go-grpc_opt=paths=source_relative \
		-I webook/api/proto \
		-I /opt/homebrew/include \
		webook/api/proto/intr/v1/intr.proto \
		webook/api/proto/account/v1/account.proto \
		webook/api/proto/payment/v1/payment.proto \
		webook/api/proto/reward/v1/reward.proto \
		webook/api/proto/comment/v1/comment.proto \
		webook/api/proto/follow/v1/follow.proto \
		webook/api/proto/feed/v1/feed.proto \
		webook/api/proto/search/v1/sync.proto \
		webook/api/proto/search/v1/search.proto \
		webook/api/proto/tag/v1/tag.proto
