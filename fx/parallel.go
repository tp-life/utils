package fx

import "github.com/tp-life/utils/threading"

// Parallel runs fns parallelly and waits for done.
func Parallel(fns ...func()) {
	group := threading.NewRoutineGroup()
	for _, fn := range fns {
		group.RunSafe(fn)
	}
	group.Wait()
}
