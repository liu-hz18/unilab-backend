#include <iostream>
#include <cmath>
#include <cstdio>
#include <cstdlib>
#include <unistd.h>
#include <asm/unistd.h>
#include <sys/ptrace.h>
#include <sys/wait.h>
#include <sys/resource.h>
#include <sys/user.h>
#include <fcntl.h>
#include <cstring>
#include <string>
#include <vector>
#include <set>
#include <argp.h>
#include "uoj_env.h"
using namespace std;

struct RunResult {
	int result;
	int ust;
	int usm;
	int exit_code;

	RunResult(int _result, int _ust = -1, int _usm = -1, int _exit_code = -1)
			: result(_result), ust(_ust), usm(_usm), exit_code(_exit_code) {
		if (result != RS_AC) {
			ust = -1, usm = -1;
		}
	}
};

struct RunProgramConfig
{
	int time_limit;
	int real_time_limit;
	int memory_limit;
	int output_limit;
	int stack_limit;
	string input_file_name;
	string output_file_name;
	string error_file_name;
	string result_file_name;
	string work_path;
	string type;
	vector<string> extra_readable_files;
	vector<string> extra_writable_files;
	bool allow_proc;
	bool safe_mode;
	bool need_show_trace_details;
	bool userss;

	string program_name;
	string program_basename;
	string program_dir;
	vector<string> argv;
};

int put_result(string result_file_name, RunResult res) {
	FILE *f;
	if (result_file_name == "stdout") {
		f = stdout;
	} else if (result_file_name == "stderr") {
		f = stderr;
	} else {
		f = fopen(result_file_name.c_str(), "w");
	}
	fprintf(f, "%d %d %d %d\n", res.result, res.ust, res.usm, res.exit_code);
	if (f != stdout && f != stderr) {
		fclose(f);
	}
	if (res.result == RS_JGF) {
		return 1;
	} else {
		return 0;
	}
}

char self_path[PATH_MAX + 1] = {};

#include "run_program_conf.h"

// name, key, arg, flags, doc, group
argp_option run_program_argp_options[] =
{	
	{"tl"                 , 'T', "TIME_LIMIT"  , 0, "Set time limit (in ms)"                                ,  1},
	{"rtl"                , 'R', "TIME_LIMIT"  , 0, "Set real time limit (in ms)"                           ,  2},
	{"ml"                 , 'M', "MEMORY_LIMIT", 0, "Set memory limit (in kb)"                              ,  3},
	{"ol"                 , 'O', "OUTPUT_LIMIT", 0, "Set output limit (in kb)"                              ,  4},
	{"sl"                 , 'S', "STACK_LIMIT" , 0, "Set stack limit (in kb)"                               ,  5},
	{"in"                 , 'i', "IN"          , 0, "Set input file name"                                   ,  6},
	{"out"                , 'o', "OUT"         , 0, "Set output file name"                                  ,  7},
	{"err"                , 'e', "ERR"         , 0, "Set error file name"                                   ,  8},
	{"work-path"          , 'w', "WORK_PATH"   , 0, "Set the work path of the program"                      ,  9},
	{"type"               , 't', "TYPE"        , 0, "Set the program type (for some program such as python)", 10},
	{"res"                , 'r', "RESULT_FILE" , 0, "Set the file name for outputing the result            ", 10},
	{"add-readable"       , 500, "FILE"        , 0, "Add a readable file"                                   , 11},
	{"add-writable"       , 505, "FILE"        , 0, "Add a writable file"                                   , 11},
	{"unsafe"             , 501, 0             , 0, "Don't check dangerous syscalls"                        , 12},
	{"show-trace-details" , 502, 0             , 0, "Show trace details"                                    , 13},
	{"allow-proc"         , 503, 0             , 0, "Allow fork, exec, vfork, nanosleep, clone... etc."     , 14},
	{"add-readable-raw"   , 504, "FILE"        , 0, "Add a readable (don't transform to its real path)"     , 15},
	{"add-writable-raw"   , 506, "FILE"        , 0, "Add a writable (don't transform to its real path)"     , 15},
	{"use-rss"            , 507, 0             , 0, "Use RSS as the memory value (use AS as default)"       , 16},
	{0}
};

error_t run_program_argp_parse_opt (int key, char *arg, struct argp_state *state)
{
	RunProgramConfig *config = (RunProgramConfig*)state->input;

	switch (key)
	{
		case 'T':
			config->time_limit = atoi(arg);
			break;
		case 'R':
			config->real_time_limit = atoi(arg);
			break;
		case 'M':
			config->memory_limit = atoi(arg);
			break;
		case 'O':
			config->output_limit = atoi(arg);
			break;
		case 'S':
			config->stack_limit = atoi(arg);
			break;
		case 'i':
			config->input_file_name = arg;
			break;
		case 'o':
			config->output_file_name = arg;
			break;
		case 'e':
			config->error_file_name = arg;
			break;
		case 'w':
			config->work_path = realpath(arg);
			if (config->work_path.empty()) {
				argp_usage(state);
			}
			break;
		case 'r':
			config->result_file_name = arg;
			break;
		case 't':
			config->type = arg;
			break;
		case 500:
			config->extra_readable_files.push_back(realpath(arg));
			break;
		case 501:
			config->safe_mode = false;
			break;
		case 502:
			config->need_show_trace_details = true;
			break;
		case 503:
			config->allow_proc = true;
			break;
		case 504:
			config->extra_readable_files.push_back(arg);
			break;
		case 505:
			config->extra_writable_files.push_back(realpath(arg));
			break;
		case 506:
			config->extra_writable_files.push_back(arg);
			break;
		case 507:
			config->userss = true;
			break;
		case ARGP_KEY_ARG:
			config->argv.push_back(arg);
			for (int i = state->next; i < state->argc; i++) {
				config->argv.push_back(state->argv[i]);
			}
			state->next = state->argc;
			break;
		case ARGP_KEY_END:
			if (state->arg_num == 0) {
				argp_usage(state);
			}
			break;
		default:
			return ARGP_ERR_UNKNOWN;
	}
	return 0;
}
char run_program_argp_args_doc[] = "program arg1 arg2 ...";
char run_program_argp_doc[] = "run_program: a tool to run program safely";

argp run_program_argp = {
	run_program_argp_options,
	run_program_argp_parse_opt,
	run_program_argp_args_doc,
	run_program_argp_doc
};

RunProgramConfig run_program_config;

void parse_args(int argc, char **argv) {
	run_program_config.time_limit = 1000;
	run_program_config.real_time_limit = -1;
	run_program_config.memory_limit = 256*1024;
	run_program_config.output_limit = 32*1024;
	run_program_config.stack_limit = 8*1024;
	run_program_config.input_file_name = "stdin";
	run_program_config.output_file_name = "stdout";
	run_program_config.error_file_name = "/dev/null";	// dump away error logs by default
	run_program_config.work_path = "";
	run_program_config.result_file_name = "stdout";
	run_program_config.type = "default";
	run_program_config.safe_mode = true;
	run_program_config.need_show_trace_details = false;
	run_program_config.allow_proc = false;
	run_program_config.userss = false;

	argp_parse(&run_program_argp, argc, argv, ARGP_NO_ARGS | ARGP_IN_ORDER, 0, &run_program_config);

	if (run_program_config.real_time_limit == -1)
		run_program_config.real_time_limit = run_program_config.time_limit + 2000;
	run_program_config.stack_limit = min(run_program_config.stack_limit, run_program_config.memory_limit);

	if (!run_program_config.work_path.empty()) {
		if (chdir(run_program_config.work_path.c_str()) == -1) {
			exit(put_result(run_program_config.result_file_name, RS_JGF));
		}
	}

	run_program_config.program_dir = dirname(run_program_config.argv[0]) + "/";
	if (run_program_config.type == "java8" || run_program_config.type == "java11" || run_program_config.type == "java14" || run_program_config.type == "java17") {
		run_program_config.program_name = run_program_config.argv[0];
		run_program_config.argv[0] = basename(run_program_config.argv[0]); // add `basename` for unilab
		run_program_config.userss = true; // add this for unilab
	} else if (run_program_config.type == "js" || run_program_config.type == "go") {
		run_program_config.userss = true;
		run_program_config.program_name = realpath(run_program_config.argv[0]);
	} else {
		run_program_config.program_name = realpath(run_program_config.argv[0]);
	}

	if (run_program_config.type == "compiler") {
		run_program_config.userss = true;
	}
	
	run_program_config.program_basename = basename(run_program_config.program_name);
	if (run_program_config.work_path.empty()) {
		run_program_config.work_path = dirname(run_program_config.program_name);
		run_program_config.argv[0] = "./" + run_program_config.program_basename;

		if (chdir(run_program_config.work_path.c_str()) == -1) {
			exit(put_result(run_program_config.result_file_name, RS_JGF));
		}
	}

	if (run_program_config.type == "python2.7") {
		string pre[4] = {"/usr/bin/python2.7", "-E", "-s", "-B"};
		run_program_config.argv.insert(run_program_config.argv.begin(), pre, pre + 4);
	} else if (run_program_config.type == "python3") {
		string pre[3] = {"/usr/bin/python3.10", "-I", "-B"};
		run_program_config.argv.insert(run_program_config.argv.begin(), pre, pre + 3);
	} else if (run_program_config.type == "java8") {
		// string pre[10] = {"/usr/lib/jvm/java-8-openjdk-amd64/bin/java", "-Xms32m", "-Xmx1024m", "-Xss1m", "-Djava.security.manager", "-Djava.security.policy=/home/cslab/unilab/unilab-backend/policy/java.policy", "-Djava.awt.headless=true", "-Dfile.encoding=UTF-8", "-classpath", run_program_config.program_dir};
		// run_program_config.argv.insert(run_program_config.argv.begin(), pre, pre + 10);
		string pre[8] = {"/usr/lib/jvm/java-8-openjdk-amd64/bin/java", "-Xms32m", "-Xmx1024m", "-Xss1m", "-Djava.awt.headless=true", "-Dfile.encoding=UTF-8", "-classpath", run_program_config.program_dir};
		run_program_config.argv.insert(run_program_config.argv.begin(), pre, pre + 8);
	} else if (run_program_config.type == "java11") {
		// string pre[10] = {"/usr/lib/jvm/java-11-openjdk-amd64/bin/java", "-Xms32m", "-Xmx1024m", "-Xss1m", "-Djava.security.manager", "-Djava.security.policy=/home/cslab/unilab/unilab-backend/policy/java.policy", "-Djava.awt.headless=true", "-Dfile.encoding=UTF-8", "-classpath", run_program_config.program_dir};
		// run_program_config.argv.insert(run_program_config.argv.begin(), pre, pre + 10);
		string pre[8] = {"/usr/lib/jvm/java-11-openjdk-amd64/bin/java", "-Xms32m", "-Xmx1024m", "-Xss1m", "-Djava.awt.headless=true", "-Dfile.encoding=UTF-8", "-classpath", run_program_config.program_dir};
		run_program_config.argv.insert(run_program_config.argv.begin(), pre, pre + 8);
	} else if (run_program_config.type == "java14") {
		string pre[8] = {"/usr/lib/jvm/java-14-openjdk-amd64/bin/java", "-Xms32m", "-Xmx1024m", "-Xss1m", "-Djava.awt.headless=true", "-Dfile.encoding=UTF-8", "-classpath", run_program_config.program_dir};
		run_program_config.argv.insert(run_program_config.argv.begin(), pre, pre + 8);
	} else if (run_program_config.type == "java17") {
		string pre[8] = {"/usr/lib/jvm/java-17-oracle-amd64/bin/java", "-Xms32m", "-Xmx1024m", "-Xss1m", "-Djava.awt.headless=true", "-Dfile.encoding=UTF-8", "-classpath", run_program_config.program_dir};
		run_program_config.argv.insert(run_program_config.argv.begin(), pre, pre + 8);
	} else if (run_program_config.type == "js") {
		string pre[1] = {"/usr/bin/node"};
		run_program_config.argv.insert(run_program_config.argv.begin(), pre, pre + 1);
	}
}

void set_limit(int r, int rcur, int rmax = -1)  {
	if (rmax == -1)
		rmax = rcur;
	// struct rlimit {
	// 	rlim_t    rlim_cur;    /* soft limit: current limit */
	// 	rlim_t    rlim_max;    /* hard limit: maximum value for rlim_cur */
	// };
	struct rlimit l;
	if (getrlimit(r, &l) == -1) {
		exit(55);
	}
	l.rlim_cur = rcur;
	l.rlim_max = rmax;
	if (setrlimit(r, &l) == -1) {
		exit(55);
	}
}

void run_child() {
	set_limit(RLIMIT_CPU,
		(int) ceil(run_program_config.time_limit / 1000.0),
		(int) ceil(run_program_config.real_time_limit / 1000.0));
	set_limit(RLIMIT_FSIZE, run_program_config.output_limit << 10);
	set_limit(RLIMIT_STACK, run_program_config.stack_limit << 10);

	if (run_program_config.input_file_name != "stdin") {
		if (freopen(run_program_config.input_file_name.c_str(), "r", stdin) == NULL) {
			exit(11);
		}
	}
	if (run_program_config.output_file_name != "stdout" && run_program_config.output_file_name != "stderr") {
		if (freopen(run_program_config.output_file_name.c_str(), "w", stdout) == NULL) {
			exit(12);
		}
	}

	if (run_program_config.error_file_name != "stderr") {
		if (run_program_config.error_file_name == "stdout") {
			if (dup2(1, 2) == -1) {
				exit(13);
			}
		} else {
			if (freopen(run_program_config.error_file_name.c_str(), "w", stderr) == NULL) {
				exit(14);
			}
		}
		
		if (run_program_config.output_file_name == "stderr") {
			if (dup2(2, 1) == -1) {
				exit(15);
			}
		}
	}

	char *env_path_str = getenv("PATH");
	char *env_lang_str = getenv("LANG");
	char *env_shell_str = getenv("SHELL");
	string env_path = env_path_str ? env_path_str : "";
	string env_lang = env_lang_str ? env_lang_str : "";
	string env_shell = env_shell_str ? env_shell_str : "";

	clearenv();
	setenv("USER", "poor_program", 1);
	setenv("LOGNAME", "poor_program", 1);
	setenv("HOME", run_program_config.work_path.c_str(), 1);
	if (env_lang_str) {
		setenv("LANG", env_lang.c_str(), 1);
	}
	if (env_path_str) {
		setenv("PATH", env_path.c_str(), 1);
	}
	setenv("PWD", run_program_config.work_path.c_str(), 1);
	if (env_shell_str) {
		setenv("SHELL", env_shell.c_str(), 1);
	}

	char **program_c_argv = new char*[run_program_config.argv.size() + 1];
	for (size_t i = 0; i < run_program_config.argv.size(); i++) {
		program_c_argv[i] = new char[run_program_config.argv[i].size() + 1];
		strcpy(program_c_argv[i], run_program_config.argv[i].c_str());
		// cout << program_c_argv[i] << endl;
	}
	program_c_argv[run_program_config.argv.size()] = NULL;

	if (ptrace(PTRACE_TRACEME, 0, NULL, NULL) == -1) {
		exit(16);
	}
	if (execv(program_c_argv[0], program_c_argv) == -1) {
		exit(17);
	}
}

const int MaxNRPChildren = 50;
struct rp_child_proc {
	pid_t pid;
	int mode;
};
int n_rp_children;
pid_t rp_timer_pid;
rp_child_proc rp_children[MaxNRPChildren];

int rp_children_pos(pid_t pid) {
	for (int i = 0; i < n_rp_children; i++) {
		if (rp_children[i].pid == pid) {
			return i;
		}
	}
	return -1;
}

int rp_children_add(pid_t pid) {
	if (n_rp_children == MaxNRPChildren) {
		return -1;
	}
	rp_children[n_rp_children].pid = pid;
	rp_children[n_rp_children].mode = -1;
	n_rp_children++;
	return 0;
}

void rp_children_del(pid_t pid) {
	int new_n = 0;
	for (int i = 0; i < n_rp_children; i++) {
		if (rp_children[i].pid != pid) {
			rp_children[new_n++] = rp_children[i];
		}
	}
	n_rp_children = new_n;
}

void stop_child(pid_t pid) {
	kill(pid, SIGKILL);
}

void stop_all() {
	kill(rp_timer_pid, SIGKILL);
	for (int i = 0; i < n_rp_children; i++) {
		kill(rp_children[i].pid, SIGKILL);
	}
}

RunResult trace_children() {
	rp_timer_pid = fork(); // pid < 0: error, pid == 0: this is child process, pid > 0: this is father process, son process is `pid`
	if (rp_timer_pid == -1) {
		stop_all();
		return RunResult(RS_JGF);
	} else if (rp_timer_pid == 0) {
		struct timespec ts;
		ts.tv_sec = (int) ceil(run_program_config.real_time_limit / 1000.0);
		ts.tv_nsec = 0;
		nanosleep(&ts, NULL);
		exit(0);
	}

	if (run_program_config.need_show_trace_details) {
		cerr << "timerpid " << rp_timer_pid << endl;
	}

	pid_t prev_pid = -1;
	int useras = 0;
	while (true) {
		int stat = 0;
		int sig = 0;
		struct rusage ruse;
		
		// pid_t wait4(pid_t pid, int *status, int options, struct rusage *rusage)
		// 挂起当前进程，等待指定的子进程状态改变
		// pid: 要关注的子进程的pid, -1 means wait for ANY child process whose process group ID is equal to the absolute value of pid
		// status: 子进程的返回状态. 
		//   WIFEXITED(status): is non-zero if the child exited normally
		//   WEXITSTATUS(status): can only be evaluated if WIFEXITED returned non-zero
		//   WIFSIGNALED(status): returns true if the child process exited because of a signal which was not caught
		//   WTERMSIG(status): returns the number of the signal that caused the child process to terminate.
		//   WIFSTOPPED(status): returns true if the child process which caused the return is currently stopped; this is only possible if the call was done using WUNTRACED.
		//   WSTOPSIG(status): returns the number of the signal which caused the child to stop. This macro can only be evaluated if WIFSTOPPED returned non-zero.
		// options: 进程等待选项
		//   WNOHANG: 立即返回
		//   WUNTRACED: 等子进程状态发生变化后才返回
		//   __WALL: Wait for all children, regardless of type
		//   
		// rusage: 死亡进程资源使用记录
		//  struct rusage {
		//  	struct timeval ru_utime; /* user time used */
		// 	    struct timeval ru_stime; /* system time used */
		// 		long   ru_maxrss;        /* maximum resident set size */
		// 		long   ru_ixrss;         /* integral shared memory size */
		// 		long   ru_idrss;         /* integral unshared data size */
		// 		long   ru_isrss;         /* integral unshared stack size */
		// 		long   ru_minflt;        /* page reclaims */
		// 		long   ru_majflt;        /* page faults */
		// 		long   ru_nswap;         /* swaps */
		// 		long   ru_inblock;       /* block input operations */
		// 		long   ru_oublock;       /* block output operations */
		// 		long   ru_msgsnd;        /* messages sent */
		// 		long   ru_msgrcv;        /* messages received */
		// 		long   ru_nsignals;      /* signals received */
		// 		long   ru_nvcsw;         /* voluntary context switches */
		// 		long   ru_nivcsw;        /* involuntary context switches */
		// };
		pid_t pid = wait4(-1, &stat, __WALL, &ruse);
		if (run_program_config.need_show_trace_details) {
			// if (prev_pid != pid) {
			// 	cerr << "----------" << pid << "----------" << endl;
			// }
			prev_pid = pid;
		}
		if (pid == rp_timer_pid) {
			if (WIFEXITED(stat) || WIFSIGNALED(stat)) {
				stop_all();
				fprintf(stderr, "WIFEXITED(stat) || WIFSIGNALED(stat)\n");
				return RunResult(RS_TLE);
			}
			continue;
		}
		
		int p = rp_children_pos(pid);
		if (p == -1) {
			if (run_program_config.need_show_trace_details) {
				fprintf(stderr, "new_proc  %lld\n", (long long int)pid);
			}
			if (rp_children_add(pid) == -1) {
				stop_child(pid);
				stop_all();
				return RunResult(RS_DGS);
			}
			p = n_rp_children - 1;
		}

		int usertim = ruse.ru_utime.tv_sec * 1000 + ruse.ru_utime.tv_usec / 1000 +
		              ruse.ru_stime.tv_sec * 1000 + ruse.ru_stime.tv_usec / 1000;
		int userrss = ruse.ru_maxrss;
		if (!run_program_config.userss)
		{
			char statm_path[255];
			sprintf(statm_path, "/proc/%d/statm", pid);
			FILE *statm = fopen(statm_path, "r");
			if (statm)
			{
				int size, resident, shared, trs, lrs, drs, dt;
				if (fscanf(statm, "%d%d%d%d%d%d%d", &size, &resident, &shared, &trs, &lrs, &drs, &dt) != 7)
				{
					cerr << "reading statm failed" << endl;
					stop_all();
					return RunResult(RS_JGF);
				}
				fclose(statm);
				useras = max(useras, drs * 4);
			}
		}
		int usermem = run_program_config.userss ? userrss : useras;
		if (usertim > run_program_config.time_limit) {
			stop_all();
			fprintf(stderr, "usertim > run_program_config.time_limit     : %d s\n", usertim);
			return RunResult(RS_TLE);
		}
		if (usermem > run_program_config.memory_limit) {
			stop_all();
			fprintf(stderr, "MLE! mem usage: %d. rss: %d, ras: %d", usermem, userrss, useras);
			return RunResult(RS_MLE);
		}

		if (WIFEXITED(stat)) { // 进程正常结束
			if (run_program_config.need_show_trace_details) {
				fprintf(stderr, "exit     : %d\n", WEXITSTATUS(stat));
			}
			if (rp_children[0].mode == -1) {
				stop_all();
				return RunResult(RS_JGF, -1, -1, WEXITSTATUS(stat));
			} else {
				if (pid == rp_children[0].pid) {
					stop_all();
					return RunResult(RS_AC, usertim, usermem, WEXITSTATUS(stat));
				} else {
					rp_children_del(pid);
					continue;
				}
			}
		}

		if (WIFSIGNALED(stat)) { // 进程异常终止
			if (run_program_config.need_show_trace_details) {
				fprintf(stderr, "sig exit : %d\n", WTERMSIG(stat));
			}
			if (pid == rp_children[0].pid) {
				switch(WTERMSIG(stat)) {
				case SIGXCPU: // nearly impossible
					stop_all();
					fprintf(stderr, "WTERMSIG(stat) == SIGXCPU\n");
					return RunResult(RS_TLE);
				case SIGXFSZ:
					stop_all();
					return RunResult(RS_OLE);
				default:
					stop_all();
					return RunResult(RS_RE);
				}
			} else {
				rp_children_del(pid);
				continue;
			}
		}
		
		if (WIFSTOPPED(stat)) { // 进程处于暂停状态
			sig = WSTOPSIG(stat); // 使得进程暂停的信号编号
			
			if (rp_children[p].mode == -1) {
				if ((p == 0 && sig == SIGTRAP) || (p != 0 && sig == SIGSTOP)) {
					if (p == 0) {
						int ptrace_opt = PTRACE_O_EXITKILL | PTRACE_O_TRACESYSGOOD;
						if (run_program_config.safe_mode) {
							ptrace_opt |= PTRACE_O_TRACECLONE | PTRACE_O_TRACEFORK | PTRACE_O_TRACEVFORK;
							ptrace_opt |= PTRACE_O_TRACEEXEC;
						}
						if (ptrace(PTRACE_SETOPTIONS, pid, NULL, ptrace_opt) == -1) {
							stop_all();
							return RunResult(RS_JGF);
						}
					}
					sig = 0;
				}
				rp_children[p].mode = 0;
			} else if (sig == (SIGTRAP | 0x80)) {
				if (rp_children[p].mode == 0) {
					if (run_program_config.safe_mode) {
						if (!check_safe_syscall(pid, run_program_config.need_show_trace_details)) {
							stop_all();
							return RunResult(RS_DGS);
						}
					}
					rp_children[p].mode = 1;
				} else {
					if (run_program_config.safe_mode) {
						on_syscall_exit(pid, run_program_config.need_show_trace_details);
					}
					rp_children[p].mode = 0;
				}
				
				sig = 0;
			} else if (sig == SIGTRAP) {
				switch ((stat >> 16) & 0xffff) {
					case PTRACE_EVENT_CLONE:
					case PTRACE_EVENT_FORK:
					case PTRACE_EVENT_VFORK:
						sig = 0;
						break;
					case PTRACE_EVENT_EXEC:
						rp_children[p].mode = 1;
						sig = 0;
						break;
					case 0:
						break;
					default:
						stop_all();
						return RunResult(RS_JGF);
				}
			}

			if (sig != 0) {
				if (run_program_config.need_show_trace_details) {
					fprintf(stderr, "sig      : %d\n", sig);
				}
			}
			
			switch(sig) {
			case SIGXCPU:
				stop_all();
				fprintf(stderr, "WSTOPSIG(stat) == SIGXCPU\n");
				return RunResult(RS_TLE);
			case SIGXFSZ:
				stop_all();
				return RunResult(RS_OLE);
			}
		}
		
		ptrace(PTRACE_SYSCALL, pid, NULL, sig);
	}
}

RunResult run_parent(pid_t pid) {
	init_conf(run_program_config);
	
	n_rp_children = 0;

	rp_children_add(pid);
	return trace_children();
}

int main(int argc, char **argv) {
	self_path[readlink("/proc/self/exe", self_path, PATH_MAX)] = '\0';
	parse_args(argc, argv);

	pid_t pid = fork();
	if (pid == -1) {
		return put_result(run_program_config.result_file_name, RS_JGF);
	} else if (pid == 0) {
		run_child();
	} else {
		return put_result(run_program_config.result_file_name, run_parent(pid));
	}
	return put_result(run_program_config.result_file_name, RS_JGF);
}
