package consts

// client module and command
const (
	ModuleUpdate           = "update"
	ModuleExecution        = "exec"
	ModuleExecuteAssembly  = "execute_assembly"
	ModuleInlineAssembly   = "inline_assembly"
	ModuleExecuteShellcode = "execute_shellcode"
	ModuleInlineShellcode  = "inline_shellcode"
	ModuleExecutePE        = "execute_pe"
	ModuleInlinePE         = "inline_pe"
	ModuleExecuteDll       = "execute_dll"
	ModuleInlineDll        = "inline_dll"
	ModuleExecuteBof       = "bof"
	ModuleUpload           = "upload"
	ModuleDownload         = "download"
	ModulePwd              = "pwd"
	ModuleLs               = "ls"
	ModuleCd               = "cd"
	ModuleMv               = "mv"
	ModuleMkdir            = "mkdir"
	ModuleRm               = "rm"
	ModuleCat              = "cat"
	ModulePs               = "ps"
	ModuleCp               = "cp"
	ModuleChmod            = "chmod"
	ModuleChown            = "chown"
	ModuleKill             = "kill"
	ModuleWhoami           = "whoami"
	ModuleEnv              = "env"
	ModuleSetEnv           = "setenv"
	ModuleUnsetEnv         = "unsetenv"
	ModuleNetstat          = "netstat"
	ModuleCurl             = "curl"
	ModuleListModule       = "list-module"
	ModuleLoadModule       = "load-module"
	CommandSync            = "sync"
	CommandBroadcast       = "broadcast"
	CommandVersion         = "version"
	CommandNotify          = "notify"
	CommandAlias           = "alias"
	CommandAliasLoad       = "load"
	CommandAliasInstall    = "install"
	CommandAliasRemove     = "remove"
)

// ctrl type
const (
	CtrlPipelineStart = 0 + iota
	CtrlPipelineStop
)

// ctrl status
const (
	CtrlStatusSuccess = 0 + iota
	CtrlStatusFailed
)

// task error
const (
	TaskErrorOperatorError       = 2
	TaskErrorNotExpectBody       = 3
	TaskErrorFieldRequired       = 4
	TaskErrorFieldLengthMismatch = 5
	TaskErrorFieldInvalid        = 6
	TaskError                    = 99
)
