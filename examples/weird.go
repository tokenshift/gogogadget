package examples

// Test interface for different types of parameter and return lists.
//
// Unnamed and named arguments can't be mixed, but both cases need to be
// handled. For unnamed arguments, names of the form `val#` are generated for
// fields of the request/response structs (since struct fields MUST be named if
// there are duplicate types).
type Weird interface {
	UnnamedArgs(int, string, Thing) (Thing, int, string, bool, error)
	RepeatedTypes(a, b int, foo Thing) (x, y, z Thing, err error)
	MixedArgs(a int, b string, c, d Thing) (x, y, z int, thang Thing)
	UnnamedDuplicateTypes(int, int, string, Thing, string, Thing) (Thing, int, Thing, int, error)
}

type UnnamedArgsReq struct {
	val1 int
	val2 string
	val3 Thing
}

type UnnamedArgsRes struct {
	val1 Thing
	val2 int
	val3 string
	val4 bool
	val5 error
}

type RepeatedTypesReq struct {
	a, b int
	foo Thing
}

type RepeatedTypesRes struct {
	x, y, z Thing
	err error
}

type MixedArgsReq struct {
	a int
	b string
	c, d Thing
}

type MixedArgsRes struct {
	x, y, z int
	thang Thing
}

type UnnamedDuplicateTypesReq struct {
	val1 int
	val2 int
	val3 string
	val4 Thing
	val5 string
	val6 Thing
}

type UnnamedDuplicateTypesRes struct {
	val1 Thing
	val2 int
	val3 Thing
	val4 int
	val5 error
}

type Thing struct {
	A, B int
	C OtherThing
}

type OtherThing struct {
	Foo, Bar string
	Fizzbuzz *Thing
}
