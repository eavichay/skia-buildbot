# The list of files we want to go into core.js, which is concat'd and
# minified. These files should be either present in the project, brought
# into third_party/bower_compoents via bower, or in node_modules.
CORE_SOURCE_FILES = node_modules/native-promise-only/npo.js \
          third_party/bower_components/webcomponentsjs/webcomponents.min.js \
          ../res/js/common.js \
          res/js/ctfe.js

BOWER_DIR=third_party/bower_components

VULCANIZE1=true

include ../go/skiaversion/skiaversion.mk

# Require clean_webtools to ensure changes to res/common/imp/*.html are reflected.
.PHONY: ctfe
ctfe: res/css/ctfe.css clean_webtools core_js elements_html skiaversion
	go install -v ./go/ctfe

.PHONY: ctfe_debug
ctfe_debug: res/css/ctfe.css debug skiaversion
	go install -v ./go/ctfe

.PHONY: ctfe_release
ctfe_release: ctfe
	./build_ctfe_release "$(MESSAGE)"

.PHONY: ctmaster_release
ctmaster_release: all
	./build_ctmaster_release "$(MESSAGE)"

# Build debug versions of core.js and elements.html.
.PHONY: debug
debug: clean_webtools debug_core_js debug_elements_html

include ../webtools/webtools.mk

# Master scripts.
.PHONY: build_chromium
build_chromium:
	go install -v ./go/master_scripts/build_chromium/

.PHONY: create_pagesets_on_workers
create_pagesets_on_workers:
	go install -v ./go/master_scripts/create_pagesets_on_workers/

.PHONY: capture_archives_on_workers
capture_archives_on_workers:
	go install -v ./go/master_scripts/capture_archives_on_workers/

.PHONY: run_lua_on_workers
run_lua_on_workers:
	go install -v ./go/master_scripts/run_lua_on_workers/

.PHONY: capture_skps_on_workers
capture_skps_on_workers:
	go install -v ./go/master_scripts/capture_skps_on_workers/

.PHONY: pixel_diff_on_workers
pixel_diff_on_workers:
	go install -v ./go/master_scripts/pixel_diff_on_workers/

.PHONY: run_chromium_perf_on_workers
run_chromium_perf_on_workers:
	go install -v ./go/master_scripts/run_chromium_perf_on_workers/

.PHONY: run_chromium_analysis_on_workers
run_chromium_analysis_on_workers:
	go install -v ./go/master_scripts/run_chromium_analysis_on_workers/

.PHONY: metrics_analysis_on_workers
metrics_analysis_on_workers:
	go install -v ./go/master_scripts/metrics_analysis_on_workers/

# Worker scripts.
.PHONY: create_pagesets
create_pagesets:
	go install -v ./go/worker_scripts/create_pagesets/

.PHONY: build_repo
build_repo:
	go install -v ./go/worker_scripts/build_repo/

.PHONY: isolate_telemetry
isolate_telemetry:
	go install -v ./go/worker_scripts/isolate_telemetry/

.PHONY: capture_archives
capture_archives:
	go install -v ./go/worker_scripts/capture_archives/

.PHONY: run_lua
run_lua:
	go install -v ./go/worker_scripts/run_lua/

.PHONY: capture_skps
capture_skps:
	go install -v ./go/worker_scripts/capture_skps/

.PHONY: pixel_diff
pixel_diff:
	go install -v ./go/worker_scripts/pixel_diff/

.PHONY: run_chromium_perf
run_chromium_perf:
	go install -v ./go/worker_scripts/run_chromium_perf/

.PHONY: run_chromium_analysis
run_chromium_analysis:
	go install -v ./go/worker_scripts/run_chromium_analysis/

.PHONY: metrics_analysis
metrics_analysis:
	go install -v ./go/worker_scripts/metrics_analysis/

.PHONY: poller
poller:
	go install -v ./go/poller/

# This is the quick version triggered by the presubmit script.
.PHONY: testgo
testgo:
	go test ./go/... -v -test.short

.PHONY: test
test:
	go test ./go/...

.PHONY: master_scripts
master_scripts: build_chromium \
	create_pagesets_on_workers capture_archives_on_workers \
	run_lua_on_workers capture_skps_on_workers \
        pixel_diff_on_workers run_chromium_perf_on_workers poller \
        run_chromium_analysis_on_workers metrics_analysis_on_workers

.PHONY: worker_scripts
worker_scripts: create_pagesets capture_archives run_lua capture_skps \
                pixel_diff run_chromium_perf \
                run_chromium_analysis build_repo metrics_analysis \
                isolate_telemetry

.PHONY: all
all: master_scripts worker_scripts testgo

.PHONY: e2e_tests
e2e_tests:
	@echo "Running create_pagesets"
	go run go/worker_scripts/create_pagesets/main.go -logtostderr --worker_num=1 --pageset_type=Dummy1k
	@echo "Completed create_pagests. Sleeping for 15 seconds and continuing."
	sleep 15
	@echo
	@echo

	@echo "Running capture_archives"
	go run go/worker_scripts/capture_archives/main.go -logtostderr --worker_num=1 --pageset_type=Dummy1k --chromium_build=38517cc-d7e49ce -logtostderr
	@echo "Completed capture_archives. Sleeping for 15 seconds and continuing."
	sleep 15
	@echo
	@echo

	@echo "Running skpicture_printer"
	go run go/worker_scripts/capture_skps/main.go --worker_num=1 --pageset_type=Dummy1k --chromium_build=38517cc-d7e49ce -logtostderr --run_id=rmistry123 --target_platform=Linux
	@echo "Completed skpicture_printer. Sleeping for 15 seconds and continuing."
	sleep 15
	@echo
	@echo

	@echo "Running smoothness"
	go run go/worker_scripts/run_benchmark/main.go --worker_num=1 --pageset_type=Dummy1k --chromium_build=38517cc-d7e49ce -logtostderr --run_id=rmistry123 --benchmark_name=smoothness --repeat_benchmark=2 --target_platform=Linux
	@echo "Completed smoothness. Sleeping for 15 seconds and continuing."
	sleep 15
	@echo
	@echo

	@echo "Running rasterize_and_record_micro"
	go run go/worker_scripts/run_benchmark/main.go --worker_num=1 --pageset_type=Dummy1k --chromium_build=38517cc-d7e49ce -logtostderr --run_id=rmistry123 --benchmark_name=rasterize_and_record_micro --repeat_benchmark=2 --target_platform=Linux
	@echo "Completed rasterize_and_record_micro. Sleeping for 15 seconds and continuing."
	sleep 15
	@echo
	@echo

	@echo "Running repaint"
	go run go/worker_scripts/run_benchmark/main.go --worker_num=1 --pageset_type=Dummy1k --chromium_build=38517cc-d7e49ce -logtostderr --run_id=rmistry123 --benchmark_name=repaint --repeat_benchmark=2 --target_platform=Linux
	@echo "Completed repaint. Sleeping for 15 seconds and continuing."
	sleep 15
	@echo
	@echo

	@echo "Running lua scripts"
	go run go/worker_scripts/run_lua/main.go --worker_num=1 --pageset_type=Dummy1k --chromium_build=38517cc-d7e49ce --run_id=rmistry123 --logtostderr
	@echo "Completed lua scripts. Sleeping for 15 seconds and continuing."
	sleep 15
	@echo
	@echo
