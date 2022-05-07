package judger

const (
	RunFinished         uint32 = 0
	UnknownError        uint32 = 1
	RuntimeError        uint32 = 2
	MemoryLimitExceeded uint32 = 3
	TimeLimitExceeded   uint32 = 4
	OutputLimitExceeded uint32 = 5
	DangerousSystemCall uint32 = 6
	JudgeFailed         uint32 = 7
	CompileError        uint32 = 8
	WrongAnswer         uint32 = 9
)

var RunResultMap = map[uint32]string{
	RunFinished:         "Accepted",
	UnknownError:        "UnknownError",
	RuntimeError:        "RuntimeError",
	MemoryLimitExceeded: "MemoryLimitExceeded",
	TimeLimitExceeded:   "TimeLimitExceeded",
	OutputLimitExceeded: "OutputLimitExceeded",
	DangerousSystemCall: "DangerousSystemCall",
	JudgeFailed:         "JudgerFailed",
	CompileError:        "CompileError",
	WrongAnswer:         "WrongAnswer",
}

const DefaultResourceLimiter = "ulimit -t 5 && ulimit -v 524288 && ulimit -f 65536"

// compiler
const GNUCompilerResourceLimiter = "ulimit -t 10 && ulimit -v 524288 && ulimit -f 65536"
const PythonCompilerResourceLimiter = "ulimit -t 10 && ulimit -v 524288 && ulimit -f 65536"
const JavaCompilerResourceLimiter = "ulimit -t 10 && ulimit -v 6291456 && ulimit -f 65536"
const GoCompilerResourceLimiter = "ulimit -t 10 && ulimit -v 1048576 && ulimit -f 65536"
const NodeResourceLimiter = "ulimit -t 10 && ulimit -v 2097152 && ulimit -f 65536"

// checker
const CheckerResourceLimiter = "ulimit -t 5 && ulimit -v 262144 && ulimit -f 65536"

// runtime
const CRuntimeResourceLimiter = "ulimit -t 5 && ulimit -v 524288 && ulimit -f 65536"
const PythonRuntimeResourceLimiter = "ulimit -t 5 && ulimit -v 524288 && ulimit -f 65536"
const JavaRuntimeResourceLimiter = "ulimit -t 10 && ulimit -v 18874368 && ulimit -f 65536"
const GoRuntimeResourceLimiter = "ulimit -t 10 && ulimit -v 1048576 && ulimit -f 65536"

type LanguageConfig struct {
	Compile       string
	RunType       string
	Executable    string
	CompileLimits string
	RuntimeLimits string
	Environments  string
	NeedFile      string
	Timeout       int
}

// TODO: add makefile and cmake support
// judger configs
var JudgerConfig = map[string]LanguageConfig{
	"c": {
		Compile:       "gcc main.c -DONLINE_JUDGE -lm -Wall -O2 -fmax-errors=5 -fdiagnostics-color=never -o main.exe",
		RunType:       "default",
		Executable:    "main.exe",
		CompileLimits: GNUCompilerResourceLimiter,
		RuntimeLimits: CRuntimeResourceLimiter,
		Environments:  "",
		NeedFile:      "main.c",
		Timeout:       5,
	},
	"c++11": {
		Compile:       "g++ main.cpp -DONLINE_JUDGE -lm -Wall -O2 -fmax-errors=5 -std=c++11 -o main.exe -fdiagnostics-color=never",
		RunType:       "default",
		Executable:    "main.exe",
		CompileLimits: GNUCompilerResourceLimiter,
		RuntimeLimits: CRuntimeResourceLimiter,
		Environments:  "",
		NeedFile:      "main.cpp",
		Timeout:       5,
	},
	"c++14": {
		Compile:       "g++ main.cpp -DONLINE_JUDGE -lm -Wall -O2 -fmax-errors=5 -std=c++14 -o main.exe -fdiagnostics-color=never",
		RunType:       "default",
		Executable:    "main.exe",
		CompileLimits: GNUCompilerResourceLimiter,
		RuntimeLimits: CRuntimeResourceLimiter,
		Environments:  "",
		NeedFile:      "main.cpp",
		Timeout:       5,
	},
	"c++17": {
		Compile:       "g++ main.cpp -DONLINE_JUDGE -lm -Wall -O2 -fmax-errors=5 -std=c++17 -o main.exe -fdiagnostics-color=never",
		RunType:       "default",
		Executable:    "main.exe",
		CompileLimits: GNUCompilerResourceLimiter,
		RuntimeLimits: CRuntimeResourceLimiter,
		Environments:  "",
		NeedFile:      "main.cpp",
		Timeout:       5,
	},
	"c++20": {
		Compile:       "g++ main.cpp -DONLINE_JUDGE -lm -Wall -O2 -fmax-errors=5 -std=c++20 -o main.exe -fdiagnostics-color=never",
		RunType:       "default",
		Executable:    "main.exe",
		CompileLimits: GNUCompilerResourceLimiter,
		RuntimeLimits: CRuntimeResourceLimiter,
		Environments:  "",
		NeedFile:      "main.cpp",
		Timeout:       5,
	},
	"java8": {
		Compile:       "/usr/lib/jvm/java-8-openjdk-amd64/bin/javac -encoding UTF-8 Main.java",
		RunType:       "java8",
		Executable:    "Main",
		CompileLimits: JavaRuntimeResourceLimiter,
		RuntimeLimits: JavaCompilerResourceLimiter,
		Environments:  "",
		NeedFile:      "Main.java",
		Timeout:       10,
	},
	"java11": {
		Compile:       "/usr/lib/jvm/java-11-openjdk-amd64/bin/javac -encoding UTF-8 Main.java",
		RunType:       "java11",
		Executable:    "Main",
		CompileLimits: JavaCompilerResourceLimiter,
		RuntimeLimits: JavaRuntimeResourceLimiter,
		Environments:  "",
		NeedFile:      "Main.java",
		Timeout:       10,
	},
	"java14": {
		Compile:       "/usr/lib/jvm/java-14-openjdk-amd64/bin/javac -encoding UTF-8 Main.java",
		RunType:       "java14",
		Executable:    "Main",
		CompileLimits: JavaCompilerResourceLimiter,
		RuntimeLimits: JavaRuntimeResourceLimiter,
		Environments:  "",
		NeedFile:      "Main.java",
		Timeout:       10,
	},
	"java17": {
		Compile:       "/usr/lib/jvm/java-17-oracle-amd64/bin/javac -encoding UTF-8 Main.java",
		RunType:       "java17",
		Executable:    "Main",
		CompileLimits: JavaCompilerResourceLimiter,
		RuntimeLimits: JavaRuntimeResourceLimiter,
		Environments:  "",
		NeedFile:      "Main.java",
		Timeout:       10,
	},
	"python2": {
		Compile:       "/usr/bin/python2.7 -E -s -B -O -c \"import py_compile\nimport sys\ntry:\n    py_compile.compile('main.py', doraise=True)\n    sys.exit(0)\nexcept Exception as e:\n    print e\n    sys.exit(1)\n\"",
		RunType:       "python2.7",
		Executable:    "main.pyo",
		CompileLimits: PythonCompilerResourceLimiter,
		RuntimeLimits: PythonRuntimeResourceLimiter,
		Environments:  "",
		NeedFile:      "main.py",
		Timeout:       10,
	},
	"python3": {
		Compile:       "/usr/bin/python3.10 -I -B -O -c \"import py_compile\nimport sys\ntry:\n    py_compile.compile('main.py', doraise=True)\n    sys.exit(0)\nexcept Exception as e:\n    print(e)\n    sys.exit(1)\n\"",
		RunType:       "python3",
		Executable:    "__pycache__/main.cpython-310.opt-1.pyc",
		CompileLimits: PythonCompilerResourceLimiter,
		RuntimeLimits: PythonRuntimeResourceLimiter,
		Environments:  "",
		NeedFile:      "main.py",
		Timeout:       10,
	},
	"rust": {
		Compile:       "~/.cargo/bin/rustc main.rs -O --color=never -o main.exe",
		RunType:       "rust",
		Executable:    "main.exe",
		CompileLimits: GNUCompilerResourceLimiter,
		RuntimeLimits: CRuntimeResourceLimiter,
		Environments:  "",
		NeedFile:      "main.rs",
		Timeout:       5,
	},
	"js": {
		Compile:       "/usr/bin/node --check main.js",
		RunType:       "js",
		Executable:    "main.js",
		CompileLimits: NodeResourceLimiter,
		RuntimeLimits: NodeResourceLimiter,
		Environments:  "",
		NeedFile:      "main.js",
		Timeout:       10,
	},
	"go": {
		Compile:       "/usr/local/go/bin/go build -ldflags=\"-s -w\" -p 1 -o main.exe main.go",
		RunType:       "go",
		Executable:    "main.exe",
		CompileLimits: GoCompilerResourceLimiter,
		RuntimeLimits: GoRuntimeResourceLimiter,
		Environments:  "",
		NeedFile:      "main.go",
		Timeout:       8,
	},
}

var ExtLint = map[string]string{
	".c":    "text/x-csrc",
	".cpp":  "text/x-c++src",
	".py":   "text/x-python",
	".java": "text/x-java",
	".js":   "text/javascript",
	".go":   "text/x-go",
	".rs":   "text/x-rustsrc",
}
