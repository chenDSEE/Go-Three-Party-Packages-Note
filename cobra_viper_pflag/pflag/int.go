package pflag

import "strconv"

// Q&A(DONE): 为什么要特意 type def 出 intValue 呢？
// 因为我们不能直接给 int 添加 method。
// 那为什么我们要个 int 添加 method 呢？
// 因为要让 int 这种类型的 flag 实现 pflag.Value interface
// 这样才能够在 FlagSet.VarP 对不同种类的 flag 进行统一的处理。
// 否则处理代码满天飞，当你支持多种不同基本类型的变量是，就会非常麻烦。
// 一旦出现了 bug，你就要修改很多代码
// -- int Value
type intValue int

// 将使用者为当前 flag 设置的默认值，直接设置到相应的变量中（set val into *p）
func newIntValue(val int, p *int) *intValue {
	*p = val
	return (*intValue)(p)
}

func (i *intValue) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	*i = intValue(v)
	return err
}

func (i *intValue) Type() string {
	return "int"
}

func (i *intValue) String() string { return strconv.Itoa(int(*i)) }

func intConv(sval string) (interface{}, error) {
	return strconv.Atoi(sval)
}

// GetInt return the int value of a flag with the given name
func (f *FlagSet) GetInt(name string) (int, error) {
	val, err := f.getFlagType(name, "int", intConv)
	if err != nil {
		return 0, err
	}
	return val.(int), nil
}

// IntVar defines an int flag with specified name, default value, and usage string.
// The argument p points to an int variable in which to store the value of the flag.
func (f *FlagSet) IntVar(p *int, name string, value int, usage string) {
	f.VarP(newIntValue(value, p), name, "", usage)
}

// IntVarP is like IntVar, but accepts a shorthand letter that can be used after a single dash.
func (f *FlagSet) IntVarP(p *int, name, shorthand string, value int, usage string) {
	f.VarP(newIntValue(value, p), name, shorthand, usage)
}

// IntVar defines an int flag with specified name, default value, and usage string.
// The argument p points to an int variable in which to store the value of the flag.
func IntVar(p *int, name string, value int, usage string) {
	CommandLine.VarP(newIntValue(value, p), name, "", usage)
}

// IntVarP is like IntVar, but accepts a shorthand letter that can be used after a single dash.
func IntVarP(p *int, name, shorthand string, value int, usage string) {
	CommandLine.VarP(newIntValue(value, p), name, shorthand, usage)
}

// Int defines an int flag with specified name, default value, and usage string.
// The return value is the address of an int variable that stores the value of the flag.
func (f *FlagSet) Int(name string, value int, usage string) *int {
	p := new(int)
	f.IntVarP(p, name, "", value, usage)
	return p
}

// IntP is like Int, but accepts a shorthand letter that can be used after a single dash.
func (f *FlagSet) IntP(name, shorthand string, value int, usage string) *int {
	p := new(int)
	f.IntVarP(p, name, shorthand, value, usage)
	return p
}

// public funtcion, 默认采用 default flag
// Int defines an int flag with specified name, default value, and usage string.
// The return value is the address of an int variable that stores the value of the flag.
func Int(name string, value int, usage string) *int {
	return CommandLine.IntP(name, "", value, usage)
}

// IntP is like Int, but accepts a shorthand letter that can be used after a single dash.
func IntP(name, shorthand string, value int, usage string) *int {
	return CommandLine.IntP(name, shorthand, value, usage)
}
