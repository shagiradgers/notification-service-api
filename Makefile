current_dir = $(shell pwd)

.build-proto-generator:
	go build github.com/shagiradgers/proto-generator

.generate-external-pb-deps:
	./proto-generator -url=https://raw.githubusercontent.com/shagiradgers/vk-notifications-api/master/api/vk/notifcations/api.proto \
		-out=pb/vk/vk.proto -out-gen=$(current_dir)
	./proto-generator -url=https://raw.githubusercontent.com/shagiradgers/telegram-notification-api/master/api/telegram_notification.proto \
		-out=pb/tg/telegram.proto -out-gen=$(current_dir)

.generate-pb:
	protoc --go_out=. --go_opt=paths=source_relative \
        --go-grpc_out=. --go-grpc_opt=paths=source_relative \
        api/notification.proto

.run:
	GOLANG_PROTOBUF_REGISTRATION_CONFLICT=warn  go run cmd/api.go

generate: .generate-pb .build-proto-generator .generate-external-pb-deps

run: .run