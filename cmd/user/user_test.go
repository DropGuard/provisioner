package main

import (
	"os"
	"syscall"
	"testing"
	"unsafe"
)

func TestIntegrationUserCreation(t *testing.T) {
	// Skip the test on developer machines to prevent intrusive local changes
	if os.Getenv("CI") != "true" {
		t.Skip("Skipping integration test in non-CI environment")
	}

	if !isAdmin() {
		t.Fatalf("Administrator privileges are required to run integration tests, but current process is not running as Administrator.")
	}

	username := "DailyUserCI"
	t.Logf("Running integration test for user creation: %s", username)

	netapi32 := syscall.NewLazyDLL("netapi32.dll")
	netUserAdd := netapi32.NewProc("NetUserAdd")
	netUserDel := netapi32.NewProc("NetUserDel")
	netLocalGroupAddMembers := netapi32.NewProc("NetLocalGroupAddMembers")

	userPtr, err := syscall.UTF16PtrFromString(username)
	if err != nil {
		t.Fatalf("failed to process username: %v", err)
	}

	// Cleanup leftover user from any aborted previous test run
	_, _, _ = netUserDel.Call(0, uintptr(unsafe.Pointer(userPtr)))

	// Clean up user at the end of the test
	t.Cleanup(func() {
		ret, _, _ := netUserDel.Call(0, uintptr(unsafe.Pointer(userPtr)))
		// 2221 is NERR_UserNotFound; we ignore that error code since it's already deleted
		if ret != 0 && ret != 2221 {
			t.Logf("cleanup: NetUserDel failed with error code %d: %v", ret, syscall.Errno(ret))
		} else {
			t.Logf("cleanup: successfully deleted test user %s", username)
		}
	})

	passPtr, _ := syscall.UTF16PtrFromString("") // No password
	ui := USER_INFO_1{
		Name:     userPtr,
		Password: passPtr,
		Priv:     1,               // USER_PRIV_USER
		Flags:    0x0040 | 0x0200, // UF_NORMAL_ACCOUNT | UF_SCRIPT
	}

	// Call NetUserAdd(servername=nil, level=1, buf, parm_err=nil)
	ret, _, _ := netUserAdd.Call(
		0,
		1,
		uintptr(unsafe.Pointer(&ui)),
		0,
	)

	if ret != 0 {
		t.Fatalf("NetUserAdd failed (error code %d): %v", ret, syscall.Errno(ret))
	}
	t.Logf("Successfully created test user %s", username)

	// Resolve the localized group name for Administrators (SID S-1-5-32-544)
	sidStr := "S-1-5-32-544"
	sid, err := syscall.StringToSid(sidStr)
	if err != nil {
		t.Fatalf("failed to convert SID: %v", err)
	}

	var nameLen, domainLen uint32 = 0, 0
	var sidUse uint32
	// Call once to get required buffer sizes
	err = syscall.LookupAccountSid(nil, sid, nil, &nameLen, nil, &domainLen, &sidUse)
	if err != nil && err != syscall.ERROR_INSUFFICIENT_BUFFER {
		t.Fatalf("failed to lookup group account sizes: %v", err)
	}

	nameBuf := make([]uint16, nameLen)
	domainBuf := make([]uint16, domainLen)
	err = syscall.LookupAccountSid(nil, sid, &nameBuf[0], &nameLen, &domainBuf[0], &domainLen, &sidUse)
	if err != nil {
		t.Fatalf("failed to lookup group account details: %v", err)
	}

	groupNameStr := syscall.UTF16ToString(nameBuf)
	t.Logf("Resolved localized Administrators group name: %s", groupNameStr)

	groupNamePtr, err := syscall.UTF16PtrFromString(groupNameStr)
	if err != nil {
		t.Fatalf("failed to process group name: %v", err)
	}

	member := LOCALGROUP_MEMBERS_INFO_3{
		DomainAndName: userPtr,
	}

	// Call NetLocalGroupAddMembers(servername=nil, groupname, level=3, buf, totalentries=1)
	ret, _, _ = netLocalGroupAddMembers.Call(
		0,
		uintptr(unsafe.Pointer(groupNamePtr)),
		3,
		uintptr(unsafe.Pointer(&member)),
		1,
	)

	if ret != 0 {
		t.Fatalf("NetLocalGroupAddMembers failed (error code %d): %v", ret, syscall.Errno(ret))
	}
	t.Logf("Successfully added test user %s to the %s group", username, groupNameStr)
}
