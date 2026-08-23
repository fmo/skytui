ifndef VERSION
$(error VERSION is required, for example: make VERSION=0.1.0)
endif

RELEASE_DIR := dist/v$(VERSION)

ARCHIVES := \
	skytui_$(VERSION)_darwin_arm64.tar.gz \
	skytui_$(VERSION)_darwin_amd64.tar.gz \
	skytui_$(VERSION)_linux_arm64.tar.gz \
	skytui_$(VERSION)_linux_amd64.tar.gz \
	skytui_$(VERSION)_windows_amd64.zip

.PHONY: release archives checksums
release: checksums

archives:
	mkdir -p $(RELEASE_DIR)/darwin_arm64 $(RELEASE_DIR)/darwin_amd64 $(RELEASE_DIR)/linux_arm64 $(RELEASE_DIR)/linux_amd64 $(RELEASE_DIR)/windows_amd64
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o $(RELEASE_DIR)/darwin_arm64/skytui
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o $(RELEASE_DIR)/darwin_amd64/skytui
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o $(RELEASE_DIR)/linux_arm64/skytui
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o $(RELEASE_DIR)/linux_amd64/skytui
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o $(RELEASE_DIR)/windows_amd64/skytui.exe
	tar -C $(RELEASE_DIR)/darwin_arm64 -czf $(RELEASE_DIR)/skytui_$(VERSION)_darwin_arm64.tar.gz skytui
	tar -C $(RELEASE_DIR)/darwin_amd64 -czf $(RELEASE_DIR)/skytui_$(VERSION)_darwin_amd64.tar.gz skytui
	tar -C $(RELEASE_DIR)/linux_arm64 -czf $(RELEASE_DIR)/skytui_$(VERSION)_linux_arm64.tar.gz skytui
	tar -C $(RELEASE_DIR)/linux_amd64 -czf $(RELEASE_DIR)/skytui_$(VERSION)_linux_amd64.tar.gz skytui
	cd $(RELEASE_DIR)/windows_amd64 && zip -q ../skytui_$(VERSION)_windows_amd64.zip skytui.exe

checksums: archives
	cd $(RELEASE_DIR) && shasum -a 256 $(ARCHIVES) > checksums.txt
