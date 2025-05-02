package main

import (
	"log"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	kernel32      = windows.NewLazySystemDLL("kernel32.dll")
	rtlCopyMemory = kernel32.NewProc("RtlCopyMemory")
)

type FILE_DISPOSITION_INFO struct {
	DeleteFile uint32
}

type FILE_RENAME_INFO struct {
	Flags          uint32
	RootDirectory  windows.Handle
	FileNameLength uint32
	FileName       [1]uint16
}

func openHandle(path string) (windows.Handle, error) {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}

	handle, err := windows.CreateFile(
		p,
		windows.DELETE|windows.SYNCHRONIZE,
		windows.FILE_SHARE_READ,
		nil,
		windows.OPEN_EXISTING,
		0,
		0,
	)
	if err != nil {
		return 0, err
	}
	return handle, nil
}

func renameHandle(handle windows.Handle) error {
	DS_STREAM_RENAME := ":randomString"
	var fRename FILE_RENAME_INFO

	lpwStream, err := windows.UTF16PtrFromString(DS_STREAM_RENAME)
	if err != nil {
		return err
	}
	fRename.FileNameLength = uint32(unsafe.Sizeof(lpwStream))

	_, _, err = rtlCopyMemory.Call(
		uintptr(unsafe.Pointer(&fRename.FileName)),
		uintptr(unsafe.Pointer(lpwStream)),
		unsafe.Sizeof(lpwStream),
	)
	if err.Error() != "The operation completed successfully." {
		return err
	}

	err = windows.SetFileInformationByHandle(
		handle,
		windows.FileRenameInfo,
		(*byte)(unsafe.Pointer(&fRename)),
		uint32(unsafe.Sizeof(fRename)+unsafe.Sizeof(lpwStream)),
	)
	if err != nil {
		return err
	}

	var fDelete = FILE_DISPOSITION_INFO{}
	fDelete.DeleteFile = 1

	err = windows.SetFileInformationByHandle(
		handle,
		windows.FileDispositionInfo,
		(*byte)(unsafe.Pointer(&fDelete)),
		uint32(unsafe.Sizeof(fDelete)),
	)
	if err != nil {
		return err
	}

	windows.CloseHandle(handle)

	return nil
}

func depositeHandle(handle windows.Handle) error {
	var fDelete = FILE_DISPOSITION_INFO{}
	fDelete.DeleteFile = 1

	err := windows.SetFileInformationByHandle(
		handle,
		windows.FileDispositionInfo,
		(*byte)(unsafe.Pointer(&fDelete)),
		uint32(unsafe.Sizeof(fDelete)),
	)
	if err != nil {
		return err
	}

	windows.CloseHandle(handle)

	return nil
}

func main() {
	ex, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}

	handle, err := openHandle(ex)
	if err != nil {
		log.Fatal(err)
	}

	err = renameHandle(handle)
	if err != nil {
		log.Fatal(err)
	}

	handle, err = openHandle(ex)
	if err != nil {
		log.Fatal(err)
	}

	err = depositeHandle(handle)
	if err != nil {
		log.Fatal(err)
	}
}
