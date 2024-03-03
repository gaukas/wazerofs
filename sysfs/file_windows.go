package sysfs

import (
	"errors"
	"syscall"
	"unsafe"

	"github.com/tetratelabs/wazero/experimental/sys"
)

const (
	nonBlockingFileReadSupported  = true
	nonBlockingFileWriteSupported = false

	_ERROR_IO_INCOMPLETE = syscall.Errno(996)
)

var kernel32 = syscall.NewLazyDLL("kernel32.dll")

// procPeekNamedPipe is the syscall.LazyProc in kernel32 for PeekNamedPipe
var (
	// procPeekNamedPipe is the syscall.LazyProc in kernel32 for PeekNamedPipe
	procPeekNamedPipe = kernel32.NewProc("PeekNamedPipe")
	// procGetOverlappedResult is the syscall.LazyProc in kernel32 for GetOverlappedResult
	procGetOverlappedResult = kernel32.NewProc("GetOverlappedResult")
	// procCreateEventW is the syscall.LazyProc in kernel32 for CreateEventW
	procCreateEventW = kernel32.NewProc("CreateEventW")
)

// readFd returns ENOSYS on unsupported platforms.
//
// PeekNamedPipe: https://learn.microsoft.com/en-us/windows/win32/api/namedpipeapi/nf-namedpipeapi-peeknamedpipe
// "GetFileType can assist in determining what device type the handle refers to. A console handle presents as FILE_TYPE_CHAR."
// https://learn.microsoft.com/en-us/windows/console/console-handles
func readFd(fd uintptr, buf []byte) (int, sys.Errno) {
	handle := syscall.Handle(fd)
	fileType, err := syscall.GetFileType(handle)
	if err != nil {
		return 0, sys.UnwrapOSError(err)
	}
	if fileType&syscall.FILE_TYPE_CHAR == 0 {
		return -1, sys.ENOSYS
	}
	n, errno := peekNamedPipe(handle)
	if errno == syscall.ERROR_BROKEN_PIPE {
		return 0, 0
	}
	if n == 0 {
		return -1, sys.EAGAIN
	}
	un, err := syscall.Read(handle, buf[0:n])
	return un, sys.UnwrapOSError(err)
}

func writeFd(fd uintptr, buf []byte) (int, sys.Errno) {
	return -1, sys.ENOSYS
}

func writeSocket(fd uintptr, buf []byte) (int, sys.Errno) {
	var done uint32
	var overlapped syscall.Overlapped
	errno := syscall.WriteFile(syscall.Handle(fd), buf, &done, &overlapped)
	if errors.Is(errno, syscall.ERROR_IO_PENDING) {
		errno = syscall.EAGAIN
	}
	return int(done), sys.UnwrapOSError(errno)
}

// peekNamedPipe partially exposes PeekNamedPipe from the Win32 API
// see https://learn.microsoft.com/en-us/windows/win32/api/namedpipeapi/nf-namedpipeapi-peeknamedpipe
func peekNamedPipe(handle syscall.Handle) (uint32, syscall.Errno) {
	var totalBytesAvail uint32
	totalBytesPtr := unsafe.Pointer(&totalBytesAvail)
	_, _, errno := syscall.SyscallN(
		procPeekNamedPipe.Addr(),
		uintptr(handle),        // [in]            HANDLE  hNamedPipe,
		0,                      // [out, optional] LPVOID  lpBuffer,
		0,                      // [in]            DWORD   nBufferSize,
		0,                      // [out, optional] LPDWORD lpBytesRead
		uintptr(totalBytesPtr), // [out, optional] LPDWORD lpTotalBytesAvail,
		0)                      // [out, optional] LPDWORD lpBytesLeftThisMessage
	return totalBytesAvail, errno
}

func rmdir(path string) sys.Errno {
	err := syscall.Rmdir(path)
	return sys.UnwrapOSError(err)
}

func getOverlappedResult(handle syscall.Handle, overlapped *syscall.Overlapped, wait bool) (uint32, syscall.Errno) {
	var totalBytesAvail uint32
	var bwait uintptr
	if wait {
		bwait = 0xFFFFFFFF
	}
	totalBytesPtr := unsafe.Pointer(&totalBytesAvail)
	_, _, errno := syscall.SyscallN(
		procGetOverlappedResult.Addr(),
		uintptr(handle),                     // [in]  HANDLE       hFile,
		uintptr(unsafe.Pointer(overlapped)), // [in]  LPOVERLAPPED lpOverlapped,
		uintptr(totalBytesPtr),              // [out] LPDWORD      lpNumberOfBytesTransferred,
		bwait)                               // [in]  BOOL         bWait
	return totalBytesAvail, errno
}

func createEventW(lpEventAttributes *syscall.SecurityAttributes, bManualReset bool, bInitialState bool, lpName *uint16) (uintptr, syscall.Errno) {
	var manualReset uintptr
	var initialState uintptr
	if bManualReset {
		manualReset = 1
	}
	if bInitialState {
		initialState = 1
	}
	handle, _, errno := syscall.SyscallN(
		procCreateEventW.Addr(),
		uintptr(unsafe.Pointer(lpEventAttributes)), // [in]     LPSECURITY_ATTRIBUTES lpEventAttributes,
		manualReset,                     // [in]     BOOL                  bManualReset,
		initialState,                    // [in]     BOOL                  bInitialState,
		uintptr(unsafe.Pointer(lpName)), // [in, opt]LPCWSTR               lpName,
	)

	return handle, errno
}
