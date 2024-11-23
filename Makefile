APP=gpsreader

ifeq ($(shell uname), Darwin)
$(info Building on macOS)
BUILD_ENV = CGO_ENABLED=1 GOARCH=arm64
endif

$(APP):  $(APP).go
	$(BUILD_ENV) go build

install:	$(APP)
	sudo cp gpsreader /usr/local/bin
	sudo cp systemd/gpsreader.service /etc/systemd/system
	sudo systemctl daemon-reload
	sudo systemctl enable gpsreader

clean:
	rm -f $(APP)

.PHONY: clean
