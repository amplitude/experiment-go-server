package evaluation

/*
#cgo darwin,amd64 CFLAGS: -I${SRCDIR}/lib/macosX64
#cgo darwin,amd64 LDFLAGS: -L${SRCDIR}/lib/macosX64 -Wl,-rpath ${SRCDIR}/lib/macosX64 -levaluation_interop

#cgo darwin,arm64 CFLAGS: -I${SRCDIR}/lib/macosArm64
#cgo darwin,arm64 LDFLAGS: -L${SRCDIR}/lib/macosArm64 -Wl,-rpath ${SRCDIR}/lib/macosArm64 -levaluation_interop

#cgo linux,amd64 CFLAGS: -I${SRCDIR}/lib/linuxX64
#cgo linux,amd64 LDFLAGS: -L${SRCDIR}/lib/linuxX64 -Wl,-rpath=${SRCDIR}/lib/linuxX64 -levaluation_interop

#cgo linux,arm64 CFLAGS: -I${SRCDIR}/lib/linuxArm64
#cgo linux,arm64 LDFLAGS: -L${SRCDIR}/lib/linuxArm64 -Wl,-rpath=${SRCDIR}/lib/linuxArm64 -levaluation_interop

#include "libevaluation_interop_api.h"
#include <stdlib.h>

typedef const char * (*evaluate) (const char * r, const char * u);
typedef void (*DisposeString) (const char* s);

const char * bridge_evaluate(evaluate f, const char * r, const char * u)
{
	return f(r, u);
}

void bridge_dispose(DisposeString f, const char * s)
{
	return f(s);
}
*/
import "C"
import "unsafe"

var lib = C.libevaluation_interop_symbols()
var root = lib.kotlin.root

func Evaluate(rules, user string) string {
	rulesCString := C.CString(rules)
	userCString := C.CString(user)
	resultCString := C.bridge_evaluate(root.evaluate, rulesCString, userCString)
	result := C.GoString(resultCString)
	C.bridge_dispose(lib.DisposeString, resultCString)
	C.free(unsafe.Pointer(rulesCString))
	C.free(unsafe.Pointer(userCString))
	return result
}
