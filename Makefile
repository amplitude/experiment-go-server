

xpmt: clean
	CGO_ENABLED=1 go build -o xpmt cmd/xpmt/main.go
#	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o cmd/xpmt/bin/amd64/xpmt cmd/xpmt/main.go
#	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -o cmd/xpmt/bin/arm64/xpmt cmd/xpmt/main.go

clean:
	rm -f xpmt

debug:
	# macosX64
	cp ../experiment-evaluation/evaluation-interop/build/bin/macosX64/debugShared/libevaluation_interop_api.h internal/evaluation/lib/macosX64/
	cp ../experiment-evaluation/evaluation-interop/build/bin/macosX64/debugShared/libevaluation_interop.dylib internal/evaluation/lib/macosX64/
	# macosArm64
	cp ../experiment-evaluation/evaluation-interop/build/bin/macosArm64/debugShared/libevaluation_interop_api.h internal/evaluation/lib/macosArm64/
	cp ../experiment-evaluation/evaluation-interop/build/bin/macosArm64/debugShared/libevaluation_interop.dylib internal/evaluation/lib/macosArm64/
	# linuxX64
	cp ../experiment-evaluation/evaluation-interop/build/bin/linuxX64/debugShared/libevaluation_interop_api.h internal/evaluation/lib/linuxX64/
	cp ../experiment-evaluation/evaluation-interop/build/bin/linuxX64/debugShared/libevaluation_interop.so internal/evaluation/lib/linuxX64/
	# linuxArm64
	cp ../experiment-evaluation/evaluation-interop/build/bin/linuxArm64/debugShared/libevaluation_interop_api.h internal/evaluation/lib/linuxArm64/
	cp ../experiment-evaluation/evaluation-interop/build/bin/linuxArm64/debugShared/libevaluation_interop.so internal/evaluation/lib/linuxArm64/

release:
	# macosX64
	cp ../experiment-evaluation/evaluation-interop/build/bin/macosX64/releaseShared/libevaluation_interop_api.h internal/evaluation/lib/macosX64/
	cp ../experiment-evaluation/evaluation-interop/build/bin/macosX64/releaseShared/libevaluation_interop.dylib internal/evaluation/lib/macosX64/
	# macosArm64
	cp ../experiment-evaluation/evaluation-interop/build/bin/macosArm64/releaseShared/libevaluation_interop_api.h internal/evaluation/lib/macosArm64/
	cp ../experiment-evaluation/evaluation-interop/build/bin/macosArm64/releaseShared/libevaluation_interop.dylib internal/evaluation/lib/macosArm64/
	# linuxX64
	cp ../experiment-evaluation/evaluation-interop/build/bin/linuxX64/releaseShared/libevaluation_interop_api.h internal/evaluation/lib/linuxX64/
	cp ../experiment-evaluation/evaluation-interop/build/bin/linuxX64/releaseShared/libevaluation_interop.so internal/evaluation/lib/linuxX64/
	# linuxArm64
	cp ../experiment-evaluation/evaluation-interop/build/bin/linuxArm64/releaseShared/libevaluation_interop_api.h internal/evaluation/lib/linuxArm64/
	cp ../experiment-evaluation/evaluation-interop/build/bin/linuxArm64/releaseShared/libevaluation_interop.so internal/evaluation/lib/linuxArm64/