package eval

import (
	"reflect"
	"strconv"
	"syscall"
	"testing"

	"github.com/elves/elvish/daemon/api"
)

func TestBuiltinPid(t *testing.T) {
	pid := strconv.Itoa(syscall.Getpid())
	builtinPid := ToString(makeBuiltinNamespace(nil)["pid"].Get())
	if builtinPid != pid {
		t.Errorf(`ev.builtin["pid"] = %v, want %v`, builtinPid, pid)
	}
}

// To be called from init in separate test files.
func addToEvalTests(tests []Test) {
	evalTests = append(evalTests, tests...)
}

var evalTests = []Test{
	// Pseudo-namespaces local: and up:
	{"x=lorem; []{local:x=ipsum; put $up:x $local:x}",
		want{out: strs("lorem", "ipsum")}},
	{"x=lorem; []{up:x=ipsum; put $x}; put $x",
		want{out: strs("ipsum", "ipsum")}},
	// Pseudo-namespace E:
	{"E:FOO=lorem; put $E:FOO", want{out: strs("lorem")}},
	{"del E:FOO; put $E:FOO", want{out: strs("")}},
}

func TestEval(t *testing.T) {
	testEval(t, dataDir, evalTests)
}

func TestMultipleEval(t *testing.T) {
	texts := []string{"x=hello", "put $x"}
	outs, _, err := evalAndCollect(t, dataDir, texts, 1)
	wanted := strs("hello")
	if err != nil {
		t.Errorf("eval %s => %v, want nil", texts, err)
	}
	if !reflect.DeepEqual(outs, wanted) {
		t.Errorf("eval %s outputs %v, want %v", texts, outs, wanted)
	}
}

func BenchmarkOutputCaptureOverhead(b *testing.B) {
	op := Op{func(*EvalCtx) {}, 0, 0}
	benchmarkOutputCapture(op, b.N)
}

func BenchmarkOutputCaptureValues(b *testing.B) {
	op := Op{func(ec *EvalCtx) {
		ec.ports[1].Chan <- String("test")
	}, 0, 0}
	benchmarkOutputCapture(op, b.N)
}

func BenchmarkOutputCaptureBytes(b *testing.B) {
	bytesToWrite := []byte("test")
	op := Op{func(ec *EvalCtx) {
		ec.ports[1].File.Write(bytesToWrite)
	}, 0, 0}
	benchmarkOutputCapture(op, b.N)
}

func BenchmarkOutputCaptureMixed(b *testing.B) {
	bytesToWrite := []byte("test")
	op := Op{func(ec *EvalCtx) {
		ec.ports[1].Chan <- Bool(false)
		ec.ports[1].File.Write(bytesToWrite)
	}, 0, 0}
	benchmarkOutputCapture(op, b.N)
}

func benchmarkOutputCapture(op Op, n int) {
	ev := NewEvaler(api.NewClient("/invalid"), nil, "", nil)
	ec := NewTopEvalCtx(ev, "[benchmark]", "", []*Port{{}, {}, {}})
	for i := 0; i < n; i++ {
		pcaptureOutput(ec, op)
	}
}
