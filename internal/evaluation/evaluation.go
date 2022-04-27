package evaluation

/*
#cgo darwin,amd64 CFLAGS: -I${SRCDIR}/lib/macosX64
#cgo darwin,amd64 LDFLAGS: -L${SRCDIR}/lib/macosX64 -levaluation_interop

#cgo darwin,arm64 CFLAGS: -I${SRCDIR}/lib/macosArm64
#cgo darwin,arm64 LDFLAGS: -L${SRCDIR}/lib/macosArm64 -levaluation_interop

#cgo linux,amd64 CFLAGS: -I${SRCDIR}/lib/linuxX64
#cgo linux,amd64 LDFLAGS: -L${SRCDIR}/lib/linuxX64 -levaluation_interop

#cgo linux,arm64 CFLAGS: -I${SRCDIR}/lib/linuxArm64
#cgo linux,arm64 LDFLAGS: -L${SRCDIR}/lib/linuxArm64 -levaluation_interop

#include "libevaluation_interop_api.h"

typedef const char * (*evaluate) (const char * r, const char * u);

const char * bridge_evaluate(evaluate f, const char * r, const char * u)
{
	return f(r, u);
}
*/
import "C"

func Evaluate(rules, user string) string {
	rulesCString := C.CString(rules)
	userCString := C.CString(user)
	root := C.libevaluation_interop_symbols().kotlin.root
	resultCString := C.bridge_evaluate(root.evaluate, rulesCString, userCString)
	return C.GoString(resultCString)
}
