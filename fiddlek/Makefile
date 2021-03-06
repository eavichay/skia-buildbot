# The list of files we want to go into core.js, which is concat'd and
# minified. These files should be either present in the project, brought
# into third_party/bower_components via bower, or in node_modules.
CORE_SOURCE_FILES = node_modules/native-promise-only/npo.js \
  third_party/bower_components/webcomponentsjs/webcomponents-lite.min.js \
  ../res/js/common.js

BOWER_DIR=third_party/bower_components

VULCANIZE1=true

.PHONY: build
build: elements_html core_js
	go install -v ./go/fiddle
	go install -v ./go/fiddler
	go install -v ./go/fiddlecli

.PHONY: image
image: build
	CGO_ENABLED=0 GOOS=linux go install -a -installsuffix cgo ./go/fiddle
	docker pull gcr.io/skia-public/skia-release:prod
	SKIP_UPLOAD=1 ./build_fiddle_release
	SKIP_UPLOAD=1 ./build_fiddler_release

.PHONY: release
release: build
	CGO_ENABLED=0 GOOS=linux go install -a -installsuffix cgo ./go/fiddle
	docker pull gcr.io/skia-public/skia-release:prod
	./build_fiddle_release
	./build_fiddler_release

.PHONY: push
push:
	pushk fiddle fiddler

.PHONY: cli
cli:
	go install -v ./go/fiddlecli

.PHONY: testgo
testgo:
	go test -v ./...

.PHONY: run_local_fiddle
run_local_fiddle:
	fiddle --local --logtostderr --port=:8080

.PHONY: run_local_fiddler
run_local_fiddler:
	docker run -ti -p8000:8000  fiddler


# Build debug versions of core.js and elements.html.
.PHONY: debug
debug: clean_webtools debug_core_js debug_elements_html

include ../webtools/webtools.mk
