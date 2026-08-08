ifndef VERSION
$(error VERSION is required, for example: make VERSION=0.1.0)
endif

RELEASE_DIR := dist/v$(VERSION)

ARCHIVE_ARM64 := skytui_$(VERSION)_darwin_arm64.tar.gz
ARCHIVE_AMD64 := skytui_$(VERSION)_darwin_amd64.tar.gz

.PHONY: checksums
checksums:
	cd $(RELEASE_DIR) && shasum -a 256 $(ARCHIVE_ARM64) > checksums.txt && shasum -a 256 $(ARCHIVE_AMD64) >> checksums.txt
