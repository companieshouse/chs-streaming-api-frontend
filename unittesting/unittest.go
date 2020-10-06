package unittesting

import (
	"runtime"
	"strings"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	// InternalStatus - slice to collect internal statuses/internal errors for functions without a return
	InternalStatus []string
)

// ResetInternalStatus - clears InternalStatus at Test start
func ResetInternalStatus() {
	InternalStatus = make([]string, 1)
}

// CollectInternalStatus - appends errors in the collecting array internalStatus
func CollectInternalStatus(err error) {
	if err != nil {
		InternalStatus = append(InternalStatus, err.Error())
	}
}

// CheckInternalStatus - util to check expected vs collected results
func CheckInternalStatus(expected []string) {
	sl := make([]string, 1)
	for _, r := range expected {
		if err := GetVarsError(r); err != nil {
			sl = append(sl, err.Error())
		}
	}
	s1 := strings.Join(InternalStatus, "/")
	s2 := strings.Join(sl, "/")
	So(s1, ShouldEqual, s2)
}

// GenericMap - Map to store any generic type
type GenericMap map[string]interface{}

// UnitTest - struct to define a Unit Test
type UnitTest struct {
	Title string // Test title
	Given string
	When  string
	Then  string
	Vars  GenericMap // List of any Variables required by the test
}

// TestsList is the map of one or more Unit Tests
type TestsList map[string]UnitTest

// Tests is the table for all the Unit Tests
type Tests struct {
	RunningTest string // key to use for look-up in the Tests list to retrieve the running one
	List        TestsList
}

// CurrentTests - points to the Current Table of Tests (must then be init via a call to InitTests)
var CurrentTests *Tests

// InitTests - must be called first by Unit Tests files to have a real Table of Tests to work on
func InitTests(t *Tests) {
	CurrentTests = t
}

// GetRunningTest - returns the Unit Test under execution
func (t Tests) GetRunningTest() UnitTest {
	return t.List[t.RunningTest]
}

// InitRunningTest - Init required resources before running a Test
func (t *Tests) InitRunningTest(testID string) UnitTest {
	t.RunningTest = testID
	InitTests(t)
	return t.List[testID]
}

// GetVarsVal - alias function to call GetMapVal for Vars
func (t Tests) GetVarsVal(key string) interface{} {
	unitTest := t.List[t.RunningTest]
	return unitTest.Vars[key]
}

// GetGiven - returns formatted Given
func (u *UnitTest) GetGiven() string {
	return "Given " + u.Given
}

// GetWhen - returns formatted When
func (u *UnitTest) GetWhen() string {
	when := u.When
	if len(when) == 0 {
		when = u.Title
	}
	return "When " + when
}

// GetThen - returns formatted Then
func (u *UnitTest) GetThen() string {
	return "Then " + u.Then
}

// GetVarsError - convenient method to call GetStubVal for 'error' types
func GetVarsError(key string) (err error) {
	i := CurrentTests.GetVarsVal(key)
	if i != nil {
		err = i.(error)
	}
	return
}

// GetVarsString - convenient method to call GetVarsVal for 'string' types
func GetVarsString(key string) (s string) {
	i := CurrentTests.GetVarsVal(key)
	if i != nil {
		s = i.(string)
	}
	return
}

// GetVarsInt - convenient method to call GetVarsVal for 'int' types
func GetVarsInt(key string) (v int) {
	i := CurrentTests.GetVarsVal(key)
	if i != nil {
		v = i.(int)
	}
	return
}

// GetVarsBool - convenient method to call GetVarsVal for 'bool' types
func GetVarsBool(key string) (b bool) {
	i := CurrentTests.GetVarsVal(key)
	if i != nil {
		b = i.(bool)
	}
	return
}

// GetCallerName - returns the name of the Caller to a function
func GetCallerName() string {
	fpcs := make([]uintptr, 1)
	runtime.Callers(3, fpcs)
	fun := runtime.FuncForPC(fpcs[0] - 1)
	s := strings.Split(fun.Name(), ".") // package_name.func_name
	return s[len(s)-1]                  // func_name
}
