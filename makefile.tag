# Makefile for automatic semantic versioning and git tagging
# Usage:
#   make tag: increments patch version and tags (default)
#   make tag TYPE=minor: increments minor version and tags
#   make tag TYPE=major: increments major version and tags

# default version type
TYPE ?= patch

# get version from latest git tag, default to 0.0.0 if no tags exist
CURRENT_VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "0.0.0")

# extract major, minor, patch numbers
MAJOR := $(word 1, $(subst ., ,$(CURRENT_VERSION)))
MINOR := $(word 2, $(subst ., ,$(CURRENT_VERSION)))
PATCH := $(word 3, $(subst ., ,$(CURRENT_VERSION)))

# calculate new version based on type
NEW_VERSION = $(shell \
	if [ "$(TYPE)" = "major" ]; then \
		echo $$(( $(MAJOR) + 1 )).0.0; \
	elif [ "$(TYPE)" = "minor" ]; then \
		echo $(MAJOR).$$(( $(MINOR) + 1 )).0; \
	else \
		echo $(MAJOR).$(MINOR).$$(( $(PATCH) + 1 )); \
	fi)

tag:
	@echo "current version: $(CURRENT_VERSION)"
	@echo "new version: $(NEW_VERSION)"
	@git tag -a $(NEW_VERSION) -m "version $(NEW_VERSION)"
	@git push --tags
	@echo "tagged and pushed $(NEW_VERSION)"


untag:
	@if [ "$(CURRENT_VERSION)" = "0.0.0" ]; then \
		echo "No tags to delete"; \
	else \
		echo "deleting latest tag: $(CURRENT_VERSION)"; \
		git tag -d $(CURRENT_VERSION); \
		git push origin :refs/tags/$(CURRENT_VERSION); \
		echo "latest tag deleted"; \
	fi

.PHONY: tag untag
